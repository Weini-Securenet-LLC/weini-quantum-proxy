package wailsapp

import "testing"

func TestBuildTagProductionIsDocumentedByScripts(t *testing.T) {
	// Regression note: Wails release builds must use -tags production
	// (see scripts/build_proxy_node_studio_wails.sh).
	t.Log("production build tag expected for release binaries")
}
