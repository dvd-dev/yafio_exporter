package main

import "github.com/prometheus/client_golang/prometheus"

type MetricsJobs struct {
	JobName           string
	GroupID           int
	Error             prometheus.GaugeVec
	Eta               prometheus.GaugeVec
	Elapsed           prometheus.GaugeVec
	JobOptions        MetricsJobOptions
	Read              MetricsStats
	Write             MetricsStats
	Trim              MetricsStats
	Sync              MetricsStats
	JobRuntime        prometheus.GaugeVec
	UsrCpu            prometheus.GaugeVec
	SysCpu            prometheus.GaugeVec
	Ctx               prometheus.GaugeVec
	MajF              prometheus.GaugeVec
	MinF              prometheus.GaugeVec
	IoDepthLevel      MetricsDepth
	IoDepthSubmit     MetricsDepth
	IoDepthComplete   MetricsDepth
	LatencyNs         MetricsLatency
	LatencyUs         MetricsLatency
	LatencyMs         MetricsLatency
	LatencyDepth      prometheus.GaugeVec
	LatencyTarget     prometheus.GaugeVec
	LatencyPercentile prometheus.GaugeVec
	LatencyWindow     prometheus.GaugeVec
}
type MetricsJobOptions struct {
	Name     string
	BS       string
	IoDepth  string
	Size     string
	RW       string
	RampTime string
	RunTime  string
}
type MetricsStats struct {
	IOBytes     prometheus.GaugeVec
	IOKBytes    prometheus.GaugeVec
	BWBytes     prometheus.GaugeVec
	BW          prometheus.GaugeVec
	Iops        prometheus.GaugeVec
	Runtime     prometheus.GaugeVec
	TotalIos    prometheus.GaugeVec
	ShortIos    prometheus.GaugeVec
	DropIos     prometheus.GaugeVec
	SlatNs      MetricsNS
	ClatNs      MetricsNS
	LatNs       MetricsNS
	BwMin       prometheus.GaugeVec
	BwMax       prometheus.GaugeVec
	BwAgg       prometheus.GaugeVec
	BwMean      prometheus.GaugeVec
	BwDev       prometheus.GaugeVec
	BwSamples   prometheus.GaugeVec
	IopsMin     prometheus.GaugeVec
	IopsMax     prometheus.GaugeVec
	IopsMean    prometheus.GaugeVec
	IopsStdDev  prometheus.GaugeVec
	IopsSamples prometheus.GaugeVec
}
type MetricsNS struct {
	Min        prometheus.GaugeVec
	Max        prometheus.GaugeVec
	Mean       prometheus.GaugeVec
	StdDev     prometheus.GaugeVec
	N          prometheus.GaugeVec
	Percentile MetricsLatPercentile
}
type MetricsLatPercentile struct {
	Percentile100  prometheus.GaugeVec
	Percentile500  prometheus.GaugeVec
	Percentile1000 prometheus.GaugeVec
	Percentile2000 prometheus.GaugeVec
	Percentile3000 prometheus.GaugeVec
	Percentile4000 prometheus.GaugeVec
	Percentile5000 prometheus.GaugeVec
	Percentile6000 prometheus.GaugeVec
	Percentile7000 prometheus.GaugeVec
	Percentile8000 prometheus.GaugeVec
	Percentile9000 prometheus.GaugeVec
	Percentile9500 prometheus.GaugeVec
	Percentile9900 prometheus.GaugeVec
	Percentile9950 prometheus.GaugeVec
	Percentile9990 prometheus.GaugeVec
	Percentile9995 prometheus.GaugeVec
	Percentile9999 prometheus.GaugeVec
}
type MetricsDepth struct {
	FioDepth0    prometheus.GaugeVec
	FioDepth1    prometheus.GaugeVec
	FioDepth2    prometheus.GaugeVec
	FioDepth4    prometheus.GaugeVec
	FioDepth8    prometheus.GaugeVec
	FioDepth16   prometheus.GaugeVec
	FioDepth32   prometheus.GaugeVec
	FioDepth64   prometheus.GaugeVec
	FioDepthGE64 prometheus.GaugeVec
}
type MetricsLatency struct {
	FioLat2      prometheus.GaugeVec
	FioLat4      prometheus.GaugeVec
	FioLat10     prometheus.GaugeVec
	FioLat20     prometheus.GaugeVec
	FioLat50     prometheus.GaugeVec
	FioLat100    prometheus.GaugeVec
	FioLat250    prometheus.GaugeVec
	FioLat500    prometheus.GaugeVec
	FioLat750    prometheus.GaugeVec
	FioLat1000   prometheus.GaugeVec
	FioLat2000   prometheus.GaugeVec
	FioLatGE2000 prometheus.GaugeVec
}

type MetricsDiskUtil struct {
	Name        string
	ReadIos     prometheus.GaugeVec
	WriteIos    prometheus.GaugeVec
	ReadMerges  prometheus.GaugeVec
	WriteMerges prometheus.GaugeVec
	ReadTicks   prometheus.GaugeVec
	WriteTicks  prometheus.GaugeVec
	InQueue     prometheus.GaugeVec
	Util        prometheus.GaugeVec
}
