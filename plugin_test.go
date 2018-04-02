package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"
	"testing"
)

// View the report
func TestGenerateReport(t *testing.T) {
	sampleK8s, err := getSampleK8s()
	Fail(t, err)

	var plugin = &Plugin{
		ID:          "test-plugin",
		Label:       "Test Weave Plugin",
		Description: "Used for testing",
		Interfaces:  []string{"reporter"},
		APIVersion:  1,

		sync: sync.Mutex{},
	}

	plugin.WeaveReportInit()

	for _, obj := range sampleK8s {
		plugin.syncAdd(&obj)

		raw, _ := json.Marshal(&obj)
		jsonIndented, _ := PrettyFmt(raw)
		fmt.Printf("%s:\n%s\n\n", obj.GetName(), jsonIndented)
	}

	plugin.LogReport()
}

func getSampleK8s() ([]K8sMock, error) {
	raw, err := ioutil.ReadFile("./__tests__/sample-data.json")
	if err != nil {
		return nil, fmt.Errorf("Couldn't load sample-data")
	}

	sampleK8s := []K8sMock{}
	err = json.Unmarshal(raw, &sampleK8s)
	if err != nil {
		return nil, fmt.Errorf("Couldn't marshal sample-data")
	}

	return sampleK8s, nil
}

func Fail(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}
