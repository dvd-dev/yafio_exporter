package main

import (
	"fmt"
	"go/types"
	"log"
	"reflect"
	"regexp"
	"slices"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

var prefix = "fio"
var promRegistry = prometheus.NewRegistry()
var pkg, _ = loadPackage(".")
var metricsJobsStruct = loadStruct(pkg, "MetricsJobs")
var numbersRex = regexp.MustCompile(`(int|float)[\d]*`)
var latnsRex = regexp.MustCompile("^(?<kind>[C|S]*[L|l]atNs)(?<type>.*)")
var perclatnsRex = regexp.MustCompile("^Percentile(?<kind>[C|S]*[L|l]atNs)(?<type>.*)")
var percentileRex = regexp.MustCompile("^Percentile(?<units>[0-9]{1,2})(?<float>[0-9]{2})")
var fioLatRex = regexp.MustCompile("^FioLat(?<op>GE)*(?<num>[0-9]+)")
var fioDepthRex = regexp.MustCompile("^FioDepth(?<op>GE)*(?<num>[0-9]+)")
var labelsFromMetricsRex = regexp.MustCompile("((?<metric>^percentile_[s|c]*lat_(ns|us|ms)_(read|write|trim|sync))_percentile(?<bucket>[0-9]+)|^((?<metric>io_depth_(complete|level|submit))_depth|(?<metric>latency_(us|ms|ns)_lat))[_]*(?<bucket>[0-9]+|ge[0-9]+))$")
var gauges = make(map[string]prometheus.GaugeVec)
var metricsMap = make(map[string][]string)
var labelsMap = make(map[string]string)

func Build(l *FioResult, testid string) {
	prometheus.NewRegistry()
	if len(testid) > 0 {
		labelsMap["testid"] = testid
	}
	parseStruct((l.Jobs)[0], metricsJobsStruct, "", "", true, testid)
	for j := 0; j < len(l.Jobs); j++ {
		var f = (l.Jobs)[j]
		parseStruct(f, metricsJobsStruct, "", "", false, testid)
	}
}

func parseStruct(f interface{}, s *types.Struct, struct_name string, parent_struct string, generate bool, testid string) {
	fields := reflect.TypeOf(f)
	values := reflect.ValueOf(f)
	num := fields.NumField()
	parent_name := fmt.Sprintf("%s%s", struct_name, parent_struct)
	for i := 0; i < num; i++ {
		field := fields.Field(i)
		value := values.Field(i)
		generateStruct(field, value, parent_name, s, generate, testid)
	}
}

func generateStruct(field reflect.StructField, value reflect.Value, parent_name string, s *types.Struct, generate bool, testid string) {
	metric_name := generateMetricName(field.Name, parent_name)
	new_metric_name, bucket, additionalLabels := getBuckets(metric_name)
	if len(additionalLabels) > 0 {
		log.Printf("Using new_metric_name %v instead of %v", new_metric_name, metric_name)
		metric_name = new_metric_name
	} else {
		log.Printf("No change to metric_name %v (parent_name: %v field.Name: %v value: %v)", metric_name, parent_name, field, value)
	}
	valueType := value.Type().String()
	isNumber := numbersRex.FindStringIndex(valueType)
	val := float64(0)
	if strings.Contains(valueType, "float") {
		val = value.Float()
	} else if len(isNumber) > 0 {
		val = float64(value.Int())
	}
	lowerFieldName := strings.ToLower(field.Name)
	converted, err := fieldByName(field.Name, s)
	convertedType := converted.Type().String()
	if err != nil {
		log.Fatalf("Unable to find field %s from %v", field.Name, s.String())
	}
	if len(isNumber) > 0 && strings.Contains(convertedType, "GaugeVec") {
		if !inMap(metric_name, metricsMap[parent_name]) {
			metricsMap[parent_name] = append(metricsMap[parent_name], metric_name)
		} else if generate {
			return
		} else if gauge, ok := gauges[metric_name]; ok {
			if len(testid) > 0 {
				if len(additionalLabels) > 0 {
					gauge.WithLabelValues(labelsMap["bs"], labelsMap["jobname"], labelsMap["iodepth"], labelsMap["size"], labelsMap["rw"], labelsMap["testid"], bucket).Set(val)
				} else {
					gauge.WithLabelValues(labelsMap["bs"], labelsMap["jobname"], labelsMap["iodepth"], labelsMap["size"], labelsMap["rw"], labelsMap["testid"]).Set(val)
				}
			} else {
				if len(additionalLabels) > 0 {
					gauge.WithLabelValues(labelsMap["bs"], labelsMap["jobname"], labelsMap["iodepth"], labelsMap["size"], labelsMap["rw"], bucket).Set(val)
				} else {
					gauge.WithLabelValues(labelsMap["bs"], labelsMap["jobname"], labelsMap["iodepth"], labelsMap["size"], labelsMap["rw"]).Set(val)
				}
			}
			return
		} else {
			gauges[metric_name] = generateGaugeVec(field.Name, parent_name)
		}
	} else if len(isNumber) > 0 {
		return
	} else if strings.Contains(convertedType, "Metrics") {
		parseStruct(value.Interface(), converted.Type().Underlying().(*types.Struct), field.Name, parent_name, generate, testid)
	} else if convertedType == "string" {
		_, ok := labelsMap[lowerFieldName]
		if !ok {
			labelsMap[lowerFieldName] = value.String()
		}
	} else {
		log.Fatalf("Unknown type of field: %s => %v (%v)", field.Name, convertedType, valueType)
	}
}

func generateMetricName(name string, structName string) string {
	snake_name := ToSnakeCase(name)
	struct_name := ToSnakeCase(strings.ReplaceAll(structName, "Metrics", ""))
	metric_name := snake_name
	if len(structName) > 0 {
		metric_name = fmt.Sprintf("%s_%s", struct_name, snake_name)
	}
	// dirty hack
	metric_name = strings.ReplaceAll(
		strings.ReplaceAll(metric_name, "ok_bytes", "o_kbytes"),
		"_fio", "")
	metric_name = strings.ReplaceAll(metric_name, "_b_w", "_bw")
	metric_name = strings.ReplaceAll(metric_name, "_i_o_", "_io_")
	return metric_name
}
func getBuckets(metric_name string) (string, string, []string) {
	fromMetricsMatch := labelsFromMetricsRex.FindStringSubmatch(metric_name)
	var additionalLabels []string
	var bucket string
	if len(fromMetricsMatch) > 0 {
		for i, name := range labelsFromMetricsRex.SubexpNames() {
			if len(fromMetricsMatch[i]) == 0 {
				continue
			}
			if name == "metric" {
				metric_name = fromMetricsMatch[i]
				additionalLabels = append(additionalLabels, "bucket")
			} else if name == "bucket" {
				bucket = fromMetricsMatch[i]
			}
		}
	}

	return metric_name, bucket, additionalLabels

}

func generateGaugeVec(name string, structName string) prometheus.GaugeVec {
	metric_name := generateMetricName(name, structName)
	new_metric_name, _, additionalLabels := getBuckets(metric_name)
	labels := slices.Concat(savedLabels, additionalLabels)
	help := getHelp(name, structName, metric_name)
	if len(additionalLabels) > 0 {
		metric_name = new_metric_name
	}
	log.Printf("Generating gauge for %v (%v) with parent %v: %s", name, metric_name, structName, help)
	gauge := *prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: prefix,
			Name:      metric_name,
			Help:      help,
		}, labels,
	)
	promRegistry.MustRegister(gauge)
	return gauge
}

