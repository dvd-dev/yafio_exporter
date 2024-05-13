package main

// Prometheus exporter for fio benchmarks

// By default a chosen benchmark job will run periodically with the results
// being exported in Prometheus format

// inspired by
// https://github.com/neoaggelos/fio-exporter
// https://github.com/fritchie/fio_benchmark_exporter
// https://github.com/mwennrich/fio_benchmark_exporter

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var savedLabels = []string{"bs", "jobname", "iodepth", "rw", "size"}
var cmdCleanup = regexp.MustCompile(`--(output-format|group_reporting|name|[c|s]*lat_percentiles|runtime|status-interval|time_based)[=]*[^\s]*`)
var alwaysOn = "--output-format=json --lat_percentiles=1 --clat_percentiles=1 --time_based=1 --group_reporting"

// A CmdIO defines a cmd that will feed the i3bar.
type CmdIO struct {
	// Cmd is the command being run
	Cmd *exec.Cmd
	// reader is the underlying stream where Cmd outputs data.
	reader io.ReadCloser
	// writer is the underlying stream where Cmd outputs data.
	writer io.WriteCloser
}

// NewCmdIO creates a new CmdIO from command c.
// c must be properly quoted for a shell as it's passed to sh -c.
func NewCmdIO(c string) (*CmdIO, error) {
	cmd := exec.Command(os.Getenv("SHELL"), "-c", c)
	reader, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	writer, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	cmdio := CmdIO{
		Cmd:    cmd,
		reader: reader,
		writer: writer,
	}
	return &cmdio, nil
}
func (c *CmdIO) Close() error {
	if err := c.Cmd.Process.Signal(syscall.SIGTERM); err != nil {
		log.Println(err)
		if err := c.Cmd.Process.Kill(); err != nil {
			return err
		}
	}
	if err := c.Cmd.Process.Release(); err != nil {
		return err
	}
	if err := c.reader.Close(); err != nil {
		return err
	}
	if err := c.writer.Close(); err != nil {
		return err
	}
	return nil
}

func main() {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	var wg sync.WaitGroup
	fioFlags := flag.String("flags", "", "flags passed to fio")
	name := flag.String("name", "benchmark", "Job name")
	testid := flag.String("testid", "", "Test ID label to use with K6")
	preset := flag.String("preset", "", "Preset name (iops, latency or throughput)")
	port := flag.String("port", "9996", "metrics tcp listen port")
	noExporter := flag.Bool("noExporter", false, "Disables the prometheus endpoint")
	statusUpdateInterval := flag.String("statusUpdateInterval", "5", "metric update interval in seconds when statusUpdates enabled")
	runtime := flag.String("runtime", "60", "Runtime")
	flag.Parse()
	var fioArgs strings.Builder
	var presetFlags string
	if len(*testid) > 0 {
		savedLabels = append(savedLabels, "testid")
	}
	switch *preset {
	case "iops":
		presetFlags = "--numjobs=4 --ioengine=libaio --direct=1 --size=1G --bs=4k --iodepth=128 --readwrite=randrw"
	case "latency":
		presetFlags = "--numjobs=1 --ioengine=libaio --direct=1 --size=1G --bs=4k --iodepth=1 --readwrite=randrw"
	case "throughput":
		presetFlags = "--numjobs=4 --ioengine=libaio --direct=1 --size=1G --bs=128k --iodepth=64 --readwrite=rw"
	}
	if len(presetFlags) > 0 {
		fioArgs.WriteString(presetFlags)
	}
	fioArgs.WriteString(*fioFlags)
	if !strings.Contains(*fioFlags, "--numjobs") {
		fioArgs.WriteString(" --numjobs=1")
	}
	if !strings.Contains(*fioFlags, "--ioengine") {
		fioArgs.WriteString(" --ioengine=libaio")
	}
	if !strings.Contains(*fioFlags, "--direct") {
		fioArgs.WriteString(" --direct=1")
	}
	if !strings.Contains(*fioFlags, "--bs") {
		fioArgs.WriteString(" --bs=4k")
	}
	if !strings.Contains(*fioFlags, "--readwrite") {
		fioArgs.WriteString(" --readwrite=randrw")
	}
	if !strings.Contains(*fioFlags, "--iodepth") {
		fioArgs.WriteString(" --iodepth=128")
	}
	if !strings.Contains(*fioFlags, "--size") {
		fioArgs.WriteString(" --size=64m")
	}
	cmd := fmt.Sprintf(
		"fio --name=%s --runtime=%s --status-interval=%s %s %s",
		*name,
		*runtime,
		*statusUpdateInterval,
		alwaysOn,
		cmdCleanup.ReplaceAllString(fioArgs.String(), ""),
	)
	fio, err := NewCmdIO(cmd)
	if err != nil {
		log.Fatalf("NewCmd fails: %s", err)
		os.Exit(1)
	}
	wg.Add(1)
	go func() {
		ch := make(chan struct{}, 1)
		ch <- struct{}{}
		log.Printf("Running fio with command: %s", cmd)
		if err := fio.Cmd.Start(); err != nil {
			log.Fatalf("Failed to start fio: %s", err)
			os.Exit(1)
		}
		r := bufio.NewReader(fio.reader)
		dec := json.NewDecoder(r)
		for {
			var j FioResult
			if err := dec.Decode(&j); err == io.EOF {
				break
			} else if err != nil {
				log.Fatalf("Failed to decode JSON: %s", err)
			}
			//fmt.Print(j)
			fmt.Print(j.Print())
			Build(&j, *testid)
		}
		if err := fio.Cmd.Wait(); err != nil {
			log.Fatalf("Unable to wait: %s", err)
			return
		}
		wg.Done()
	}()
	go func() {
		for {
			s := <-sigc
			switch s {
			case syscall.SIGTERM:
				fallthrough
			case os.Interrupt:
				// Kill all processes on interrupt
				log.Println("SIGINT or SIGTERM received: terminating all processes...")
				if err := fio.Close(); err != nil {
					log.Println(err)
				}
				os.Exit(0)
			}
		}
	}()
	if !*noExporter {
		http.Handle("/metrics", promhttp.HandlerFor(

			promRegistry,
			promhttp.HandlerOpts{},
		))

		log.Printf("Listening on :%s\n", *port)

		server := &http.Server{
			Addr:              ":" + *port,
			ReadHeaderTimeout: 30 * time.Second,
		}
		log.Fatal(server.ListenAndServe())
	} else {
		log.Print("Waiting for threads")
		wg.Wait()
	}

}
