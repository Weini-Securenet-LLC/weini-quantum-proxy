package proxynode

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

func makeVMessURI(t *testing.T) string {
	t.Helper()
	payload := map[string]any{
		"v":    "2",
		"ps":   "demo-vmess",
		"add":  "vmess.example.com",
		"port": "443",
		"id":   "11111111-1111-1111-1111-111111111111",
		"aid":  "0",
		"scy":  "auto",
		"net":  "ws",
		"type": "none",
		"host": "cdn.example.com",
		"path": "/ws",
		"tls":  "tls",
		"sni":  "sni.example.com",
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	return "vmess://" + base64.StdEncoding.EncodeToString(raw)
}

func samplePayload(t *testing.T) map[string]any {
	t.Helper()
	return map[string]any{
		"nodes": []any{
			"ss://YWVzLTI1Ni1nY206cGFzczFAc3MuZXhhbXBsZS5jb206ODM4OA#ss-demo",
			map[string]any{"uri": makeVMessURI(t)},
			map[string]any{
				"type":    "vless",
				"server":  "vless.example.com",
				"port":    8443,
				"uuid":    "22222222-2222-2222-2222-222222222222",
				"network": "ws",
				"tls":     true,
				"path":    "/vless",
				"sni":     "vless-sni.example.com",
				"name":    "vless-demo",
			},
			map[string]any{
				"type":     "trojan",
				"server":   "trojan.example.com",
				"port":     443,
				"password": "secret",
				"network":  "tcp",
				"name":     "trojan-demo",
			},
		},
	}
}

func TestNormalizeNodesExtractsSupportedProtocols(t *testing.T) {
	nodes := NormalizeNodes(samplePayload(t), SupportedProtocols)
	if len(nodes) != 4 {
		t.Fatalf("expected 4 nodes, got %d", len(nodes))
	}
	got := []string{nodes[0].Protocol, nodes[1].Protocol, nodes[2].Protocol, nodes[3].Protocol}
	want := []string{"ss", "vmess", "vless", "trojan"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("protocol[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestSummarizeNodesCountsEachProtocol(t *testing.T) {
	nodes := NormalizeNodes(samplePayload(t), SupportedProtocols)
	summary := SummarizeNodes(nodes, SupportedProtocols)
	if summary.TotalNodes != 4 {
		t.Fatalf("total nodes = %d, want 4", summary.TotalNodes)
	}
	if summary.ProtocolCounts["ss"] != 1 || summary.ProtocolCounts["vmess"] != 1 || summary.ProtocolCounts["vless"] != 1 || summary.ProtocolCounts["trojan"] != 1 {
		t.Fatalf("unexpected protocol counts: %#v", summary.ProtocolCounts)
	}
}

func TestBuildOutputIncludesPreviewAndSourceURL(t *testing.T) {
	nodes := NormalizeNodes(samplePayload(t), SupportedProtocols)
	output := BuildOutput("http://example.com/list.json", SupportedProtocols, nodes)
	if output.SourceURL != "http://example.com/list.json" {
		t.Fatalf("source url = %q", output.SourceURL)
	}
	if output.TotalNodes != 4 {
		t.Fatalf("total nodes = %d, want 4", output.TotalNodes)
	}
	if len(output.HostsPreview) == 0 {
		t.Fatal("expected hosts preview to be populated")
	}
}
