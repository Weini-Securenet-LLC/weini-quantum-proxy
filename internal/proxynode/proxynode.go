package proxynode

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

var SupportedProtocols = []string{"ss", "vmess", "vless", "trojan"}

const DefaultURL = "http://sanbuziyou.icu/list.json"

type Node struct {
	Protocol     string `json:"protocol"`
	Name         string `json:"name"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Method       string `json:"method"`
	Credential   string `json:"credential"`
	Network      string `json:"network"`
	TLS          string `json:"tls"`
	SourceDetail string `json:"source_detail"`
	RawURI       string `json:"raw_uri"`
}

type Summary struct {
	TotalNodes     int            `json:"total_nodes"`
	ProtocolCounts map[string]int `json:"protocol_counts"`
	HostsPreview   []string       `json:"hosts_preview"`
}

type Output struct {
	SourceURL          string         `json:"source_url"`
	SupportedProtocols []string       `json:"supported_protocols"`
	TotalNodes         int            `json:"total_nodes"`
	ProtocolCounts     map[string]int `json:"protocol_counts"`
	HostsPreview       []string       `json:"hosts_preview"`
	Nodes              []Node         `json:"nodes"`
}

func ParseProtocols(csv string) []string {
	if strings.TrimSpace(csv) == "" {
		return append([]string(nil), SupportedProtocols...)
	}
	seen := map[string]bool{}
	protocols := make([]string, 0, len(SupportedProtocols))
	for _, part := range strings.Split(csv, ",") {
		p := strings.ToLower(strings.TrimSpace(part))
		if p == "" || seen[p] {
			continue
		}
		seen[p] = true
		protocols = append(protocols, p)
	}
	if len(protocols) == 0 {
		return append([]string(nil), SupportedProtocols...)
	}
	return protocols
}

func FetchAndNormalize(rawURL string, timeoutSeconds float64, protocols []string) (Output, error) {
	payload, err := FetchPayload(rawURL, timeoutSeconds)
	if err != nil {
		return Output{}, err
	}
	nodes := NormalizeNodes(payload, protocols)
	return BuildOutput(rawURL, protocols, nodes), nil
}

func FetchPayload(rawURL string, timeoutSeconds float64) (any, error) {
	client := &http.Client{Timeout: secondsToDuration(timeoutSeconds)}
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "proxy-node-studio-go/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var payload any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func NormalizeNodes(payload any, enabledProtocols []string) []Node {
	allowed := map[string]bool{}
	for _, p := range enabledProtocols {
		allowed[strings.ToLower(p)] = true
	}
	uris := discoverURIs(payload)
	seen := map[string]bool{}
	nodes := make([]Node, 0, len(uris))
	for _, uri := range uris {
		node, ok := parseNodeURI(uri)
		if !ok || !allowed[node.Protocol] {
			continue
		}
		key := fmt.Sprintf("%s|%s|%d|%s", node.Protocol, node.Host, node.Port, node.Credential)
		if seen[key] {
			continue
		}
		seen[key] = true
		nodes = append(nodes, node)
	}
	return nodes
}

func SummarizeNodes(nodes []Node, protocols []string) Summary {
	counts := map[string]int{}
	for _, p := range protocols {
		counts[p] = 0
	}
	hosts := make([]string, 0, len(nodes))
	for _, node := range nodes {
		counts[node.Protocol]++
		hosts = append(hosts, fmt.Sprintf("%s:%d", node.Host, node.Port))
	}
	sort.Strings(hosts)
	if len(hosts) > 12 {
		hosts = hosts[:12]
	}
	return Summary{TotalNodes: len(nodes), ProtocolCounts: counts, HostsPreview: hosts}
}

func BuildOutput(sourceURL string, protocols []string, nodes []Node) Output {
	summary := SummarizeNodes(nodes, protocols)
	return Output{
		SourceURL:          sourceURL,
		SupportedProtocols: append([]string(nil), protocols...),
		TotalNodes:         summary.TotalNodes,
		ProtocolCounts:     summary.ProtocolCounts,
		HostsPreview:       summary.HostsPreview,
		Nodes:              nodes,
	}
}

func ParseNodeURI(uri string) (Node, bool) {
	return parseNodeURI(uri)
}

func discoverURIs(value any) []string {
	var found []string
	var walk func(any)
	walk = func(v any) {
		switch t := v.(type) {
		case string:
			for _, token := range strings.Fields(t) {
				clean := strings.TrimRight(token, ")]}>,.;")
				if isSupportedURI(clean) {
					found = append(found, clean)
				}
			}
		case []any:
			for _, item := range t {
				walk(item)
			}
		case map[string]any:
			for _, key := range []string{"uri", "url", "link", "raw_uri", "proxy", "node"} {
				if raw, ok := t[key].(string); ok {
					for _, token := range strings.Fields(raw) {
						clean := strings.TrimRight(token, ")]}>,.;")
						if isSupportedURI(clean) {
							found = append(found, clean)
						}
					}
				}
			}
			if uri, ok := convertProxyObjectToURI(t); ok {
				found = append(found, uri)
			}
			for _, item := range t {
				walk(item)
			}
		}
	}
	walk(value)
	return found
}

func isSupportedURI(value string) bool {
	for _, protocol := range SupportedProtocols {
		if strings.HasPrefix(strings.ToLower(value), protocol+"://") {
			return true
		}
	}
	return false
}

func convertProxyObjectToURI(proxy map[string]any) (string, bool) {
	proxyType := strings.ToLower(stringValue(proxy["type"]))
	if proxyType == "" {
		proxyType = strings.ToLower(stringValue(proxy["protocol"]))
	}
	switch proxyType {
	case "ss":
		return nodeToSSURI(proxy)
	case "vmess":
		return nodeToVMessURI(proxy)
	case "vless":
		return nodeToVLessURI(proxy)
	case "trojan":
		return nodeToTrojanURI(proxy)
	default:
		return "", false
	}
}

func parseNodeURI(uri string) (Node, bool) {
	lower := strings.ToLower(uri)
	switch {
	case strings.HasPrefix(lower, "ss://"):
		return parseSSURI(uri)
	case strings.HasPrefix(lower, "vmess://"):
		return parseVMessURI(uri)
	case strings.HasPrefix(lower, "vless://"):
		return parseVLessOrTrojanURI(uri, "vless")
	case strings.HasPrefix(lower, "trojan://"):
		return parseVLessOrTrojanURI(uri, "trojan")
	default:
		return Node{}, false
	}
}

func parseSSURI(uri string) (Node, bool) {
	body := strings.TrimPrefix(uri, "ss://")
	name := ""
	if idx := strings.LastIndex(body, "#"); idx >= 0 {
		name, _ = url.QueryUnescape(body[idx+1:])
		body = body[:idx]
	}
	plugin := ""
	if idx := strings.Index(body, "?"); idx >= 0 {
		plugin = body[idx+1:]
		body = body[:idx]
	}
	var method, credential, hostport string
	if strings.Contains(body, "@") {
		parts := strings.SplitN(body, "@", 2)
		decoded, ok := b64DecodePad(parts[0])
		if !ok || !strings.Contains(decoded, ":") {
			return Node{}, false
		}
		creds := strings.SplitN(decoded, ":", 2)
		method, credential = creds[0], creds[1]
		hostport = parts[1]
	} else {
		decoded, ok := b64DecodePad(body)
		if !ok || !strings.Contains(decoded, "@") || !strings.Contains(decoded, ":") {
			return Node{}, false
		}
		parts := strings.SplitN(decoded, "@", 2)
		creds := strings.SplitN(parts[0], ":", 2)
		method, credential = creds[0], creds[1]
		hostport = parts[1]
	}
	host, port, ok := splitHostPort(hostport)
	if !ok {
		return Node{}, false
	}
	if name == "" {
		name = "unnamed"
	}
	return Node{Protocol: "ss", Name: name, Host: host, Port: port, Method: method, Credential: credential, Network: "tcp", TLS: "unknown", SourceDetail: plugin, RawURI: uri}, true
}

func parseVMessURI(uri string) (Node, bool) {
	payload, ok := b64DecodePad(strings.TrimPrefix(uri, "vmess://"))
	if !ok {
		return Node{}, false
	}
	var data map[string]any
	if err := json.Unmarshal([]byte(payload), &data); err != nil {
		return Node{}, false
	}
	host := stringValue(data["add"])
	port, ok := intValue(data["port"])
	if !ok || host == "" {
		return Node{}, false
	}
	detail, _ := json.Marshal(map[string]any{
		"path":        data["path"],
		"host_header": data["host"],
		"sni":         data["sni"],
		"aid":         data["aid"],
		"alpn":        data["alpn"],
	})
	return Node{Protocol: "vmess", Name: fallback(stringValue(data["ps"]), "unnamed"), Host: host, Port: port, Method: fallback(stringValue(data["scy"]), fallback(stringValue(data["type"]), "auto")), Credential: stringValue(data["id"]), Network: fallback(stringValue(data["net"]), "tcp"), TLS: fallback(stringValue(data["tls"]), fallback(stringValue(data["security"]), "none")), SourceDetail: string(detail), RawURI: uri}, true
}

func parseVLessOrTrojanURI(uri string, protocol string) (Node, bool) {
	parsed, err := url.Parse(uri)
	if err != nil || parsed.Host == "" {
		return Node{}, false
	}
	host := parsed.Hostname()
	port64, err := strconv.ParseInt(parsed.Port(), 10, 64)
	if err != nil || host == "" {
		return Node{}, false
	}
	q := parsed.Query()
	network := fallback(q.Get("type"), fallback(q.Get("net"), "tcp"))
	tlsValue := q.Get("security")
	if tlsValue == "" {
		if protocol == "trojan" {
			tlsValue = "tls"
		} else {
			tlsValue = "none"
		}
	}
	name, _ := url.QueryUnescape(parsed.Fragment)
	if name == "" {
		name = "unnamed"
	}
	detail, _ := json.Marshal(firstQueryValues(q))
	password, _ := url.QueryUnescape(parsed.User.Username())
	return Node{Protocol: protocol, Name: name, Host: host, Port: int(port64), Method: network, Credential: password, Network: network, TLS: tlsValue, SourceDetail: string(detail), RawURI: uri}, true
}

func nodeToSSURI(proxy map[string]any) (string, bool) {
	method := fallback(stringValue(proxy["cipher"]), stringValue(proxy["method"]))
	password := stringValue(proxy["password"])
	server := fallback(stringValue(proxy["server"]), stringValue(proxy["host"]))
	port, ok := intValue(proxy["port"])
	if !ok || method == "" || password == "" || server == "" {
		return "", false
	}
	userinfo := base64.RawURLEncoding.EncodeToString([]byte(method + ":" + password))
	name := url.QueryEscape(fallback(stringValue(proxy["name"]), "unnamed"))
	return fmt.Sprintf("ss://%s@%s:%d#%s", userinfo, server, port, name), true
}

func nodeToVMessURI(proxy map[string]any) (string, bool) {
	server := stringValue(proxy["server"])
	port, ok := intValue(proxy["port"])
	uuid := fallback(stringValue(proxy["uuid"]), stringValue(proxy["id"]))
	if !ok || server == "" || uuid == "" {
		return "", false
	}
	payload := map[string]any{
		"v":    "2",
		"ps":   fallback(stringValue(proxy["name"]), "unnamed"),
		"add":  server,
		"port": strconv.Itoa(port),
		"id":   uuid,
		"aid":  fallback(stringValue(proxy["alterId"]), fallback(stringValue(proxy["alter-id"]), "0")),
		"scy":  fallback(stringValue(proxy["cipher"]), fallback(stringValue(proxy["method"]), "auto")),
		"net":  fallback(stringValue(proxy["network"]), fallback(stringValue(proxy["net"]), "tcp")),
		"type": fallback(stringValue(proxy["type"]), "none"),
		"host": stringValue(proxy["host"]),
		"path": stringValue(proxy["path"]),
		"tls":  boolTLS(proxy),
		"sni":  fallback(stringValue(proxy["servername"]), fallback(stringValue(proxy["serverName"]), stringValue(proxy["sni"]))),
		"alpn": stringValue(proxy["alpn"]),
	}
	raw, _ := json.Marshal(payload)
	return "vmess://" + base64.StdEncoding.EncodeToString(raw), true
}

func nodeToVLessURI(proxy map[string]any) (string, bool) {
	server := stringValue(proxy["server"])
	port, ok := intValue(proxy["port"])
	uuid := fallback(stringValue(proxy["uuid"]), stringValue(proxy["id"]))
	if !ok || server == "" || uuid == "" {
		return "", false
	}
	params := url.Values{}
	params.Set("type", fallback(stringValue(proxy["network"]), fallback(stringValue(proxy["net"]), "tcp")))
	if boolValue(proxy["tls"]) {
		params.Set("security", "tls")
	} else {
		params.Set("security", fallback(stringValue(proxy["security"]), "none"))
	}
	if sni := fallback(stringValue(proxy["servername"]), fallback(stringValue(proxy["serverName"]), stringValue(proxy["sni"]))); sni != "" {
		params.Set("sni", sni)
	}
	if path := stringValue(proxy["path"]); path != "" {
		params.Set("path", path)
	}
	if alpn := stringValue(proxy["alpn"]); alpn != "" {
		params.Set("alpn", alpn)
	}
	name := url.QueryEscape(fallback(stringValue(proxy["name"]), "unnamed"))
	return fmt.Sprintf("vless://%s@%s:%d?%s#%s", uuid, server, port, params.Encode(), name), true
}

func nodeToTrojanURI(proxy map[string]any) (string, bool) {
	server := stringValue(proxy["server"])
	port, ok := intValue(proxy["port"])
	password := fallback(stringValue(proxy["password"]), stringValue(proxy["uuid"]))
	if !ok || server == "" || password == "" {
		return "", false
	}
	params := url.Values{}
	params.Set("type", fallback(stringValue(proxy["network"]), fallback(stringValue(proxy["net"]), "tcp")))
	params.Set("security", "tls")
	if sni := fallback(stringValue(proxy["servername"]), fallback(stringValue(proxy["serverName"]), stringValue(proxy["sni"]))); sni != "" {
		params.Set("sni", sni)
	}
	if path := stringValue(proxy["path"]); path != "" {
		params.Set("path", path)
	}
	if alpn := stringValue(proxy["alpn"]); alpn != "" {
		params.Set("alpn", alpn)
	}
	name := url.QueryEscape(fallback(stringValue(proxy["name"]), "unnamed"))
	return fmt.Sprintf("trojan://%s@%s:%d?%s#%s", password, server, port, params.Encode(), name), true
}

func firstQueryValues(v url.Values) map[string]string {
	out := map[string]string{}
	for key, values := range v {
		if len(values) > 0 {
			out[key] = values[0]
		}
	}
	return out
}

func splitHostPort(hostport string) (string, int, bool) {
	if strings.HasPrefix(hostport, "[") {
		end := strings.Index(hostport, "]")
		if end < 0 || len(hostport) <= end+2 {
			return "", 0, false
		}
		host := hostport[1:end]
		port, err := strconv.Atoi(hostport[end+2:])
		if err != nil || host == "" || port < 1 || port > 65535 {
			return "", 0, false
		}
		return host, port, true
	}
	parts := strings.Split(hostport, ":")
	if len(parts) < 2 {
		return "", 0, false
	}
	port, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil || port < 1 || port > 65535 {
		return "", 0, false
	}
	host := strings.Join(parts[:len(parts)-1], ":")
	if host == "" {
		return "", 0, false
	}
	return host, port, true
}

func b64DecodePad(value string) (string, bool) {
	for _, enc := range []*base64.Encoding{base64.StdEncoding, base64.RawStdEncoding, base64.URLEncoding, base64.RawURLEncoding} {
		if raw, err := enc.DecodeString(value); err == nil {
			return string(raw), true
		}
	}
	if mod := len(value) % 4; mod != 0 {
		value += strings.Repeat("=", 4-mod)
	}
	for _, enc := range []*base64.Encoding{base64.StdEncoding, base64.URLEncoding} {
		if raw, err := enc.DecodeString(value); err == nil {
			return string(raw), true
		}
	}
	return "", false
}

func stringValue(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case int:
		return strconv.Itoa(t)
	case bool:
		if t {
			return "true"
		}
		return "false"
	default:
		return ""
	}
}

func intValue(v any) (int, bool) {
	switch t := v.(type) {
	case int:
		return t, true
	case int64:
		return int(t), true
	case float64:
		return int(t), true
	case json.Number:
		i, err := t.Int64()
		return int(i), err == nil
	case string:
		i, err := strconv.Atoi(t)
		return i, err == nil
	default:
		return 0, false
	}
}

func boolValue(v any) bool {
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return strings.EqualFold(t, "true") || strings.EqualFold(t, "tls")
	default:
		return false
	}
}

func boolTLS(proxy map[string]any) string {
	if boolValue(proxy["tls"]) {
		return "tls"
	}
	return stringValue(proxy["security"])
}

func fallback(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func secondsToDuration(seconds float64) (d time.Duration) {
	return time.Duration(seconds * float64(time.Second))
}
