package wailsapp

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"hermes-agent/proxy-node-studio/internal/globalproxy"
	"hermes-agent/proxy-node-studio/internal/proxynode"
)

type NodeTestRequest struct {
	URIs []string `json:"uris"`
}

type NodeTestResult struct {
	Node      proxynode.Node `json:"node"`
	Usable    bool           `json:"usable"`
	LatencyMS int64          `json:"latency_ms"`
	Error     string         `json:"error,omitempty"`
}

type NodeProber interface {
	Probe(node proxynode.Node) NodeTestResult
}

type HostPinger interface {
	Ping(ctx context.Context, host string) (time.Duration, error)
}

type App struct {
	ctx          context.Context
	activeNode   *proxynode.Node
	proxyRuntime globalproxy.RuntimeController
	nodeProber   NodeProber
}

type TrafficLogSnapshot struct {
	RawTail     string `json:"raw_tail"`
	TrafficTail string `json:"traffic_tail"`
	LogPath     string `json:"log_path"`
	Connected   bool   `json:"connected"`
}

var (
	ansiEscapePattern       = regexp.MustCompile(`\x1b\[[0-9;]*m`)
	singboxLogPrefixPattern = regexp.MustCompile(`^(?P<tz>[+-]\d{4})\s+(?P<date>\d{4}-\d{2}-\d{2})\s+(?P<clock>\d{2}:\d{2}:\d{2})\s+(?P<level>[A-Z]+)\s+`)
	outboundTargetPattern   = regexp.MustCompile(`(?i)^outbound/([^\[]+)\[([^\]]+)\]:\s*outbound connection to\s+([^\s]+)$`)
	dnsLookupPattern        = regexp.MustCompile(`(?i)^dns[^:]*:\s*exchange\s+([^\s]+)\s+for\s+([^\s]+)`)
	logIPCache              sync.Map
	logDestinationResolver  = defaultLogDestinationResolver
)

type HealthResponse struct {
	DefaultURL string   `json:"default_url"`
	Protocols  []string `json:"protocols"`
	OK         bool     `json:"ok"`
}

type FetchRequest struct {
	URL       string   `json:"url"`
	Protocols []string `json:"protocols"`
	Timeout   float64  `json:"timeout"`
}

type ProxyConnectRequest struct {
	Mode      string `json:"mode"`
	MixedPort int    `json:"mixed_port"`
	SocksPort int    `json:"socks_port"`
}

func New() *App {
	return &App{}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	binaryPath, err := resolveBundledSingBoxBinaryPath()
	if err != nil {
		binaryPath = defaultSingBoxBinaryPath()
	}
	if a.proxyRuntime == nil {
		a.proxyRuntime = globalproxy.NewRuntime(globalproxy.RuntimeOptions{
			WorkDir:    defaultProxyWorkDir(),
			BinaryPath: binaryPath,
		})
	}
	if a.nodeProber == nil {
		a.nodeProber = runtimeNodeProber{
			pinger: commandHostPinger{},
		}
	}
}

func (a *App) Shutdown(ctx context.Context) {
	_ = ctx
	a.stopProxyRuntimeQuick()
}

func (a *App) BeforeClose(ctx context.Context) bool {
	_ = ctx
	a.stopProxyRuntimeQuick()
	return false
}

func (a *App) stopProxyRuntime() {
	if a.proxyRuntime != nil {
		_, _ = a.proxyRuntime.Stop()
	}
}

func (a *App) stopProxyRuntimeQuick() {
	if a.proxyRuntime == nil {
		return
	}
	if fast, ok := a.proxyRuntime.(globalproxy.FastStopper); ok {
		_, _ = fast.StopQuick()
		return
	}
	_, _ = a.proxyRuntime.Stop()
}

func (a *App) Health() HealthResponse {
	return HealthResponse{
		DefaultURL: proxynode.DefaultURL,
		Protocols:  append([]string(nil), proxynode.SupportedProtocols...),
		OK:         true,
	}
}