func getHelp(name string, structName string, metric_name string) string {
	var helpSection strings.Builder
	helpSection.WriteString("Metrics")
	var help strings.Builder
	var unit string
	var latnsMatch = latnsRex.FindStringSubmatch(structName)
	var perclatnsMatch = perclatnsRex.FindStringSubmatch(structName)
	var percentileMatch = percentileRex.FindStringSubmatch(name)
	var fioLatMatch = fioLatRex.FindStringSubmatch(name)
	var fioDepthMatch = fioDepthRex.FindStringSubmatch(name)
	if structName == "" {
		helpSection.WriteString("Jobs")
	} else if structName == "Read" || structName == "Write" {
		helpSection.WriteString("Stats")
	} else if len(fioLatMatch) > 0 {
		operator := ""
		num := ""
		for i, name := range fioLatRex.SubexpNames() {
			if name == "op" {
				operator = ">="
			} else if name == "num" {
				num = fioLatMatch[i]
			}
		}
		if len(operator) == 0 {
			operator = "<"
		}
		help.WriteString(fmt.Sprintf("Overall Latency %s %s", operator, num))
	} else if len(fioDepthMatch) > 0 {
		operator := ""
		num := ""
		for i, name := range fioDepthRex.SubexpNames() {
			if name == "op" {
				operator = ">="
			} else if name == "num" {
				num = fioDepthMatch[i]
			}
		}
		if len(operator) == 0 {
			operator = "<"
		}
		help.WriteString(fmt.Sprintf("%s %s %s", structName, operator, num))
		unit = "io"
	} else if len(latnsMatch) > 0 {
		help.WriteString(latencyHelpString(latnsMatch, latnsRex))
	} else if len(perclatnsMatch) > 0 {
		help.WriteString(latencyHelpString(perclatnsMatch, perclatnsRex))
		var p strings.Builder
		for i, name := range percentileRex.SubexpNames() {
			if name == "units" {
				p.WriteString(percentileMatch[i])
			} else if name == "float" && percentileMatch[i] != "00" {
				p.WriteString(fmt.Sprintf(".%v", percentileMatch[i]))
			}
		}
		if p.String() == "1" {
			p.WriteString("st")
		} else {
			p.WriteString("th")
		}
		help.WriteString(fmt.Sprintf(" (%s percentile)", p.String()))
	}
	if help.Len() == 0 {
		help.WriteString(metricsConfig[helpSection.String()][name]["Help"])
	}
	if help.Len() == 0 {
		help.WriteString(metricsConfig["MetricsStats"][name]["Help"])
	}
	if strings.Contains(metric_name, "_samples") {
		unit = "sample"
	} else if strings.Contains(metric_name, "_ns_") {
		unit = "nsec"
	} else if strings.Contains(metric_name, "_ms_") {
		unit = "msec"
	} else if strings.Contains(metric_name, "_us_") {
		unit = "usec"
	} else if strings.Contains(metric_name, "_iops") {
		unit = "io/s"
	} else if strings.Contains(metric_name, "_ios") {
		unit = "io"
	} else if strings.HasSuffix(metric_name, "_f") {
		unit = "fault"
	} else if metric_name == "ctx" {
		unit = "context"
	} else if strings.HasSuffix(metric_name, "_cpu") {
		unit = "%"
	} else if strings.HasSuffix(metric_name, "_bytes") {
		unit = "bytes"
	} else if strings.HasSuffix(metric_name, "_kbytes") {
		unit = "kbytes"
	} else if strings.Contains(metric_name, "_bw") {
		unit = "kbytes"
	} else if strings.HasSuffix(metric_name, "runtime") {
		unit = "msec"
	}
	return fmt.Sprintf("%s (%s)", help.String(), unit)
}

