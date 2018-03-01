package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	app_v1 "k8s.io/api/apps/v1"
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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

func (p *Plugin) HandleReport(w http.ResponseWriter, r *http.Request) {
	raw, err := json.Marshal(&p.Report)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatalf("JSON Marshall Error %v", err)
	}

	debug(func() {
		jsonIndented, _ := prettyprint(raw)
		fmt.Printf("Report\n%s\n", jsonIndented)
	})

	w.WriteHeader(http.StatusOK)
	w.Write(raw)
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
	client := getK8sClient()

	done := make(chan bool)

	for _, retriever := range k8sRetrievers {
		go retriever(client, &p.Report, done)
	}

	// Wait for all the queries to exit
	for range k8sRetrievers {
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

func getK8sClient() *kubernetes.Clientset {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return client
}

type k8sRetriever func(*kubernetes.Clientset, *WeaveReport, chan<- bool)

func getDaemonSets(client *kubernetes.Clientset, rpt *WeaveReport, done chan<- bool) {
	daemonsets, _ := client.Apps().DaemonSets("").List(meta_v1.ListOptions{})

	for _, k8sobject := range daemonsets.Items {
		rpt.AddToReport(&k8sobject)
	}

	done <- true
}

func getDeployments(client *kubernetes.Clientset, rpt *WeaveReport, done chan<- bool) {
	deployments, _ := client.Apps().Deployments("").List(meta_v1.ListOptions{})

	for _, k8sobject := range deployments.Items {
		rpt.AddToReport(&k8sobject)
	}

	done <- true
}

func getServices(client *kubernetes.Clientset, rpt *WeaveReport, done chan<- bool) {
	services, _ := client.CoreV1().Services("").List(meta_v1.ListOptions{})
	for _, k8sobject := range services.Items {
		// fmt.Printf("\n\nFound services %s\n", k8sobject.GetName())
		annotations := k8sobject.GetAnnotations()
		if url, ok := annotations["console"]; ok {
			// url := fmt.Sprintf("console=%s", url)
			fmt.Printf("\n\nFound annotations %s\n", url)
		}
		rpt.AddToReport(&k8sobject)
	}

	done <- true
}

func getStatefulSets(client *kubernetes.Clientset, rpt *WeaveReport, done chan<- bool) {
	statefulsets, _ := client.Apps().StatefulSets("").List(meta_v1.ListOptions{})
	for _, k8sobject := range statefulsets.Items {
		rpt.AddToReport(&k8sobject)
	}

	done <- true
}

var k8sRetrievers = [...]k8sRetriever{
	getDaemonSets,
	getDeployments,
	getStatefulSets,
	getServices,
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

// Helper Functions

func prettyprint(b []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "  ")
	return out.Bytes(), err
}

func debug(do func()) {
	isDebug, ok := os.LookupEnv("DEBUG")
	if ok && isDebug == "true" {
		do()
	}
}
