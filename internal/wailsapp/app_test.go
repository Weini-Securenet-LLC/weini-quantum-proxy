package wailsapp

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"hermes-agent/proxy-node-studio/internal/globalproxy"
	"hermes-agent/proxy-node-studio/internal/proxynode"
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

func TestHealthReturnsDefaults(t *testing.T) {
	app := New()
	app.Startup(context.Background())
	resp := app.Health()
	if resp.DefaultURL == "" {
		t.Fatal("expected default url")
	}
	if len(resp.Protocols) != 4 {
		t.Fatalf("expected 4 protocols, got %d", len(resp.Protocols))
	}
}

func TestFetchNodesUsesBackendParser(t *testing.T) {
	fixture := map[string]any{
		"nodes": []any{
			"ss://YWVzLTI1Ni1nY206cGFzczFAc3MuZXhhbXBsZS5jb206ODM4OA#ss-demo",
			map[string]any{"uri": makeVMessURI(t)},
		},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(fixture)
	}))
	defer server.Close()

	app := New()
	app.Startup(context.Background())
	out, err := app.FetchNodes(FetchRequest{URL: server.URL, Protocols: []string{"ss", "vmess"}, Timeout: 5})
	if err != nil {
		t.Fatalf("FetchNodes error: %v", err)
	}
	if out.TotalNodes != 2 {
		t.Fatalf("expected 2 nodes, got %d", out.TotalNodes)
	}
	if out.ProtocolCounts["ss"] != 1 || out.ProtocolCounts["vmess"] != 1 {
		t.Fatalf("unexpected counts: %#v", out.ProtocolCounts)
	}
}

func TestActivateProxyURIParsesAndStoresSelectedNode(t *testing.T) {
	app := New()
	app.Startup(context.Background())
	uri := "trojan://secret@example.com:443?type=tcp#demo"
	selected, err := app.ActivateProxyURI(uri)
	if err != nil {
		t.Fatalf("ActivateProxyURI error: %v", err)
	}
	if selected.Protocol != "trojan" {
		t.Fatalf("protocol = %q, want trojan", selected.Protocol)
	}
	if selected.RawURI != uri {
		t.Fatalf("raw uri = %q, want %q", selected.RawURI, uri)
	}
	if app.activeNode == nil || app.activeNode.RawURI != uri {
		t.Fatal("expected active node to be stored")
	}
}

func TestActivateProxyURIRejectsUnsupportedScheme(t *testing.T) {
	app := New()
	app.Startup(context.Background())
	if _, err := app.ActivateProxyURI("https://example.com"); err == nil {
		t.Fatal("expected unsupported scheme error")
	}
}

func TestFilterTrafficLinesHumanizesFriendlyOutput(t *testing.T) {
	originalResolver := logDestinationResolver
	logDestinationResolver = func(host string) []string {
		if host == "example.com" {
			return []string{"93.184.216.34"}
		}
		return nil
	}
	defer func() { logDestinationResolver = originalResolver }()

	raw := "+0800 2026-04-29 13:04:51  \x1b[36mINFO\x1b[0m [ \x1b[38;5;142m1224758654\x1b[0m 0ms] outbound/trojan[proxy]: outbound connection to example.com:443"
	formatted := filterTrafficLines(raw)
	if strings.Contains(formatted, "example.com") {
		t.Fatalf("formatted log should replace domain with resolved IP, got %q", formatted)
	}
	for _, snippet := range []string{"2026-04-29 13:04:51", "信息", "出站 TROJAN（proxy）正在连接 93.184.216.34:443"} {
		if !strings.Contains(formatted, snippet) {
			t.Fatalf("formatted log missing %q: %q", snippet, formatted)
		}
	}
}

func TestGetActiveProxyNodeReturnsStoredNode(t *testing.T) {
	app := New()
	app.Startup(context.Background())
	uri := "ss://YWVzLTI1Ni1nY206cGFzczFAc3MuZXhhbXBsZS5jb206ODM4OA#ss-demo"
	selected, err := app.ActivateProxyURI(uri)
	if err != nil {
		t.Fatalf("ActivateProxyURI error: %v", err)
	}
	current, ok := app.GetActiveProxyNode()
	if !ok {
		t.Fatal("expected active node")
	}
	if current.Protocol != selected.Protocol || current.RawURI != selected.RawURI {
		t.Fatal("active node mismatch")
	}
}

