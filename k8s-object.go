package main

import (
	"fmt"
	"net"
	"os"
	"strings"

	v1 "k8s.io/api/apps/v1"
)

var formatStrings = map[string]string{
	"host":       "/console/platform/nodes/%v",                   // ip
	"deployment": "/console/workloads/deployments/%s/%s",         // namespace, deployment
	"pod":        "/console/workloads/deployments/%s/%s/pods/%s", // namespace, deployment, pod
}

type K8SObject interface {
	GetName() string
	GetNamespace() string
}

type Host struct {
	Name string
	IP   string
}

func (h *Host) GetName() string      { return h.Name }
func (h *Host) GetNamespace() string { return h.IP }

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

func GetPlatformUrl(obj K8SObject) (string, error) {
	switch obj.(type) {
	case *Host:
		return fmt.Sprintf(formatStrings["host"], obj.GetNamespace()), nil
	case *v1.Deployment:
		return fmt.Sprintf(formatStrings["deployment"], obj.GetNamespace(), obj.GetName()), nil
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
	case *v1.Deployment:
		return fmt.Sprintf("%s;<deployment>", obj.GetName()), nil
	}

	return "", fmt.Errorf("No Compatible Type Found")
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
