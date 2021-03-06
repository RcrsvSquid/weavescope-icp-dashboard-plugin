package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	app_v1 "k8s.io/api/apps/v1"
	core_v1 "k8s.io/api/core/v1"
)

type Plugin struct {
	ID          string   `json:"id"`
	Label       string   `json:"label"`
	Description string   `json:"description,omitempty"`
	Interfaces  []string `json:"interfaces"`
	APIVersion  int      `json:"api_version,omitempty"`

	Report WeaveReport

	// This is a concurrently accessed data structure
	// acquire the lock before mutating
	sync sync.Mutex
}

func (p *Plugin) GenerateReport(queries []K8sQuery) {
	startTime := time.Now()

	p.WeaveReportInit()

	client := GetK8sClient()

	done := make(chan bool)

	// Execute queries concurrently
	for _, k8sQuery := range queries {
		go queryWorker(client, k8sQuery, p.syncAdd, done)
	}

	// Wait for all the queries to exit
	for range K8sQueries {
		<-done
	}

	log.Printf("Probe finished in %v\n", time.Since(startTime))
}

func (p *Plugin) pollK8s() {
	// Get the report before waiting
	p.GenerateReport(K8sQueries)

	ticker := time.NewTicker(10 * time.Second)

	for range ticker.C {
		p.GenerateReport(K8sQueries)
	}
}

func (p *Plugin) HandleReport(w http.ResponseWriter, r *http.Request) {
	p.sync.Lock()
	raw, err := json.Marshal(&p.Report)
	p.sync.Unlock()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatalf("JSON Marshall Error %v", err)
	}

	Debug(func() { p.LogReport() })

	w.WriteHeader(http.StatusOK)
	w.Write(raw)
}

// Create a WeaveReport for the current plugin and copy over the plugin config
func (p *Plugin) WeaveReportInit() {
	p.Report = WeaveReport{Plugins: []PluginSpec{{
		ID:          p.ID,
		Label:       p.Label,
		Description: p.Description,
		Interfaces:  p.Interfaces,
		APIVersion:  p.APIVersion,
	}}}
}

// Return the correct topology for a given kubernetes object
func SelectTopology(w *WeaveReport, obj K8sObject) *Topology {
	var top *Topology

	switch obj.(type) {
	case *app_v1.Deployment, *K8sMock:
		top = &w.Deployment

	case *app_v1.DaemonSet:
		top = &w.DaemonSet

	case *core_v1.Service:
		top = &w.Service

	case *app_v1.StatefulSet:
		top = &w.StatefulSet
	}

	return top
}

func (p *Plugin) syncAdd(obj K8sObject) {
	top := SelectTopology(&p.Report, obj)

	weaveID, _ := GetWeaveID(obj)

	metaID, metaTemplate := GetMetaTemplate()
	metaLatestID, metaLatest := GetMetaLatest(obj)

	// tableID, tableTemplate := GetWeaveTable(obj)
	// latestKey, latest := GetLatest(obj)

	p.sync.Lock()
	top.AddMetadataTemplate(metaID, metaTemplate)
	top.AddLatest(weaveID, metaLatestID, metaLatest)

	// top.AddLatest(weaveID, latestKey, latest)
	// top.AddTableTemplate(tableID, tableTemplate)
	p.sync.Unlock()
}

func (p *Plugin) LogReport() {
	raw, _ := json.Marshal(&p.Report)
	jsonIndented, _ := PrettyFmt(raw)
	fmt.Printf("Report\n%s\n", jsonIndented)
}
