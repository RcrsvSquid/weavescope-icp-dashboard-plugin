package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	app_v1 "k8s.io/api/apps/v1"
	core_v1 "k8s.io/api/core/v1"
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

type Controls struct {
	Chaos    bool
	LoadTest bool
}

type Plugin struct {
	ID          string   `json:"id"`
	Label       string   `json:"label"`
	Description string   `json:"description,omitempty"`
	Interfaces  []string `json:"interfaces"`
	APIVersion  int      `json:"api_version,omitempty"`

	Controls Controls
	Report   WeaveReport
}

func (p *Plugin) GenerateReport() {
	startTime := time.Now()

	p.Report = WeaveReport{Plugins: []PluginSpec{{
		ID:          p.ID,
		Label:       p.Label,
		Description: p.Description,
		Interfaces:  p.Interfaces,
		APIVersion:  p.APIVersion,
	}}}

	//  TODO: Create the client once as a constant <28-02-18, sidney> //
	client := GetK8sClient()

	done := make(chan bool)

	for _, query := range K8sQueries {
		go query(client, &p.Report, done)
	}

	// Wait for all the queries to exit
	for range K8sQueries {
		<-done
	}

	log.Printf("Probe finished in %v...\n", time.Since(startTime))
}

func (p *Plugin) pollK8s() {
	for {
		time.Sleep(10 * time.Second)
		p.GenerateReport()
	}
}

func (p *Plugin) HandleReport(w http.ResponseWriter, r *http.Request) {
	raw, err := json.Marshal(&p.Report)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatalf("JSON Marshall Error %v", err)
	}

	Debug(func() {
		jsonIndented, _ := PrettyFmt(raw)
		fmt.Printf("Report\n%s\n", jsonIndented)
	})

	w.WriteHeader(http.StatusOK)
	w.Write(raw)
}

func (w *WeaveReport) AddToReport(obj K8SObject) {
	switch obj.(type) {

	case *app_v1.Deployment:
		w.Deployment.Add(obj)

	case *app_v1.DaemonSet:
		w.DaemonSet.Add(obj)

	case *core_v1.Service:
		w.Service.Add(obj)

	case *app_v1.StatefulSet:
		w.StatefulSet.Add(obj)

	case *core_v1.Pod:
		w.Pods.Add(obj)
	}
}

func (top *Topology) Add(obj K8SObject) {
	top.AddLatest(obj)
	top.AddTableTemplate(obj)
}

func (top *Topology) AddLatest(obj K8SObject) {
	latestKey, latest := GetLatest(obj)
	weaveID, _ := GetWeaveID(obj)

	if len(top.Nodes) == 0 {
		top.Nodes = make(map[string]Node)
	}

	top.Nodes[weaveID] = Node{
		Latest: map[string]LatestSample{latestKey: latest},
	}
}

func (top *Topology) AddMetadataTemplate(obj K8SObject) {
	id, metaTableTemplate := GetWeaveMetaData(obj)

	if len(top.MetadataTemplates) == 0 {
		top.MetadataTemplates = make(map[string]Metadata)
	}

	top.MetadataTemplates[id] = metaTableTemplate
}

func (top *Topology) AddTableTemplate(obj K8SObject) {
	id, tableTemplate := GetWeaveTable(obj)

	if len(top.TableTemplates) == 0 {
		top.TableTemplates = make(map[string]Table)
	}

	top.TableTemplates[id] = tableTemplate
}
