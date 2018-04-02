package main

import (
	"fmt"

	app_v1 "k8s.io/api/apps/v1"
	core_v1 "k8s.io/api/core/v1"
)

const PREFIX string = "icp-link-"

var BASE_URL string = GetEnv("ICP_DASHBOARD", "/console")

var formatStrings = map[string]string{
	"host":        BASE_URL + "/platform/nodes/%v",            // ip
	"deployment":  BASE_URL + "/workloads/deployments/%s/%s",  // namespace, deployment
	"daemonset":   BASE_URL + "/workloads/daemonsets/%s/%s",   // namespace, daemonset
	"service":     BASE_URL + "/access/services/%s/%s",        // namespace, daemonset
	"statefulset": BASE_URL + "/workloads/statefulsets/%s/%s", // namespace, daemonset
}

type K8sNamer interface {
	GetName() string
	GetNamespace() string
}

func GetPlatformUrl(obj K8sNamer) (string, error) {
	switch obj.(type) {
	case *app_v1.DaemonSet:
		return fmt.Sprintf(formatStrings["daemonset"], obj.GetNamespace(), obj.GetName()), nil

	case *app_v1.Deployment, *K8sMock:
		return fmt.Sprintf(formatStrings["deployment"], obj.GetNamespace(), obj.GetName()), nil

	case *core_v1.Service:
		return fmt.Sprintf(formatStrings["service"], obj.GetNamespace(), obj.GetName()), nil

	case *app_v1.StatefulSet:
		return fmt.Sprintf(formatStrings["statefulset"], obj.GetNamespace(), obj.GetName()), nil

	default:
		return "", fmt.Errorf("No Compatible Type Found")
	}
}

func GetMetaTemplate() (string, Metadata) {
	id := fmt.Sprintf("%s-meta", PREFIX)

	metadata := Metadata{
		ID:       id,
		Label:    "ICP Dashboard",
		DataType: "link",
		Priority: 1.1,
		From:     "latest",
	}

	return id, metadata
}

func GetMetaLatest(obj K8sNamer) (string, LatestSample) {
	url, _ := GetPlatformUrl(obj)

	// Meta data table ID
	id := fmt.Sprintf("%s-meta", PREFIX)
	latest := LatestSample{Value: url}

	return id, latest
}
