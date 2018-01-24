package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

func setupSocket(socketPath string) (net.Listener, error) {
	os.RemoveAll(filepath.Dir(socketPath))
	if err := os.MkdirAll(filepath.Dir(socketPath), 0700); err != nil {
		return nil, fmt.Errorf("failed to create directory %q: %v", filepath.Dir(socketPath), err)
	}
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %q: %v", socketPath, err)
	}

	log.Printf("Listening on: unix://%s", socketPath)
	return listener, nil
}

func setupSignals(socketPath string) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-interrupt
		os.RemoveAll(filepath.Dir(socketPath))
		os.Exit(0)
	}()
}

// Report generates a json report consumable by the weavescope application
func Report(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.String())

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{
	  "Plugins": [
		{
		  "id":          "scope-test",
		  "label":       "Scope Test",
		  "description": "Testing the weave scope plugins",
		  "interfaces":  ["reporter"],
		  "api_version": "1"
		}
	  ]
	}`))
}

func main() {
	const socketPath = "/var/run/scope/plugins/scope-test/scope-test.sock"
	hostID, _ := os.Hostname()

	// Handle the exit signal
	setupSignals(socketPath)

	log.Printf("Starting on %s...\n", hostID)

	listener, err := setupSocket(socketPath)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		listener.Close()
		os.RemoveAll(filepath.Dir(socketPath))
	}()

	http.HandleFunc("/report", Report)
	// http.HandleFunc("/control", Control)

	if err := http.Serve(listener, nil); err != nil {
		log.Printf("error: %v", err)
	}
}

type report struct {
	Host struct {
		Nodes           map[string]node           `json:"nodes"`
		MetricTemplates map[string]metricTemplate `json:"metric_templates"`
		Controls        map[string]control        `json:"controls"`
	}
	Plugins []struct {
		ID          string   `json:"id"`
		Label       string   `json:"label"`
		Description string   `json:"description,omitempty"`
		Interfaces  []string `json:"interfaces"`
		APIVersion  string   `json:"api_version,omitempty"`
	}
}
