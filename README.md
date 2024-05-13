# Yet Another Fio Exporter

Prometheus exporter for [fio](https://github.com/axboe/fio) benchmarks. Fio (Flexible I/O Tester) is a tool for storage performance benchmarking.

This is a fork from [this version of fio_benchmark_exporter](https://github.com/mwennrich/fio_benchmark_exporter) and the goal here is to parse `json` output instead of the `terse` output format so that it's easy to export all the metrics.

It's my first go-lang project so it's probably unoptimal and ugly code. Please submit PRs to improve the code.

## Building and running

### Build

```
go build .
```

A sample Dockerfile, docker-compose.yaml, kustomization.yaml and kubernetes manifests are also provided.

### Running

Running the exporter requires fio and the libaio development packages to be installed on the host.

```
./yafio_exporter <flags>
```

For a kubernetes deployment edit kustomization.yaml as needed (you will probably need to change the storageClass in resources/pvc.yaml) and apply the resources:

```
kustomize build | kubectl apply -f -
```

#### Usage

```
./yafio_exporter -h
```

#### Flags

```bash
$ ./yafio_exporter -h
Usage of ./yafio_exporter:
  -flags string
        flags passed to fio
  -name string
        Job name (default "benchmark")
  -noExporter
        Disables the prometheus endpoint
  -port string
        metrics tcp listen port (default "9996")
  -preset string
        Preset name (iops, latency or throughput)
  -runtime string
        Runtime (default "60")
  -statusUpdateInterval string
        metric update interval in seconds when statusUpdates enabled (default "6")
  -testid string
        Test ID label to use with K6
```
#### Predefined Benchmarks

| Name             | Equivalent fio command when used with all defaults |
|------------------|-----------------------------------------------------------------------------------------------------------------|
| iops             | fio --name=iops --numjobs=4 --ioengine=libaio --direct=1 --bs=4k --iodepth=128 --readwrite=randrw --size=1G     |
| latency          | fio --name=latency --numjobs=1 --ioengine=libaio --direct=1 --bs=4k --iodepth=1 --readwrite=randrw --size=1G    |
| throughput       | fio --name=throughput --numjobs=4 --ioengine=libaio --direct=1 --bs=128k --iodepth=64 --readwrite=rw  --size=1G |