package weaveplugin

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

var formatStrings = map[string]string{
	"host":       "/console/platform/nodes/%v",                   // ip
	"deployment": "/console/workloads/deployments/%s/%s",         // namespace, deployment
	"pod":        "/console/workloads/deployments/%s/%s/pods/%s", // namespaec, deployment, pod
}

type PluginSpec struct {
	ID          string   `json:"id"`
	Label       string   `json:"label"`
	Description string   `json:"description,omitempty"`
	Interfaces  []string `json:"interfaces"`
	APIVersion  int      `json:"api_version,omitempty"`
}

type LatestSample struct {
	Value     string    `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

type Node struct {
	Latest map[string]LatestSample `json:"latest"`
}

type Metadata struct {
	ID       string  `json:"id"`
	Label    string  `json:"label"`
	DataType string  `json:"dataType"`
	Priority float64 `json:"priority"`
	From     string  `json:"from"`
}

type TableColumn struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	DataType string `json:"dataType"`
}

type Table struct {
	ID      string        `json:"id"`
	Label   string        `json:"label"`
	Prefix  string        `json:"prefix"`
	Type    string        `json:"type"`
	Columns []TableColumn `json:"columns,omitempty"`
}

type Topology struct {
	Nodes             map[string]Node     `json:"nodes"`
	MetadataTemplates map[string]Metadata `json:"metadata_templates,omitempty"`
	TableTemplates    map[string]Table    `json:"table_templates,omitempty"`
}

type WeaveReport struct {
	Host Topology `json:"Host,omitempty"`

	Pods Topology `json:"Pods,omitempty"`

	Deployment Topology `json:"Deployment,omitempty"`

	Plugins []PluginSpec `json:"Plugins"`
}

type WeaveT struct {
	ID      string
	Type    string
	TableID string
}

func (w *WeaveT) GetID() string {
	return w.ID
}

func (w *WeaveT) GetType() string {
	return w.Type
}

func (w *WeaveT) GetTableID() string {
	return w.TableID
}

type Host struct {
	WeaveT
	IP string
}

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

	h.ID = hostID
	h.Type = "host"
	h.IP = GetOutboundIP()

	h.TableID = fmt.Sprintf("table-%s")
}

func (h *Host) GetPlatformUrl() (string, error) {
	if urlFormat, ok := formatStrings[h.Type]; ok {
		return fmt.Sprintf(urlFormat, h.IP), nil
	}

	return "", fmt.Errorf("URL Format String Not Found")
}

type K8SObject interface {
	GetID() string
	GetPlatformUrl() (string, error)
	GetTableID() string
	GetType() string
	Init()
}

func GetWeaveID(obj K8SObject) string {
	return fmt.Sprintf("%s;<%s>", obj.GetID(), obj.GetType())
}

func GetMetaDataTableID(obj K8SObject) string {
	return fmt.Sprintf("%s-1___%s-column-1", obj.GetTableID(), obj.GetTableID())
}

func GetWeaveTable(obj K8SObject) Table {
	column := TableColumn{
		ID:       fmt.Sprintf("%s-column-1", obj.GetTableID()),
		Label:    fmt.Sprintf("%s ICP Link", obj.GetType()),
		DataType: "",
	}

	return Table{
		ID:      obj.GetTableID(),
		Label:   "",
		Prefix:  fmt.Sprintf("%s-", obj.GetTableID()),
		Type:    "multicolumn-table",
		Columns: []TableColumn{column},
	}
}

func GetWeaveMetaData(obj K8SObject) Metadata {
	return Metadata{
		ID:       fmt.Sprintf("%s-meta", obj.GetTableID()),
		Label:    "Multi-Column Links",
		DataType: "",
		Priority: 0.1,
		From:     "latest",
	}
}

func GetLatestURL(obj K8SObject) map[string]LatestSample {
	url, _ := obj.GetPlatformUrl()

	return map[string]LatestSample{
		GetMetaDataTableID(obj): {Value: url},
	}
}
