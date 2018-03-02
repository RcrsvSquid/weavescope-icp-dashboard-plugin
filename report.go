package main

import (
	"time"
)

type PluginSpec struct {
	ID          string   `json:"id"`
	Label       string   `json:"label"`
	Description string   `json:"description,omitempty"`
	Interfaces  []string `json:"interfaces"`
	APIVersion  int      `json:"api_version,omitempty"`
}

type LatestSample struct {
	Value     string    `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

type Node struct {
	Latest map[string]LatestSample `json:"latest"`
}

type Metadata struct {
	ID       string  `json:"id"`
	Label    string  `json:"label"`
	DataType string  `json:"dataType"`
	Priority float64 `json:"priority"`
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
	Nodes             map[string]Node     `json:"nodes"`
	MetadataTemplates map[string]Metadata `json:"metadata_templates,omitempty"`
	TableTemplates    map[string]Table    `json:"table_templates,omitempty"`
}

type WeaveReport struct {
	Plugins []PluginSpec `json:"Plugins"`

	DaemonSet   Topology `json:"DaemonSet,omitempty"`
	Deployment  Topology `json:"Deployment,omitempty"`
	Pods        Topology `json:"Pods,omitempty"`
	Service     Topology `json:"Service,omitempty"`
	StatefulSet Topology `json:"StatefulSet,omitempty"`
}

func (top *Topology) AddLatest(weaveID string, latestKey string, latest LatestSample) {
	if len(top.Nodes) == 0 {
		top.Nodes = make(map[string]Node)
	}

	top.Nodes[weaveID] = Node{
		Latest: map[string]LatestSample{latestKey: latest},
	}
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
