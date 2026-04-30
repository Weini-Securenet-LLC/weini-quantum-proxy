package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMainForcesRuntimeRelaunchElevation(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("main.go"))
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	text := string(data)
	if !strings.Contains(text, "MaybeRelaunchElevated()") {
		t.Fatal("main.go should force runtime self-elevation before Wails startup")
	}
}
