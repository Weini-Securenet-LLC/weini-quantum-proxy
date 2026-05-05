package wailsapp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFrontendRemovesPingAndTTLControls(t *testing.T) {
	base := filepath.Join("..", "..", "cmd", "proxy-node-studio-wails", "web")
	indexData, err := os.ReadFile(filepath.Join(base, "index.html"))
	if err != nil {
		t.Fatalf("read index.html: %v", err)
	}
	appData, err := os.ReadFile(filepath.Join(base, "app.js"))
	if err != nil {
		t.Fatalf("read app.js: %v", err)
	}
	indexText := string(indexData)
	appText := string(appData)

	for _, forbidden := range []string{"刷新 Ping", "TTL / Ping", "Ping 已刷新", "refreshTtlBtn", "refreshNodeTtl", "ttlByUri"} {
		if strings.Contains(indexText, forbidden) || strings.Contains(appText, forbidden) {
			t.Fatalf("unexpected leftover ping/ttl text: %q", forbidden)
		}
	}
}

func TestDisconnectSuccessMessageUsesRequestedToastText(t *testing.T) {
	appPath := filepath.Join("..", "..", "cmd", "proxy-node-studio-wails", "web", "app.js")
	data, err := os.ReadFile(appPath)
	if err != nil {
		t.Fatalf("read app.js: %v", err)
	}
	text := string(data)
	if !strings.Contains(text, "showToast('已成功断开链接', 'success')") {
		t.Fatal("expected disconnect toast text to match requested message")
	}
}