func (a *App) FetchNodes(req FetchRequest) (proxynode.Output, error) {
	url := req.URL
	if url == "" {
		url = proxynode.DefaultURL
	}
	protocols := req.Protocols
	if len(protocols) == 0 {
		protocols = proxynode.SupportedProtocols
	}
	timeout := req.Timeout
	if timeout <= 0 {
		timeout = 20
	}
	return proxynode.FetchAndNormalize(url, timeout, protocols)
}

func (a *App) ActivateProxyURI(uri string) (proxynode.Node, error) {
	target, err := validateProxyURI(uri)
	if err != nil {
		return proxynode.Node{}, err
	}
	node, ok := proxynode.ParseNodeURI(target)
	if !ok {
		return proxynode.Node{}, fmt.Errorf("解析代理链接失败")
	}
	a.activeNode = &node
	return node, nil
}

func (a *App) GetActiveProxyNode() (proxynode.Node, bool) {
	if a.activeNode == nil {
		return proxynode.Node{}, false
	}
	return *a.activeNode, true
}

func (a *App) TestProxyNodes(req NodeTestRequest) ([]NodeTestResult, error) {
	nodes, err := parseRequestedNodes(req)
	if err != nil {
		return nil, err
	}
	results := make([]NodeTestResult, 0, len(nodes))
	for _, node := range nodes {
		results = append(results, a.nodeProber.Probe(node))
	}
	return results, nil
}

func (a *App) AutoSelectFastestNode(req NodeTestRequest) (NodeTestResult, error) {
	results, err := a.TestProxyNodes(req)
	if err != nil {
		return NodeTestResult{}, err
	}
	usable := make([]NodeTestResult, 0, len(results))
	for _, result := range results {
		if result.Usable {
			usable = append(usable, result)
		}
	}
	if len(usable) == 0 {
		return NodeTestResult{}, fmt.Errorf("没有找到可用节点")
	}
	sort.SliceStable(usable, func(i, j int) bool {
		return usable[i].LatencyMS < usable[j].LatencyMS
	})
	best := usable[0]
	nodeCopy := best.Node
	a.activeNode = &nodeCopy
	return best, nil
}

func (a *App) ConnectGlobalProxy(req ProxyConnectRequest) (globalproxy.Status, error) {
	if a.activeNode == nil {
		return globalproxy.Status{}, fmt.Errorf("请先选择节点")
	}
	if a.proxyRuntime == nil {
		return globalproxy.Status{}, fmt.Errorf("代理运行时尚未初始化")
	}
	mode := globalproxy.Mode(strings.ToLower(strings.TrimSpace(req.Mode)))
	if mode == "" {
		mode = globalproxy.ModeTUN
	}
	if mode != globalproxy.ModeTUN && mode != globalproxy.ModeSystem {
		return globalproxy.Status{}, fmt.Errorf("不支持的代理模式: %s", req.Mode)
	}
	return a.proxyRuntime.Start(*a.activeNode, globalproxy.Options{
		Mode:      mode,
		MixedPort: req.MixedPort,
		SocksPort: req.SocksPort,
	})
}

func (a *App) DisconnectGlobalProxy() (globalproxy.Status, error) {
	if a.proxyRuntime == nil {
		return globalproxy.Status{}, nil
	}
	return a.proxyRuntime.Stop()
}

func (a *App) GetGlobalProxyStatus() globalproxy.Status {
	if a.proxyRuntime == nil {
		return globalproxy.Status{}
	}
	return a.proxyRuntime.Status()
}

func (a *App) GetTrafficLogSnapshot() TrafficLogSnapshot {
	status := a.GetGlobalProxyStatus()
	if strings.TrimSpace(status.LogPath) == "" {
		return TrafficLogSnapshot{Connected: status.Connected}
	}
	rawTail := readLogTail(status.LogPath, 8192)
	return TrafficLogSnapshot{
		RawTail:     rawTail,
		TrafficTail: filterTrafficLines(rawTail),
		LogPath:     status.LogPath,
		Connected:   status.Connected,
	}
}

func validateProxyURI(uri string) (string, error) {
	trimmed := strings.TrimSpace(uri)
	if trimmed == "" {
		return "", fmt.Errorf("代理链接不能为空")
	}
	for _, r := range trimmed {
		if r < 32 || r == 127 {
			return "", fmt.Errorf("代理链接不能包含控制字符")
		}
	}
	lower := strings.ToLower(trimmed)
	for _, scheme := range proxynode.SupportedProtocols {
		if strings.HasPrefix(lower, scheme+"://") {
			return trimmed, nil
		}
	}
	return "", fmt.Errorf("不支持的代理协议")
}

