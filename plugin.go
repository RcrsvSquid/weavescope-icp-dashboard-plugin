package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	app_v1 "k8s.io/api/apps/v1"
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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

	DaemonSet Topology `json:"DaemonSet,omitempty"`

	Service Topology `json:"Service,omitempty"`

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

func prettyprint(b []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "  ")
	return out.Bytes(), err
}

func (p *Plugin) HandleReport(w http.ResponseWriter, r *http.Request) {
	// log.Printf("HandleReport ...\n")
	rpt := p.Report

	raw, err := json.Marshal(&rpt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatalf("JSON Marshall Error %v", err)
	}

	// jsonIndented, _ := prettyprint(raw)
	// fmt.Printf("%s\n", jsonIndented)

	w.WriteHeader(http.StatusOK)
	w.Write(raw)
}

// TODO: Make this work
func (p *Plugin) GetTopLevelK8S(clientset *kubernetes.Clientset, controllers []K8SObject) {
	for _, k8sobject := range controllers {
		// fmt.Printf("\n\nFound Controller %s\n", k8sobject.GetName())

		p.Report.AddToReport(k8sobject)
		labels := k8sobject.GetLabels()

		if appName, ok := labels["app"]; ok {
			selector := fmt.Sprintf("app=%s", appName)

			pods, err := clientset.
				CoreV1().
				Pods("").
				List(meta_v1.ListOptions{LabelSelector: selector})

			if err != nil {
				panic(err.Error())
			}

			fmt.Printf("There are %d pods in %s\n", len(pods.Items), appName)

			for _, pod := range pods.Items {
				fmt.Printf("Found pod %s\n", pod.GetName())
				p.Report.AddToReport(&pod)
			}
		}
	}
}

func handleControllerErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func (w *WeaveReport) AddToReport(obj K8SObject) {
	switch obj.(type) {
	case *Host:
		w.Host.Add(obj)

	case *app_v1.Deployment:
		w.Deployment.Add(obj)

	case *app_v1.DaemonSet:
		w.DaemonSet.Add(obj)

	case *core_v1.Service:
		w.Service.Add(obj)

	case *core_v1.Pod:
		w.Pods.Add(obj)
	}
}

func (top *Topology) Add(obj K8SObject) {
	top.AddLatest(obj)
	// fmt.Printf("%T\n", obj)
	// top.AddMetadataTemplate(obj)
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
