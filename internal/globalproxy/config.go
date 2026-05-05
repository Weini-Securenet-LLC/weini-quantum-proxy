package globalproxy

import (
	"encoding/json"
	"fmt"
	"strings"

	"hermes-agent/proxy-node-studio/internal/proxynode"
)

type Mode string

const (
	ModeSystem Mode = "system"
	ModeTUN    Mode = "tun"
)

type Options struct {
	Mode      Mode
	MixedPort int
	SocksPort int
}

func BuildConfig(node proxynode.Node, opts Options) (string, error) {
	if opts.Mode == "" {
		opts.Mode = ModeTUN
	}
	if opts.MixedPort == 0 {
		opts.MixedPort = 7890
	}
	if opts.SocksPort == 0 {
		opts.SocksPort = 1080
	}
	outbound, err := outboundForNode(node)
	if err != nil {
		return "", err
	}
	config := map[string]any{
		"log": map[string]any{
			"level":     "info",
			"timestamp": true,
		},
		"dns": map[string]any{
			"strategy": "prefer_ipv4",
			"servers": []any{
				map[string]any{"tag": "remote", "type": "udp", "server": "1.1.1.1", "server_port": 53},
				map[string]any{"tag": "direct-dns", "type": "udp", "server": "223.5.5.5", "server_port": 53},
			},
			"final": "remote",
		},
		"inbounds": buildInbounds(opts),
		"outbounds": []any{
			outbound,
			map[string]any{"type": "direct", "tag": "direct"},
			map[string]any{"type": "block", "tag": "block"},
		},
		"route": map[string]any{
			"auto_detect_interface":   true,
			"final":                   "proxy",
			"default_domain_resolver": "remote",
		},
	}
	raw, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func buildInbounds(opts Options) []any {
	inbounds := []any{}
	if opts.Mode == ModeTUN {
		inbounds = append(inbounds, map[string]any{
			"type":           "tun",
			"tag":            "tun-in",
			"interface_name": "ProxyNodeStudio",
			"address": []string{
				"172.19.0.1/30",
				"fdfe:dcba:9876::1/126",
			},
			"auto_route":   true,
			"strict_route": true,
			"stack":        "system",
		})
	}
	inbounds = append(inbounds,
		map[string]any{
			"type":        "mixed",
			"tag":         "mixed-in",
			"listen":      "127.0.0.1",
			"listen_port": opts.MixedPort,
		},
		map[string]any{
			"type":        "socks",
			"tag":         "socks-in",
			"listen":      "127.0.0.1",
			"listen_port": opts.SocksPort,
		},
	)
	return inbounds
}

func outboundForNode(node proxynode.Node) (map[string]any, error) {
	details := nodeDetails(node)
	switch strings.ToLower(node.Protocol) {
	case "trojan":
		outbound := map[string]any{
			"type":        "trojan",
			"tag":         "proxy",
			"server":      node.Host,
			"server_port": node.Port,
			"password":    node.Credential,
		}
		applyTLSAndTransport(outbound, node, details)
		return outbound, nil
	case "ss":
		return map[string]any{
			"type":        "shadowsocks",
			"tag":         "proxy",
			"server":      node.Host,
			"server_port": node.Port,
			"method":      fallback(node.Method, "aes-128-gcm"),
			"password":    node.Credential,
		}, nil
	case "vmess":
		outbound := map[string]any{
			"type":        "vmess",
			"tag":         "proxy",
			"server":      node.Host,
			"server_port": node.Port,
			"uuid":        node.Credential,
			"security":    fallback(node.Method, "auto"),
		}
		if aid := firstString(details, "aid", "alter_id", "alterId"); aid != "" && aid != "0" {
			outbound["alter_id"] = aid
		}
		applyTLSAndTransport(outbound, node, details)
		return outbound, nil
	case "vless":
		outbound := map[string]any{
			"type":        "vless",
			"tag":         "proxy",
			"server":      node.Host,
			"server_port": node.Port,
			"uuid":        node.Credential,
		}
		if flow := firstString(details, "flow"); flow != "" {
			outbound["flow"] = flow
		}
		applyTLSAndTransport(outbound, node, details)
		return outbound, nil
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", node.Protocol)
	}
}

func applyTLSAndTransport(outbound map[string]any, node proxynode.Node, details map[string]string) {
	if tlsCfg := tlsConfig(node, details); len(tlsCfg) > 0 {
		outbound["tls"] = tlsCfg
	}
	if transport := transportConfig(node, details); len(transport) > 0 {
		outbound["transport"] = transport
	}
}

func tlsConfig(node proxynode.Node, details map[string]string) map[string]any {
	enabled := strings.EqualFold(node.TLS, "tls") || strings.EqualFold(details["security"], "tls") || strings.EqualFold(details["tls"], "tls")
	if !enabled {
		return nil
	}
	cfg := map[string]any{"enabled": true}
	if serverName := firstString(details, "sni", "serverName", "servername"); serverName != "" {
		cfg["server_name"] = serverName
	}
	if insecure := firstString(details, "allowInsecure", "skip-cert-verify", "insecure"); strings.EqualFold(insecure, "true") {
		cfg["insecure"] = true
	}
	if alpn := firstString(details, "alpn"); alpn != "" {
		cfg["alpn"] = splitCSV(alpn)
	}
	return cfg
}

func transportConfig(node proxynode.Node, details map[string]string) map[string]any {
	network := strings.ToLower(firstString(details, "type", "net", "network"))
	if network == "" {
		network = strings.ToLower(node.Network)
	}
	switch network {
	case "", "tcp":
		return nil
	case "ws":
		cfg := map[string]any{"type": "ws"}
		if path := firstString(details, "path"); path != "" {
			cfg["path"] = path
		}
		if host := firstString(details, "host", "host_header", "Host"); host != "" {
			cfg["headers"] = map[string]any{"Host": host}
		}
		return cfg
	case "grpc":
		cfg := map[string]any{"type": "grpc"}
		if service := firstString(details, "serviceName", "service_name"); service != "" {
			cfg["service_name"] = service
		}
		return cfg
	case "httpupgrade":
		cfg := map[string]any{"type": "httpupgrade"}
		if host := firstString(details, "host", "host_header", "Host"); host != "" {
			cfg["host"] = host
		}
		if path := firstString(details, "path"); path != "" {
			cfg["path"] = path
		}
		return cfg
	default:
		return map[string]any{"type": network}
	}
}

func nodeDetails(node proxynode.Node) map[string]string {
	if strings.TrimSpace(node.SourceDetail) == "" {
		return map[string]string{}
	}
	var out map[string]string
	if err := json.Unmarshal([]byte(node.SourceDetail), &out); err == nil {
		return out
	}
	var generic map[string]any
	if err := json.Unmarshal([]byte(node.SourceDetail), &generic); err != nil {
		return map[string]string{}
	}
	out = make(map[string]string, len(generic))
	for key, value := range generic {
		out[key] = fmt.Sprint(value)
	}
	return out
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func firstString(m map[string]string, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(m[key]); value != "" {
			return value
		}
	}
	return ""
}

func fallback(value string, defaultValue string) string {
	if strings.TrimSpace(value) == "" {
		return defaultValue
	}
	return value
}
