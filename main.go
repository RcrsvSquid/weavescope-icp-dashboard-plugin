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
	"time"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	var plugin = &Plugin{
		ID:          "icp-dashboard",
		Label:       "ICP Dashboard",
		Description: "Links into the ICP Dashboard",
		Interfaces:  []string{"reporter"},
		APIVersion:  1,

		Controls: Controls{false, false},
	}

	hostID, _ := os.Hostname()
	log.Printf("Starting on %s...\n", hostID)

	// We put the socket in a sub-directory to have more control on the permissions
	socketPath := fmt.Sprintf("/var/run/scope/plugins/%s/%s.sock", plugin.ID, plugin.ID)

	// Handle the exit signal
	setupSignals(socketPath)

	listener, err := setupSocket(socketPath)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		listener.Close()
		os.RemoveAll(filepath.Dir(socketPath))
	}()

	go startPollingK8sObject(plugin)

	http.HandleFunc("/report", plugin.HandleReport)
	if err := http.Serve(listener, nil); err != nil {
		log.Printf("error: %v", err)
	}
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

func startPollingK8sObject(p *Plugin) {
	for {
		time.Sleep(10 * time.Second)
		generateReport(p)
	}
}

type k8sRetriever func(clientset *kubernetes.Clientset, rpt *WeaveReport, done chan<- bool)

func getDaemonSets(clientset *kubernetes.Clientset, rpt *WeaveReport, done chan<- bool) {
	daemonsets, _ := clientset.Apps().DaemonSets("").List(meta_v1.ListOptions{})

	for _, k8sobject := range daemonsets.Items {
		rpt.AddToReport(&k8sobject)
	}

	done <- true
}

func getDeployments(clientset *kubernetes.Clientset, rpt *WeaveReport, done chan<- bool) {
	deployments, _ := clientset.Apps().Deployments("").List(meta_v1.ListOptions{})

	for _, k8sobject := range deployments.Items {
		rpt.AddToReport(&k8sobject)
	}

	done <- true
}

func getServices(clientset *kubernetes.Clientset, rpt *WeaveReport, done chan<- bool) {
	services, _ := clientset.CoreV1().Services("").List(meta_v1.ListOptions{})
	for _, k8sobject := range services.Items {
		// fmt.Printf("\n\nFound services %s\n", k8sobject.GetName())
		annotations := k8sobject.GetAnnotations()
		if url, ok := annotations["adminConsoleUrl"]; ok {
			url := fmt.Sprintf("adminConsoleUrl=%s", url)
			fmt.Printf("\n\nFound annotations %s\n", url)
			rpt.AddToReport(&k8sobject)
		}
	}

	done <- true
}

func getStatefulSets(clientset *kubernetes.Clientset, rpt *WeaveReport, done chan<- bool) {
	statefulsets, _ := clientset.Apps().StatefulSets("").List(meta_v1.ListOptions{})
	for _, k8sobject := range statefulsets.Items {
		rpt.AddToReport(&k8sobject)
	}

	done <- true
}

func generateReport(p *Plugin) {
	startTime := time.Now()
	// var hostNode = &Host{}
	// hostNode.Init()

	k8sRetrievers := [...]k8sRetriever{
		getDaemonSets,
		getDeployments,
		getStatefulSets,
		getServices,
	}

	rpt := WeaveReport{Plugins: []PluginSpec{{
		ID:          p.ID,
		Label:       p.Label,
		Description: p.Description,
		Interfaces:  p.Interfaces,
		APIVersion:  p.APIVersion,
	}}}

	// rpt.AddToReport(hostNode)

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	done := make(chan bool)

	for _, retriever := range k8sRetrievers {
		go retriever(clientset, &rpt, done)
	}

	// Wait for all the queries to exit
	for range k8sRetrievers {
		<-done
	}

	log.Printf("Probe finished in %v...\n", time.Since(startTime))
	p.Report = rpt
}
