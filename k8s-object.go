package main

import (
	"fmt"
	"net"
	"os"
	"strings"

	app_v1 "k8s.io/api/apps/v1"
	core_v1 "k8s.io/api/core/v1"
	types "k8s.io/apimachinery/pkg/types"
)

var formatStrings = map[string]string{
	"host":       "/console/platform/nodes/%v",           // ip
	"deployment": "/console/workloads/deployments/%s/%s", // namespace, deployment
	"daemon_set": "/console/workloads/daemonsets/%s/%s",  // namespace, daemonset
	// "pod":        "/console/workloads/deployments/%s/%s/pods/%s", // namespace, deployment, pod
	"pod": "/console/workloads/deployments/%s/%s/pods", // namespace, deployment, pod
}

type K8SObject interface {
	GetUID() types.UID
	GetName() string
	GetNamespace() string
	GetLabels() map[string]string
}

type Host struct {
	UID  types.UID
	Name string
	IP   string
}

func (h *Host) GetName() string              { return h.Name }
func (h *Host) GetUID() types.UID            { return h.UID }
func (h *Host) GetNamespace() string         { return h.IP }
func (h *Host) GetLabels() map[string]string { return map[string]string{} }

func GetOutboundIP() string {
	conn, _ := net.Dial("udp", "8.8.8.8:80")
	defer conn.Close()
	localAddr := conn.LocalAddr().String()
	idx := strings.LastIndex(localAddr, ":")
	return localAddr[0:idx]
}

func (h *Host) Init() {
	hostID, _ := os.Hostname()

	fmt.Println("Found host", hostID)

	h.Name = hostID
	h.IP = GetOutboundIP()
}

// TODO: Take BaseUrl as the first param
func GetPlatformUrl(obj K8SObject) (string, error) {
	switch obj.(type) {
	case *Host:
		return fmt.Sprintf(formatStrings["host"], obj.GetNamespace()), nil

	case *app_v1.Deployment:
		return fmt.Sprintf(formatStrings["deployment"], obj.GetNamespace(), obj.GetName()), nil

	case *app_v1.DaemonSet:
		return fmt.Sprintf(formatStrings["daemon_set"], obj.GetNamespace(), obj.GetName()), nil

	// TODO: Fix this link
	case *core_v1.Pod:
		return fmt.Sprintf(formatStrings["pod"], obj.GetNamespace(), obj.GetName()), nil

	default:
		return "", fmt.Errorf("No Compatible Type Found")
	}
}

func GetTableID(obj K8SObject) string { return fmt.Sprintf("table-%s", obj.GetName()) }

func GetWeaveID(obj K8SObject) (string, error) {
	switch obj.(type) {
	case *Host:
		return fmt.Sprintf("%s;<host>", obj.GetName()), nil
	// case Pod:
	// 	return fmt.Sprintf("%s;<pod>", obj.GetName())
	case *app_v1.Deployment:
		return fmt.Sprintf("%s;<deployment>", obj.GetUID()), nil

	case *app_v1.DaemonSet:
		return fmt.Sprintf("%s;<daemonset>", obj.GetUID()), nil

	case *core_v1.Pod:
		return fmt.Sprintf("%s;<pod>", obj.GetUID()), nil

	default:
		return "", fmt.Errorf("No Compatible Type Found")
	}
}

func GetWeaveTable(obj K8SObject) (id string, table Table) {
	id = GetTableID(obj)

	table = Table{
		ID:     id,
		Label:  "",
		Prefix: fmt.Sprintf("%s-", GetTableID(obj)),
		Type:   "multicolumn-table",
		Columns: []TableColumn{{
			ID:       fmt.Sprintf("%s-column-1", GetTableID(obj)),
			Label:    "ICP Link",
			DataType: "",
		}},
	}

	return
}

func GetWeaveMetaData(obj K8SObject) (id string, metadata Metadata) {
	id = fmt.Sprintf("%s-meta", GetTableID(obj))

	metadata = Metadata{
		ID:       id,
		Label:    "Multi-Column Links",
		DataType: "",
		Priority: 0.1,
		From:     "latest",
	}

	return
}

func GetLatest(obj K8SObject) (id string, latest LatestSample) {
	url, _ := GetPlatformUrl(obj)

	// Meta data table ID
	id = fmt.Sprintf("%s-1___%s-column-1", GetTableID(obj), GetTableID(obj))
	latest = LatestSample{Value: url}

	return
}
