package main

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type K8sQuery func(*kubernetes.Clientset, func(K8sObject))

var K8sQueries = [...]K8sQuery{
	MapDaemonSets,
	MapDeployments,
	MapStatefulSets,
	MapServices,
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

func MapDaemonSets(client *kubernetes.Clientset, do func(K8sObject)) {
	daemonsets, _ := client.Apps().DaemonSets("").List(meta_v1.ListOptions{})

	for _, k8sobject := range daemonsets.Items {
		do(&k8sobject)
	}
}

func MapDeployments(client *kubernetes.Clientset, do func(K8sObject)) {
	deployments, _ := client.Apps().Deployments("").List(meta_v1.ListOptions{})

	for _, k8sobject := range deployments.Items {
		do(&k8sobject)
	}
}

func MapServices(client *kubernetes.Clientset, do func(K8sObject)) {
	services, _ := client.CoreV1().Services("").List(meta_v1.ListOptions{})

	for _, k8sobject := range services.Items {
		do(&k8sobject)
	}
}

func MapStatefulSets(client *kubernetes.Clientset, do func(K8sObject)) {
	statefulsets, _ := client.Apps().StatefulSets("").List(meta_v1.ListOptions{})

	for _, k8sobject := range statefulsets.Items {
		do(&k8sobject)
	}
}
