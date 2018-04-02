package main

import (
	"fmt"
	"time"
)

type MetricSample struct {
	Value float64   `json:"value"`
	Date  time.Time `json:"date"`
}

type MetricData struct {
	Samples []MetricSample `json:"samples"`
	Min     float64        `json:"min"`
	Max     float64        `json:"max"`
}

func (data *MetricData) AddSample(sample MetricSample) {
	// Maintain Min and Max
	if sample.Value > data.Max {
		data.Max = sample.Value
	}

	if sample.Value < data.Min {
		data.Min = sample.Value
	}

	data.Samples = append(data.Samples, sample)
}

type LatestSample struct {
	Value     string    `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

type Node struct {
	Latest  map[string]LatestSample `json:"latest,omitempty"`
	Metrics map[string]MetricData   `json:"metrics,omitempty"`
}

type Metric struct {
	ID       string  `json:"id"`
	Label    string  `json:"label"`
	Format   string  `json:"format"`
	Priority float32 `json:"priority"`
	Min      int     `json:"min,omitempty"`
	Max      int     `json:"max,omitempty"`
}

type Metadata struct {
	ID       string  `json:"id"`
	Label    string  `json:"label"`
	DataType string  `json:"dataType"`
	Priority float32 `json:"priority"`
	From     string  `json:"from"`
}

type TableColumn struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	DataType string `json:"dataType"`
}

type Table struct {
	ID      string        `json:"id"`
	Label   string        `json:"label"`
	Prefix  string        `json:"prefix"`
	Type    string        `json:"type"`
	Columns []TableColumn `json:"columns,omitempty"`
}

type Topology struct {
	Nodes             map[string]*Node    `json:"nodes,omitempty"`
	MetadataTemplates map[string]Metadata `json:"metadata_templates,omitempty"`
	TableTemplates    map[string]Table    `json:"table_templates,omitempty"`
	MetricTemplates   map[string]Metric   `json:"metric_templates,omitempty"`
}

type PluginSpec struct {
	ID          string   `json:"id"`
	Label       string   `json:"label"`
	Description string   `json:"description,omitempty"`
	Interfaces  []string `json:"interfaces"`
	APIVersion  int      `json:"api_version,omitempty"`
}

type WeaveReport struct {
	Plugins []PluginSpec `json:"Plugins"`

	DaemonSet   Topology `json:"DaemonSet,omitempty"`
	Deployment  Topology `json:"Deployment,omitempty"`
	Pods        Topology `json:"Pods,omitempty"`
	Service     Topology `json:"Service,omitempty"`
	StatefulSet Topology `json:"StatefulSet,omitempty"`
}

// Retrieves a nested node from a topology by id
func (top *Topology) getNode(weaveID string) *Node {
	if len(top.Nodes) == 0 {
		top.Nodes = make(map[string]*Node)
	}

	if _, ok := top.Nodes[weaveID]; !ok {
		top.Nodes[weaveID] = &Node{
			Latest:  make(map[string]LatestSample),
			Metrics: make(map[string]MetricData),
		}
	}

	return top.Nodes[weaveID]
}

// Adds a metric sample to an existing metric data object
const MAX_METRICS int = 50

func (top *Topology) AddMetric(weaveID string, metricKey string, sample MetricSample) {
	node := top.getNode(weaveID)

	fmt.Printf("Adding: %s %s %+v\n\n", weaveID, metricKey, sample.Value)
	var metricData MetricData
	var ok bool

	if metricData, ok = node.Metrics[metricKey]; !ok {
		metricData = MetricData{}
	}

	if len(metricData.Samples) > MAX_METRICS {
		newMetricData := MetricData{}
		for i := MAX_METRICS / 2; i < len(metricData.Samples); i++ {
			newMetricData.AddSample(metricData.Samples[i])
		}

		node.Metrics[metricKey] = newMetricData
	} else {
		metricData.AddSample(sample)
		node.Metrics[metricKey] = metricData
	}
}

// Add a metric data object
func (top *Topology) AddMetricData(weaveID string, metricKey string, metricData MetricData) {
	node := top.getNode(weaveID)
	node.Metrics[metricKey] = metricData

	fmt.Printf("%+v\n%+v\n", top, node.Metrics[metricKey])
}

func (top *Topology) AddLatest(weaveID string, latestKey string, latest LatestSample) {
	node := top.getNode(weaveID)
	node.Latest[latestKey] = latest
}

func (top *Topology) AddMetricTemplate(metricID string, metricTemplate Metric) {
	if len(top.MetricTemplates) == 0 {
		top.MetricTemplates = make(map[string]Metric)
	}

	top.MetricTemplates[metricID] = metricTemplate
}

func (top *Topology) AddMetadataTemplate(metaID string, metaTableTemplate Metadata) {
	if len(top.MetadataTemplates) == 0 {
		top.MetadataTemplates = make(map[string]Metadata)
	}

	top.MetadataTemplates[metaID] = metaTableTemplate
}

func (top *Topology) AddTableTemplate(tableID string, tableTemplate Table) {
	if len(top.TableTemplates) == 0 {
		top.TableTemplates = make(map[string]Table)
	}

	top.TableTemplates[tableID] = tableTemplate
}
