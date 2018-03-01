package main

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type K8sQuery func(*kubernetes.Clientset, *WeaveReport, chan<- bool)

var K8sQueries = [...]K8sQuery{
	GetDaemonSets,
	GetDeployments,
	GetStatefulSets,
	GetServices,
}

func GetK8sClient() *kubernetes.Clientset {
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

func GetDaemonSets(client *kubernetes.Clientset, rpt *WeaveReport, done chan<- bool) {
	daemonsets, _ := client.Apps().DaemonSets("").List(meta_v1.ListOptions{})

	for _, k8sobject := range daemonsets.Items {
		rpt.AddToReport(&k8sobject)
	}

	done <- true
}

func GetDeployments(client *kubernetes.Clientset, rpt *WeaveReport, done chan<- bool) {
	deployments, _ := client.Apps().Deployments("").List(meta_v1.ListOptions{})

	for _, k8sobject := range deployments.Items {
		rpt.AddToReport(&k8sobject)
	}

	done <- true
}

func GetServices(client *kubernetes.Clientset, rpt *WeaveReport, done chan<- bool) {
	services, _ := client.CoreV1().Services("").List(meta_v1.ListOptions{})

	for _, k8sobject := range services.Items {
		rpt.AddToReport(&k8sobject)
	}

	done <- true
}

func GetStatefulSets(client *kubernetes.Clientset, rpt *WeaveReport, done chan<- bool) {
	statefulsets, _ := client.Apps().StatefulSets("").List(meta_v1.ListOptions{})

	for _, k8sobject := range statefulsets.Items {
		rpt.AddToReport(&k8sobject)
	}

	done <- true
}