func parseRequestedNodes(req NodeTestRequest) ([]proxynode.Node, error) {
	if len(req.URIs) == 0 {
		return nil, fmt.Errorf("没有提供节点链接")
	}
	nodes := make([]proxynode.Node, 0, len(req.URIs))
	for _, uri := range req.URIs {
		target, err := validateProxyURI(uri)
		if err != nil {
			return nil, err
		}
		node, ok := proxynode.ParseNodeURI(target)
		if !ok {
			return nil, fmt.Errorf("解析代理链接失败")
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func defaultProxyWorkDir() string {
	base, err := os.UserConfigDir()
	if err != nil || base == "" {
		base = os.TempDir()
	}
	return filepath.Join(base, "ProxyNodeStudio")
}

func filterTrafficLines(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return "暂无流量日志"
	}
	lines := strings.Split(raw, "\n")
	kept := make([]string, 0, len(lines))
	for _, line := range lines {
		friendly := formatTrafficLine(line)
		lower := strings.ToLower(strings.TrimSpace(friendly))
		if lower == "" {
			continue
		}
		if strings.Contains(lower, "入口") || strings.Contains(lower, "出站") || strings.Contains(lower, "连接") || strings.Contains(lower, "dns") || strings.Contains(lower, "路由") || strings.Contains(lower, "信息") || strings.Contains(lower, "警告") || strings.Contains(lower, "错误") {
			kept = append(kept, friendly)
		}
	}
	if len(kept) == 0 {
		return "暂无流量日志"
	}
	if len(kept) > 80 {
		kept = kept[len(kept)-80:]
	}
	return strings.Join(kept, "\n")
}

func formatTrafficLine(line string) string {
	cleaned := strings.TrimSpace(stripANSIEscapeCodes(line))
	if cleaned == "" {
		return ""
	}
	prefix := singboxLogPrefixPattern.FindStringSubmatch(cleaned)
	if len(prefix) > 0 {
		timestamp := strings.TrimSpace(prefix[2] + " " + prefix[3])
		level := localizeLogLevel(prefix[4])
		message := strings.TrimSpace(cleaned[len(prefix[0]):])
		message = trimLeadingLogMetadata(message)
		message = humanizeTrafficMessage(message)
		if message == "" {
			message = "空日志"
		}
		return fmt.Sprintf("%s｜%s｜%s", timestamp, level, message)
	}
	return humanizeTrafficMessage(cleaned)
}

func stripANSIEscapeCodes(text string) string {
	return ansiEscapePattern.ReplaceAllString(text, "")
}

func trimLeadingLogMetadata(text string) string {
	trimmed := strings.TrimSpace(text)
	for strings.HasPrefix(trimmed, "[") {
		closing := strings.Index(trimmed, "]")
		if closing <= 0 {
			break
		}
		candidate := trimmed[:closing+1]
		if strings.Contains(strings.ToLower(candidate), "proxy") || strings.Contains(strings.ToLower(candidate), "dns") {
			break
		}
		trimmed = strings.TrimSpace(trimmed[closing+1:])
	}
	return trimmed
}

func localizeLogLevel(level string) string {
	switch strings.ToUpper(strings.TrimSpace(level)) {
	case "DEBUG":
		return "调试"
	case "INFO":
		return "信息"
	case "WARN", "WARNING":
		return "警告"
	case "ERROR":
		return "错误"
	default:
		return strings.TrimSpace(level)
	}
}

func humanizeTrafficMessage(message string) string {
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return ""
	}
	if match := outboundTargetPattern.FindStringSubmatch(trimmed); len(match) == 4 {
		protocol := strings.ToUpper(strings.TrimSpace(match[1]))
		outboundTag := strings.TrimSpace(match[2])
		target := humanizeResolvedTarget(match[3])
		if outboundTag == "" {
			return fmt.Sprintf("出站 %s 正在连接 %s", protocol, target)
		}
		return fmt.Sprintf("出站 %s（%s）正在连接 %s", protocol, outboundTag, target)
	}
	if match := dnsLookupPattern.FindStringSubmatch(trimmed); len(match) == 3 {
		return fmt.Sprintf("DNS 查询：%s -> %s", match[2], match[1])
	}
	trimmed = strings.ReplaceAll(trimmed, "inbound/", "入口/")
	trimmed = strings.ReplaceAll(trimmed, "outbound/", "出站/")
	trimmed = strings.ReplaceAll(trimmed, "route", "路由")
	trimmed = strings.ReplaceAll(trimmed, "connection", "连接")
	trimmed = strings.ReplaceAll(trimmed, "detour", "绕行")
	return strings.TrimSpace(trimmed)
}

func humanizeResolvedTarget(target string) string {
	host, port := splitHostPortLoose(target)
	if host == "" {
		return target
	}
	resolvedIPs := logDestinationResolver(host)
	if len(resolvedIPs) == 0 {
		return target
	}
	if port == "" {
		return strings.Join(resolvedIPs, ", ")
	}
	parts := make([]string, 0, len(resolvedIPs))
	for _, ip := range resolvedIPs {
		parts = append(parts, net.JoinHostPort(ip, port))
	}
	return strings.Join(parts, ", ")
}

func splitHostPortLoose(target string) (string, string) {
	if host, port, err := net.SplitHostPort(target); err == nil {
		return strings.Trim(host, "[]"), port
	}
	lastColon := strings.LastIndex(target, ":")
	if lastColon <= 0 || lastColon == len(target)-1 {
		return strings.Trim(target, "[]"), ""
	}
	host := strings.Trim(target[:lastColon], "[]")
	port := target[lastColon+1:]
	if _, err := strconv.Atoi(port); err != nil {
		return strings.Trim(target, "[]"), ""
	}
	return host, port
}

func defaultLogDestinationResolver(host string) []string {
	host = strings.TrimSpace(strings.Trim(host, "[]"))
	if host == "" {
		return nil
	}
	if ip := net.ParseIP(host); ip != nil {
		return []string{ip.String()}
	}
	if cached, ok := logIPCache.Load(host); ok {
		if ips, ok := cached.([]string); ok {
			return append([]string(nil), ips...)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil || len(addrs) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(addrs))
	resolved := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		ip := addr.IP.String()
		if ip == "" {
			continue
		}
		if _, ok := seen[ip]; ok {
			continue
		}
		seen[ip] = struct{}{}
		resolved = append(resolved, ip)
	}
	if len(resolved) > 0 {
		logIPCache.Store(host, resolved)
	}
	return append([]string(nil), resolved...)
}

func readLogTail(path string, maxBytes int) string {
	if strings.TrimSpace(path) == "" {
		return ""
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	if maxBytes > 0 && len(data) > maxBytes {
		data = data[len(data)-maxBytes:]
	}
	return strings.TrimSpace(string(data))
}

func defaultSingBoxBinaryPath() string {
	if env := strings.TrimSpace(os.Getenv("SING_BOX_PATH")); env != "" {
		return env
	}
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		candidates := []string{
			filepath.Join(exeDir, "runtime", binaryName()),
			filepath.Join(exeDir, binaryName()),
			filepath.Join(exeDir, "bin", binaryName()),
		}
		for _, candidate := range candidates {
			if _, statErr := os.Stat(candidate); statErr == nil {
				return candidate
			}
		}
		return candidates[0]
	}
	return binaryName()
}

func binaryName() string {
	if runtime.GOOS == "windows" {
		return "sing-box.exe"
	}
	return "sing-box"
}

type runtimeNodeProber struct {
	pinger HostPinger
}

func (p runtimeNodeProber) Probe(node proxynode.Node) NodeTestResult {
	pinger := p.pinger
	if pinger == nil {
		pinger = commandHostPinger{}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	latency, err := pinger.Ping(ctx, node.Host)
	if err != nil {
		return NodeTestResult{Node: node, Usable: false, LatencyMS: 0, Error: err.Error()}
	}
	return NodeTestResult{Node: node, Usable: true, LatencyMS: latency.Milliseconds()}
}
