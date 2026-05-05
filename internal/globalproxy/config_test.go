package globalproxy

import (
	"encoding/json"
	"testing"

	"hermes-agent/proxy-node-studio/internal/proxynode"
)

func TestBuildConfigForTrojanTunMode(t *testing.T) {
	node := proxynode.Node{
		Protocol:   "trojan",
		Name:       "demo",
		Host:       "example.com",
		Port:       443,
		Credential: "secret",
		Network:    "tcp",
		TLS:        "tls",
		RawURI:     "trojan://secret@example.com:443?type=tcp#demo",
	}
	cfg, err := BuildConfig(node, Options{Mode: ModeTUN})
	if err != nil {
		t.Fatalf("BuildConfig error: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(cfg), &parsed); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	inbounds := parsed["inbounds"].([]any)
	if len(inbounds) < 2 {
		t.Fatal("expected tun plus local helper inbounds")
	}
	tunInbound := inbounds[0].(map[string]any)
	if tunInbound["address"] == nil {
		t.Fatal("expected tun inbound address field for modern sing-box")
	}
	if tunInbound["inet4_address"] != nil || tunInbound["inet6_address"] != nil {
		t.Fatal("did not expect deprecated tun address fields")
	}
	outbounds := parsed["outbounds"].([]any)
	if len(outbounds) < 2 {
		t.Fatal("expected primary outbound and direct outbound")
	}
}

func TestBuildConfigForVMessIncludesTLSAndWebSocketTransport(t *testing.T) {
	node := proxynode.Node{
		Protocol:     "vmess",
		Name:         "demo-vmess",
		Host:         "vmess.example.com",
		Port:         443,
		Method:       "auto",
		Credential:   "11111111-1111-1111-1111-111111111111",
		Network:      "ws",
		TLS:          "tls",
		SourceDetail: `{"path":"/ws","host_header":"cdn.example.com","sni":"edge.example.com","alpn":"h2"}`,
	}
	cfg, err := BuildConfig(node, Options{Mode: ModeTUN})
	if err != nil {
		t.Fatalf("BuildConfig error: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(cfg), &parsed); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	outbounds := parsed["outbounds"].([]any)
	proxy := outbounds[0].(map[string]any)
	tlsCfg := proxy["tls"].(map[string]any)
	if enabled, _ := tlsCfg["enabled"].(bool); !enabled {
		t.Fatal("expected tls enabled")
	}
	transport := proxy["transport"].(map[string]any)
	if transport["type"] != "ws" {
		t.Fatalf("transport type = %#v, want ws", transport["type"])
	}
}

func TestBuildConfigForVLESSPreservesSNIAndPath(t *testing.T) {
	node := proxynode.Node{
		Protocol:     "vless",
		Name:         "demo-vless",
		Host:         "vless.example.com",
		Port:         443,
		Credential:   "22222222-2222-2222-2222-222222222222",
		Network:      "ws",
		TLS:          "tls",
		SourceDetail: `{"type":"ws","security":"tls","sni":"vless-sni.example.com","path":"/socket","host":"cdn.vless.example.com"}`,
	}
	cfg, err := BuildConfig(node, Options{Mode: ModeTUN})
	if err != nil {
		t.Fatalf("BuildConfig error: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(cfg), &parsed); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	outbounds := parsed["outbounds"].([]any)
	proxy := outbounds[0].(map[string]any)
	tlsCfg := proxy["tls"].(map[string]any)
	if tlsCfg["server_name"] != "vless-sni.example.com" {
		t.Fatalf("server_name = %#v", tlsCfg["server_name"])
	}
	transport := proxy["transport"].(map[string]any)
	if transport["path"] != "/socket" {
		t.Fatalf("transport path = %#v", transport["path"])
	}
}

func TestBuildConfigIncludesDNSForTunMode(t *testing.T) {
	node := proxynode.Node{
		Protocol:   "trojan",
		Name:       "demo",
		Host:       "example.com",
		Port:       443,
		Credential: "secret",
		Network:    "tcp",
		TLS:        "tls",
	}
	cfg, err := BuildConfig(node, Options{Mode: ModeTUN})
	if err != nil {
		t.Fatalf("BuildConfig error: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(cfg), &parsed); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	dnsBlock, exists := parsed["dns"]
	if !exists {
		t.Fatal("expected config with dns block for tun mode")
	}
	dnsMap, ok := dnsBlock.(map[string]any)
	if !ok {
		t.Fatalf("dns block type = %T", dnsBlock)
	}
	servers, ok := dnsMap["servers"].([]any)
	if !ok || len(servers) == 0 {
		t.Fatal("expected dns servers in tun config")
	}
	remote := servers[0].(map[string]any)
	if remote["type"] != "udp" || remote["server"] != "1.1.1.1" || remote["server_port"] != float64(53) {
		t.Fatalf("unexpected remote dns server config: %#v", remote)
	}
	if _, exists := remote["detour"]; exists {
		t.Fatalf("dns server should not use detour in sing-box 1.13 runtime config: %#v", remote)
	}
	routeBlock, ok := parsed["route"].(map[string]any)
	if !ok {
		t.Fatalf("route block type = %T", parsed["route"])
	}
	if routeBlock["default_domain_resolver"] != "remote" {
		t.Fatalf("default_domain_resolver = %#v, want remote", routeBlock["default_domain_resolver"])
	}
}

func TestBuildConfigRejectsUnsupportedProtocol(t *testing.T) {
	_, err := BuildConfig(proxynode.Node{Protocol: "http"}, Options{Mode: ModeTUN})
	if err == nil {
		t.Fatal("expected unsupported protocol error")
	}
}
