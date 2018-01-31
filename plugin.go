package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	v1 "k8s.io/api/apps/v1"
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
	Host Topology `json:"Host,omitempty"`

	Pods Topology `json:"Pods,omitempty"`

	Deployment Topology `json:"Deployment,omitempty"`

	Plugins []PluginSpec `json:"Plugins"`
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

func (p *Plugin) HandleReport(w http.ResponseWriter, r *http.Request) {
	rpt := p.GenerateReport()

	raw, err := json.Marshal(&rpt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatalf("JSON Marshall Error %v", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(raw)
}

func (p *Plugin) GenerateReport() WeaveReport {
	var hostNode *Host = &Host{}
	hostNode.Init()

	p.Report = WeaveReport{Plugins: []PluginSpec{{
		ID:          p.ID,
		Label:       p.Label,
		Description: p.Description,
		Interfaces:  p.Interfaces,
		APIVersion:  p.APIVersion,
	}}}

	p.Report.AddToReport(hostNode)

	// config, err := rest.InClusterConfig()
	// if err != nil {
	// 	panic(err.Error())
	// }
	// // creates the clientset
	// clientset, err := kubernetes.NewForConfig(config)
	// if err != nil {
	// 	panic(err.Error())
	// }

	// deployments, err := clientset.Apps().Deployments("").List(metav1.ListOptions{})
	// if err != nil {
	// 	panic(err.Error())
	// }

	// for _, deployment := range deployments.Items {
	// 	fmt.Println(deployment.GetName())
	// }

	// fmt.Printf("There are %d deployments in the cluster\n", len(deployments.Items))

	return p.Report
}

func (w *WeaveReport) AddToReport(obj K8SObject) {
	switch obj.(type) {
	case *Host:
		w.Host.Add(obj)

	case *v1.Deployment:
		w.Deployment.Add(obj)

		// 	case *Pod:
		// 		w.Pods.Add(obj)
	}
}

func (top *Topology) Add(obj K8SObject) {
	top.AddLatest(obj)
	top.AddMetadataTemplate(obj)
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
