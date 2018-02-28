package main

import (
	"fmt"
	"os"

	app_v1 "k8s.io/api/apps/v1"
	core_v1 "k8s.io/api/core/v1"
	types "k8s.io/apimachinery/pkg/types"
)

var baseUrl = os.Getenv("ICP_DASHBOARD")

var formatStrings = map[string]string{
	"host":        baseUrl + "/platform/nodes/%v",            // ip
	"deployment":  baseUrl + "/workloads/deployments/%s/%s",  // namespace, deployment
	"daemonset":   baseUrl + "/workloads/daemonsets/%s/%s",   // namespace, daemonset
	"service":     baseUrl + "/access/services/%s/%s",        // namespace, daemonset
	"statefulset": baseUrl + "/workloads/statefulsets/%s/%s", // namespace, daemonset
	// "pod":        "/console/workloads/deployments/%s/%s/pods/%s", // namespace, deployment, pod
	"pod": baseUrl + "/workloads/deployments/%s/%s/pods", // namespace, deployment, pod
}

type K8SObject interface {
	GetUID() types.UID
	GetName() string
	GetNamespace() string
	GetLabels() map[string]string
	GetAnnotations() map[string]string
}

func GetPlatformUrl(obj K8SObject) (string, error) {
	switch obj.(type) {
	case *app_v1.Deployment:
		return fmt.Sprintf(formatStrings["deployment"], obj.GetNamespace(), obj.GetName()), nil

	case *app_v1.DaemonSet:
		return fmt.Sprintf(formatStrings["daemonset"], obj.GetNamespace(), obj.GetName()), nil

	case *core_v1.Service:
		return fmt.Sprintf(formatStrings["service"], obj.GetNamespace(), obj.GetName()), nil

	case *app_v1.StatefulSet:
		return fmt.Sprintf(formatStrings["statefulset"], obj.GetNamespace(), obj.GetName()), nil

	// TODO: Fix this link
	case *core_v1.Pod:
		return fmt.Sprintf(formatStrings["pod"], obj.GetNamespace(), obj.GetName()), nil

	default:
		return "", fmt.Errorf("No Compatible Type Found")
	}
}

func GetTableID(obj K8SObject) string { return fmt.Sprintf("%s", obj.GetName()) }

func GetWeaveID(obj K8SObject) (string, error) {
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

func GetWeaveTable(obj K8SObject) (id string, table Table) {
	id = "icp-link"

	table = Table{
		ID:     "icp-link-",
		Label:  "",
		Prefix: "icp-link-",
		Type:   "multicolumn-table",
		Columns: []TableColumn{{
			ID:       "link",
			Label:    "ICP Link",
			DataType: "link",
		}},
	}

	return
}

func GetWeaveMetaData(obj K8SObject) (id string, metadata Metadata) {
	id = fmt.Sprintf("%s-meta", GetTableID(obj))

	metadata = Metadata{
		ID:       id,
		Label:    "Multi-Column Links",
		DataType: "link",
		Priority: 0.1,
		From:     "latest",
	}

	return
}

func GetLatest(obj K8SObject) (id string, latest LatestSample) {
	url, _ := GetPlatformUrl(obj)

	// Meta data table ID
	id = fmt.Sprintf("icp-link-%s___link", GetTableID(obj))
	latest = LatestSample{Value: url}

	return
}
