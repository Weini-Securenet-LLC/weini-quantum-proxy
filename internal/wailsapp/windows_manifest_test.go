package wailsapp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWindowsManifestRequiresAdministrator(t *testing.T) {
	manifestPath := filepath.Join("..", "..", "cmd", "proxy-node-studio-wails", "app.windows.manifest")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	manifest := string(data)
	if !strings.Contains(manifest, `requestedExecutionLevel level="requireAdministrator"`) {
		t.Fatalf("manifest does not request administrator privileges: %s", manifest)
	}
}