func latencyHelpString(match []string, rex *regexp.Regexp) string {
	var kind string
	for i, name := range rex.SubexpNames() {
		if name == "kind" {
			kind = metricsConfig["MetricsStats"][match[i]]["Help"]
		} else if name == "type" {
			return fmt.Sprintf("%s %s", match[i], kind)
		}
	}
	return ""
}

var metricsConfig = map[string]map[string]map[string]string{
	"MetricsJobs": {
		"Error": {
			"Help": "Error",
		},
		"Eta": {
			"Help": "Eta",
		},
		"Elapsed": {
			"Help": "Elapsed",
		},
		"JobRuntime": {
			"Help": "JobRuntime",
		},
		"UsrCpu": {
			"Help": "User CPU",
		},
		"SysCpu": {
			"Help": "System CPU",
		},
		"Ctx": {
			"Help": "Context switching",
		},
		"MajF": {
			"Help": "Major Fault",
		},
		"MinF": {
			"Help": "Minor Fault",
		},
		"LatencyDepth": {
			"Help": "LatencyDepth",
		},
		"LatencyTarget": {
			"Help": "LatencyTarget",
		},
		"LatencyPercentile": {
			"Help": "LatencyPercentile",
		},
		"LatencyWindow": {
			"Help": "LatencyWindow",
		},
	},
	"MetricsStats": {
		"IOBytes":     {"Help": "IO Bytes"},
		"IOKBytes":    {"Help": "IO KBytes"},
		"BWBytes":     {"Help": "BW Bytes"},
		"BW":          {"Help": "BW KBytes"},
		"Iops":        {"Help": "IOPs"},
		"Runtime":     {"Help": "Runtime"},
		"TotalIos":    {"Help": "Total IOs"},
		"ShortIos":    {"Help": "Short IOs"},
		"DropIos":     {"Help": "Dropped IOs"},
		"SlatNs":      {"Help": "Submission Latency"},
		"ClatNs":      {"Help": "Completion Latency"},
		"LatNs":       {"Help": "Latency"},
		"BwMin":       {"Help": "Minimum Bandwidth"},
		"BwMax":       {"Help": "Maximum Bandwidth"},
		"BwAgg":       {"Help": "Bandwidth Aggregated"},
		"BwMean":      {"Help": "Mean Bandwidth"},
		"BwDev":       {"Help": "Bandwidth Deviation"},
		"BwSamples":   {"Help": "Bandwidth Samples"},
		"IopsMin":     {"Help": "Minimum IOPs"},
		"IopsMax":     {"Help": "Maximum IOPs"},
		"IopsMean":    {"Help": "Mean IOPs"},
		"IopsStdDev":  {"Help": "IOPs Deviation"},
		"IopsSamples": {"Help": "IOPs Samples"},
	},
	"MetricsNS": {
		"StdDev": {"Help": "Standard Deviation"},
	},
	"MetricsDiskUtil": {
		"ReadIos":     {"Help": "Read IOs"},
		"WriteIos":    {"Help": "Write IOs"},
		"ReadMerges":  {"Help": "Read Merges"},
		"WriteMerges": {"Help": "Write Merges"},
		"ReadTicks":   {"Help": "Read Ticks"},
		"WriteTicks":  {"Help": "Write Ticks"},
		"InQueue":     {"Help": "In Queue"},
		"Util":        {"Help": "Util"},
	},
}
