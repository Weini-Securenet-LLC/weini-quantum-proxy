package wailsapp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDisconnectSuccessToastIsShownBeforePanelRefresh(t *testing.T) {
	appPath := filepath.Join("..", "..", "cmd", "proxy-node-studio-wails", "web", "app.js")
	data, err := os.ReadFile(appPath)
	if err != nil {
		t.Fatalf("read app.js: %v", err)
	}
	text := string(data)
	start := strings.Index(text, "async function disconnectGlobalProxy()")
	if start < 0 {
		t.Fatal("disconnectGlobalProxy function not found")
	}
	body := text[start:]
	refreshIdx := strings.Index(body, "refreshRuntimePanels()")
	toastIdx := strings.Index(body, "showToast('已成功断开链接', 'success')")
	statusIdx := strings.Index(body, "setStatus('已成功断开链接')")
	if refreshIdx < 0 || toastIdx < 0 || statusIdx < 0 {
		t.Fatalf("missing disconnect flow markers in app.js")
	}
	if toastIdx > refreshIdx || statusIdx > refreshIdx {
		t.Fatal("disconnect success feedback must be shown before panel refresh to avoid hanging on '正在断开代理...'")
	}
}

func TestProtocolChipFilteringIsRenderedClientSide(t *testing.T) {
	appPath := filepath.Join("..", "..", "cmd", "proxy-node-studio-wails", "web", "app.js")
	data, err := os.ReadFile(appPath)
	if err != nil {
		t.Fatalf("read app.js: %v", err)
	}
	text := string(data)
	for _, marker := range []string{"function visibleNodes()", "function protocolPriority", "const nodes = visibleNodes();", "renderNodes();"} {
		if !strings.Contains(text, marker) {
			t.Fatalf("expected client-side protocol filtering marker %q", marker)
		}
	}
}