type fakeProxyRuntime struct {
	startNode      proxynode.Node
	startOpts      globalproxy.Options
	status         globalproxy.Status
	startErr       error
	stopErr        error
	stopCalls      int
	quickStopCalls int
}

type fakeNodeProber struct {
	results map[string]NodeTestResult
	calls   []string
}

func (f *fakeProxyRuntime) Start(node proxynode.Node, opts globalproxy.Options) (globalproxy.Status, error) {
	f.startNode = node
	f.startOpts = opts
	if f.startErr != nil {
		return globalproxy.Status{}, f.startErr
	}
	if !f.status.Connected {
		f.status = globalproxy.Status{Connected: true, Mode: string(opts.Mode), Node: &node, PID: 2468}
	}
	return f.status, nil
}

func (f *fakeProxyRuntime) Stop() (globalproxy.Status, error) {
	f.stopCalls++
	if f.stopErr != nil {
		return globalproxy.Status{}, f.stopErr
	}
	f.status.Connected = false
	return f.status, nil
}

func (f *fakeProxyRuntime) StopQuick() (globalproxy.Status, error) {
	f.quickStopCalls++
	if f.stopErr != nil {
		return globalproxy.Status{}, f.stopErr
	}
	f.status.Connected = false
	f.status.CheckMessage = "代理已停止"
	return f.status, nil
}

func (f *fakeProxyRuntime) Status() globalproxy.Status {
	return f.status
}

func (f *fakeNodeProber) Probe(node proxynode.Node) NodeTestResult {
	f.calls = append(f.calls, node.RawURI)
	if result, ok := f.results[node.RawURI]; ok {
		return result
	}
	return NodeTestResult{Node: node, Usable: false, Error: "missing fake result"}
}

func TestConnectGlobalProxyUsesActiveNode(t *testing.T) {
	app := New()
	app.proxyRuntime = &fakeProxyRuntime{}
	app.Startup(context.Background())
	uri := "trojan://secret@example.com:443?type=tcp#demo"
	selected, err := app.ActivateProxyURI(uri)
	if err != nil {
		t.Fatalf("ActivateProxyURI error: %v", err)
	}
	status, err := app.ConnectGlobalProxy(ProxyConnectRequest{Mode: "tun"})
	if err != nil {
		t.Fatalf("ConnectGlobalProxy error: %v", err)
	}
	if !status.Connected {
		t.Fatal("expected connected status")
	}
	if status.Node == nil || status.Node.RawURI != selected.RawURI {
		t.Fatal("expected connected node in status")
	}
}

func TestConnectGlobalProxyRequiresSelection(t *testing.T) {
	app := New()
	app.proxyRuntime = &fakeProxyRuntime{}
	app.Startup(context.Background())
	if _, err := app.ConnectGlobalProxy(ProxyConnectRequest{Mode: "tun"}); err == nil {
		t.Fatal("expected error without active node")
	}
}

func TestDisconnectGlobalProxyReturnsDisconnectedStatus(t *testing.T) {
	fake := &fakeProxyRuntime{status: globalproxy.Status{Connected: true, PID: 2468}}
	app := New()
	app.proxyRuntime = fake
	app.Startup(context.Background())
	status, err := app.DisconnectGlobalProxy()
	if err != nil {
		t.Fatalf("DisconnectGlobalProxy error: %v", err)
	}
	if status.Connected {
		t.Fatal("expected disconnected status")
	}
}

func TestShutdownStopsProxyRuntime(t *testing.T) {
	fake := &fakeProxyRuntime{status: globalproxy.Status{Connected: true, PID: 1357}}
	app := New()
	app.proxyRuntime = fake
	app.Startup(context.Background())
	app.Shutdown(context.Background())
	if fake.quickStopCalls != 1 {
		t.Fatalf("quick stop calls = %d, want 1", fake.quickStopCalls)
	}
	if fake.stopCalls != 0 {
		t.Fatalf("stop calls = %d, want 0", fake.stopCalls)
	}
	if fake.status.Connected {
		t.Fatal("expected runtime disconnected after shutdown")
	}
}

