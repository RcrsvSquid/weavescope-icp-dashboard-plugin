package main

import (
	"fmt"
	"io/ioutil"
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

func main() {
	// We put the socket in a sub-directory to have more control on the permissions
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
	if err := http.Serve(listener, nil); err != nil {
		log.Printf("error: %v", err)
	}
}

// Report is called by scope when a new report is needed. It is part of the
// "reporter" interface, which all plugins must implement.
func Report(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.String())

	raw, err := ioutil.ReadFile("./test.json")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatalf("File read error %v", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(raw)
}

// func (p *Plugin) getTopologyHost() string {
// 	return fmt.Sprintf("%s;<host>", p.HostID)
// }

// type report struct {
// 	Host struct {
// 		Nodes struct {
// 			SidneywWorker2Host struct {
// 				Latest struct {
// 					LinkTableMetaData struct {
// 						Value     string    `json:"value"`
// 						Timestamp time.Time `json:"timestamp"`
// 					} `json:"link-table-meta-data"`
// 				} `json:"latest"`
// 			} `json:"sidneyw-worker-2;<host>"`
// 		} `json:"nodes"`
// 		MetadataTemplates struct {
// 			LinkMeta struct {
// 				ID       string  `json:"id"`
// 				Label    string  `json:"label"`
// 				DataType string  `json:"dataType"`
// 				Priority float64 `json:"priority"`
// 				From     string  `json:"from"`
// 			} `json:"link-meta"`
// 		} `json:"metadata_templates"`
// 		TableTemplates struct {
// 			LinkTable struct {
// 				ID     string `json:"id"`
// 				Label  string `json:"label"`
// 				Prefix string `json:"prefix"`
// 			} `json:"link-table"`
// 		} `json:"table_templates"`
// 	} `json:"Host"`
// 	Plugins []struct {
// 		ID          string   `json:"id"`
// 		Label       string   `json:"label"`
// 		Description string   `json:"description"`
// 		Interfaces  []string `json:"interfaces"`
// 		APIVersion  string   `json:"api_version"`
// 	} `json:"Plugins"`
// }
