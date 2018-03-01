package main

import (
	"bytes"
	"encoding/json"
	"os"
)

// Pretty format an array of bytes containing json
func PrettyFmt(b []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "  ")
	return out.Bytes(), err
}

// Print if the DEBUG variable is set to true
func Debug(do func()) {
	if isDebug, ok := os.LookupEnv("DEBUG"); ok && isDebug == "true" {
		do()
	}
}

// Modified GetEnv that allows for a fallback value if the environment variable
// is not found
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}
