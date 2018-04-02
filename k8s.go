package main

import (
	"fmt"

	app_v1 "k8s.io/api/apps/v1"
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type K8sObject interface {
	GetAnnotations() map[string]string
	GetLabels() map[string]string
	GetName() string
	GetNamespace() string
	GetUID() types.UID
}

type K8sQuery func(*kubernetes.Clientset, func(K8sObject))

var K8sQueries = []K8sQuery{
	MapDaemonSets,
	MapDeployments,
	MapStatefulSets,
	MapServices,
}

func GetWeaveID(obj K8sObject) (string, error) {
	switch obj.(type) {
	case *app_v1.DaemonSet:
		return fmt.Sprintf("%s;<daemonset>", obj.GetUID()), nil

	case *app_v1.Deployment:
		return fmt.Sprintf("%s;<deployment>", obj.GetUID()), nil

	case *core_v1.Service:
		return fmt.Sprintf("%s;<service>", obj.GetUID()), nil

	case *app_v1.StatefulSet:
		return fmt.Sprintf("%s;<statefulset>", obj.GetUID()), nil

	case *K8sMock:
		return fmt.Sprintf("%s;<mock>", obj.GetUID()), nil

	default:
		return "", fmt.Errorf("No Compatible Type Found")
	}
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

func queryWorker(client *kubernetes.Clientset, query K8sQuery, do func(K8sObject), done chan<- bool) {
	query(client, do)
	done <- true
}

// A mock for testing that implements K8sObject
type K8sMock struct {
	Annotations map[string]string
	Labels      map[string]string
	Name        string
	Namespace   string
	UID         types.UID
}

func (k *K8sMock) GetAnnotations() map[string]string {
	return k.Annotations
}

func (k *K8sMock) GetLabels() map[string]string {
	return k.Labels
}

func (k *K8sMock) GetName() string {
	return k.Name
}

func (k *K8sMock) GetNamespace() string {
	return k.Namespace
}

func (k *K8sMock) GetUID() types.UID {
	return k.UID
}