func TestBeforeCloseStopsProxyRuntimeAndAllowsClose(t *testing.T) {
	fake := &fakeProxyRuntime{status: globalproxy.Status{Connected: true, PID: 2469}}
	app := New()
	app.proxyRuntime = fake
	app.Startup(context.Background())
	prevent := app.BeforeClose(context.Background())
	if prevent {
		t.Fatal("expected close to continue")
	}
	if fake.quickStopCalls != 1 {
		t.Fatalf("quick stop calls = %d, want 1", fake.quickStopCalls)
	}
	if fake.stopCalls != 0 {
		t.Fatalf("stop calls = %d, want 0", fake.stopCalls)
	}
}

func TestTestProxyNodesReturnsProbeResults(t *testing.T) {
	ssURI := "ss://YWVzLTI1Ni1nY206cGFzczFAc3MuZXhhbXBsZS5jb206ODM4OA#ss-demo"
	trojanURI := "trojan://secret@example.com:443?type=tcp#demo"
	app := New()
	app.nodeProber = &fakeNodeProber{results: map[string]NodeTestResult{
		ssURI:     {Node: proxynode.Node{Protocol: "ss", RawURI: ssURI, Host: "ss.example.com", Port: 8388}, Usable: true, LatencyMS: 120},
		trojanURI: {Node: proxynode.Node{Protocol: "trojan", RawURI: trojanURI, Host: "example.com", Port: 443}, Usable: false, Error: "timeout"},
	}}
	app.Startup(context.Background())
	results, err := app.TestProxyNodes(NodeTestRequest{URIs: []string{ssURI, trojanURI}})
	if err != nil {
		t.Fatalf("TestProxyNodes error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if !results[0].Usable || results[0].LatencyMS != 120 {
		t.Fatalf("unexpected first result: %#v", results[0])
	}
	if results[1].Usable || results[1].Error == "" {
		t.Fatalf("unexpected second result: %#v", results[1])
	}
}

func TestAutoSelectFastestNodeChoosesLowestLatencyUsableNode(t *testing.T) {
	ssURI := "ss://YWVzLTI1Ni1nY206cGFzczFAc3MuZXhhbXBsZS5jb206ODM4OA#ss-demo"
	vmessURI := makeVMessURI(t)
	app := New()
	app.nodeProber = &fakeNodeProber{results: map[string]NodeTestResult{
		ssURI:    {Node: proxynode.Node{Protocol: "ss", RawURI: ssURI, Host: "ss.example.com", Port: 8388, Name: "slow"}, Usable: true, LatencyMS: 180},
		vmessURI: {Node: proxynode.Node{Protocol: "vmess", RawURI: vmessURI, Host: "vmess.example.com", Port: 443, Name: "fast"}, Usable: true, LatencyMS: 45},
	}}
	app.Startup(context.Background())
	result, err := app.AutoSelectFastestNode(NodeTestRequest{URIs: []string{ssURI, vmessURI}})
	if err != nil {
		t.Fatalf("AutoSelectFastestNode error: %v", err)
	}
	if result.Node.RawURI != vmessURI {
		t.Fatalf("selected raw uri = %q, want fastest vmess", result.Node.RawURI)
	}
	current, ok := app.GetActiveProxyNode()
	if !ok || current.RawURI != vmessURI {
		t.Fatal("expected fastest node to become active")
	}
}

func TestAutoSelectFastestNodeRejectsWhenAllNodesFail(t *testing.T) {
	uri := "trojan://secret@example.com:443?type=tcp#demo"
	app := New()
	app.nodeProber = &fakeNodeProber{results: map[string]NodeTestResult{
		uri: {Node: proxynode.Node{Protocol: "trojan", RawURI: uri}, Usable: false, Error: "timeout"},
	}}
	app.Startup(context.Background())
	if _, err := app.AutoSelectFastestNode(NodeTestRequest{URIs: []string{uri}}); err == nil {
		t.Fatal("expected error when no usable nodes exist")
	}
}
