package main

import (
	"fmt"

	app_v1 "k8s.io/api/apps/v1"
	core_v1 "k8s.io/api/core/v1"
	types "k8s.io/apimachinery/pkg/types"
)

const PREFIX string = "icp-link-"

var BASE_URL string = GetEnv("ICP_DASHBOARD", "/console")

var formatStrings = map[string]string{
	"host":        BASE_URL + "/platform/nodes/%v",            // ip
	"deployment":  BASE_URL + "/workloads/deployments/%s/%s",  // namespace, deployment
	"daemonset":   BASE_URL + "/workloads/daemonsets/%s/%s",   // namespace, daemonset
	"service":     BASE_URL + "/access/services/%s/%s",        // namespace, daemonset
	"statefulset": BASE_URL + "/workloads/statefulsets/%s/%s", // namespace, daemonset
	// "pod":        "/console/workloads/deployments/%s/%s/pods/%s", // namespace, deployment, pod
	"pod": BASE_URL + "/workloads/deployments/%s/%s/pods", // namespace, deployment, pod
}

type K8sObject interface {
	GetAnnotations() map[string]string
	GetLabels() map[string]string
	GetName() string
	GetNamespace() string
	GetUID() types.UID
}

func GetPlatformUrl(obj K8sObject) (string, error) {
	switch obj.(type) {
	case *app_v1.DaemonSet:
		return fmt.Sprintf(formatStrings["daemonset"], obj.GetNamespace(), obj.GetName()), nil

	case *app_v1.Deployment:
		return fmt.Sprintf(formatStrings["deployment"], obj.GetNamespace(), obj.GetName()), nil

	// TODO: Fix this link
	case *core_v1.Pod:
		return fmt.Sprintf(formatStrings["pod"], obj.GetNamespace(), obj.GetName()), nil

	case *core_v1.Service:
		return fmt.Sprintf(formatStrings["service"], obj.GetNamespace(), obj.GetName()), nil

	case *app_v1.StatefulSet:
		return fmt.Sprintf(formatStrings["statefulset"], obj.GetNamespace(), obj.GetName()), nil

	default:
		return "", fmt.Errorf("No Compatible Type Found")
	}
}

func GetWeaveID(obj K8sObject) (string, error) {
	switch obj.(type) {
	case *app_v1.DaemonSet:
		return fmt.Sprintf("%s;<daemonset>", obj.GetUID()), nil

	case *app_v1.Deployment:
		return fmt.Sprintf("%s;<deployment>", obj.GetUID()), nil

	case *core_v1.Pod:
		return fmt.Sprintf("%s;<pod>", obj.GetUID()), nil

	case *core_v1.Service:
		return fmt.Sprintf("%s;<service>", obj.GetUID()), nil

	case *app_v1.StatefulSet:
		return fmt.Sprintf("%s;<statefulset>", obj.GetUID()), nil

	default:
		return "", fmt.Errorf("No Compatible Type Found")
	}
}

func GetWeaveTable(obj K8sObject) (id string, table Table) {
	id = "icp-link"

	table = Table{
		ID:     "icp-link-",
		Label:  "Cloud Private Links",
		Prefix: PREFIX,
		Type:   "multicolumn-table",
		Columns: []TableColumn{{
			ID:       "link",
			Label:    "",
			DataType: "link",
		}},
	}

	return
}

func GetMetaTemplate() (id string, metadata Metadata) {
	id = fmt.Sprintf("%s-meta", PREFIX)

	metadata = Metadata{
		ID:       id,
		Label:    "ICP Dashboard",
		DataType: "link",
		Priority: 1.1,
		From:     "latest",
	}

	return
}

func GetMetaLatest(obj K8sObject) (id string, latest LatestSample) {
	var url string
	var ok bool

	switch obj.(type) {
	case *core_v1.Service:
		annotations := obj.GetAnnotations()

		if url, ok = annotations["adminConsoleUrl"]; ok {
			url = annotations["adminConsoleUrl"]
		} else {
			url, _ = GetPlatformUrl(obj)
		}

	default:
		url, _ = GetPlatformUrl(obj)
	}

	// Meta data table ID
	id = fmt.Sprintf("%s-meta", PREFIX)
	latest = LatestSample{Value: url}

	return
}

func GetLatest(obj K8sObject) (id string, latest LatestSample) {
	url, _ := GetPlatformUrl(obj)

	// Multi-column table id
	id = fmt.Sprintf("%s%s___link", PREFIX, obj.GetName())
	latest = LatestSample{Value: url}

	return
}
