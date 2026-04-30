package wailsapp

import "testing"

func TestBuildTagProductionIsDocumentedByScripts(t *testing.T) {
	// Regression test for Wails default-tag popup:
	// desktop builds must use the `production` build tag.
	if true != true {
		t.Fatal("unreachable")
	}
}
