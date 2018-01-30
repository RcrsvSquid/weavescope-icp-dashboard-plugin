package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	weave "github.ibm.com/sidney-wijngaarde1/weave-scope-plugin/weaveplugin"
)

func main() {
	// We put the socket in a sub-directory to have more control on the permissions
	const socketPath = "/var/run/scope/plugins/icp-dashboard/icp-dashboard.sock"
	hostID, _ := os.Hostname()

	fmt.Printf("Current Host IP %s\n", weave.GetOutboundIP())

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

	var rpt *weave.WeaveReport = &weave.WeaveReport{}

	rpt.Plugins = []weave.PluginSpec{{
		ID:          "icp-dashboard",
		Label:       "ICP Dashboard",
		Description: "Links into the ICP Dashboard",
		Interfaces:  []string{"reporter"},
		APIVersion:  1,
	}}

	var hostNode *weave.Host = &weave.Host{}
	hostNode.Init()

	rpt.Host = weave.Topology{
		Nodes: map[string]weave.Node{
			weave.GetWeaveID(hostNode): weave.Node{
				Latest: weave.GetLatestURL(hostNode),
			},
		},
		MetadataTemplates: map[string]weave.Metadata{
			weave.GetMetaDataTableID(hostNode): weave.GetWeaveMetaData(hostNode),
		},
		TableTemplates: map[string]weave.Table{
			hostNode.GetTableID(): weave.GetWeaveTable(hostNode),
		},
	}

	raw, err := json.Marshal(rpt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatalf("JSON Marshall Error %v", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(raw)
}

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
