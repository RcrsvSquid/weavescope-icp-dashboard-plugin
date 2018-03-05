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

type Plugin struct {
	ID          string   `json:"id"`
	Label       string   `json:"label"`
	Description string   `json:"description,omitempty"`
	Interfaces  []string `json:"interfaces"`
	APIVersion  int      `json:"api_version,omitempty"`

	Report WeaveReport
}

// Return the correct topology for a given kubernetes object
func SelectTopoloy(w *WeaveReport, obj K8sObject) *Topology {
	var top *Topology

	switch obj.(type) {
	case *app_v1.Deployment:
		top = &w.Deployment

	case *app_v1.DaemonSet:
		top = &w.DaemonSet

	case *core_v1.Service:
		top = &w.Service

	case *app_v1.StatefulSet:
		top = &w.StatefulSet

	case *core_v1.Pod:
		top = &w.Pods
	}

	return top
}

// Compute and add the ICP link into a WeaveReport
func AddICPLink(w *WeaveReport, obj K8sObject) {
	top := SelectTopoloy(w, obj)

	latestID, latest := GetLatest(obj)
	weaveID, _ := GetWeaveID(obj)
	top.AddLatest(weaveID, latestID, latest)

	// tableID, tableTemplate := GetWeaveTable(obj)
	// top.AddTableTemplate(tableID, tableTemplate)

	metaID, metaTemplate := GetWeaveMetaData(obj)
	metaLatestID, metaLatest := GetMetaLatest(obj)
	top.AddMetadataTemplate(metaID, metaTemplate)
	top.AddLatest(weaveID, metaLatestID, metaLatest)
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

	client := GetK8sClient()

	do := func(k8sobject K8sObject) {
		AddICPLink(&p.Report, k8sobject)
	}

	done := make(chan bool)

	// Execute queries concurrently
	for _, k8sQuery := range K8sQueries {
		go func(query K8sQuery) {
			query(client, do)

			done <- true
		}(k8sQuery)
	}

	// Wait for all the queries to exit
	for range K8sQueries {
		<-done
	}

	log.Printf("Probe finished in %v\n", time.Since(startTime))
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

func (p *Plugin) pollK8s() {
	for {
		p.GenerateReport()
		time.Sleep(10 * time.Second)
	}
}
