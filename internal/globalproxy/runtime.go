package globalproxy

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"hermes-agent/proxy-node-studio/internal/proxynode"
)

type Status struct {
	Connected    bool            `json:"connected"`
	Usable       bool            `json:"usable"`
	SystemProxy  bool            `json:"system_proxy"`
	Mode         string          `json:"mode"`
	PID          int             `json:"pid"`
	BinaryPath   string          `json:"binary_path"`
	ConfigPath   string          `json:"config_path"`
	LogPath      string          `json:"log_path"`
	MixedPort    int             `json:"mixed_port"`
	SocksPort    int             `json:"socks_port"`
	CheckMessage string          `json:"check_message"`
	LastError    string          `json:"last_error"`
	Node         *proxynode.Node `json:"node,omitempty"`
}

type Process interface {
	PID() int
	Kill() error
	Wait() error
}

type Runner interface {
	Start(binary string, args []string, stdoutPath string, stderrPath string) (Process, error)
}

type RuntimeController interface {
	Start(node proxynode.Node, opts Options) (Status, error)
	Stop() (Status, error)
	Status() Status
}

type FastStopper interface {
	StopQuick() (Status, error)
}

type ConfigValidator func(binaryPath string, configPath string) error

type ReadyChecker func(status Status) error

type RuntimeOptions struct {
	WorkDir         string
	BinaryPath      string
	Runner          Runner
	ConfigValidator ConfigValidator
	ReadyChecker    ReadyChecker
}

type Runtime struct {
	mu              sync.Mutex
	workDir         string
	binaryPath      string
	runner          Runner
	process         Process
	status          Status
	configValidator ConfigValidator
	readyChecker    ReadyChecker
	proxyState      systemProxyState
	proxyApplied    bool
	statusCheckAt   time.Time
	statusChecking  bool
}

func NewRuntime(opts RuntimeOptions) *Runtime {
	workDir := opts.WorkDir
	if workDir == "" {
		workDir = defaultWorkDir()
	}
	runner := opts.Runner
	if runner == nil {
		runner = execCommandRunner{}
	}
	validator := opts.ConfigValidator
	if validator == nil {
		validator = validateConfigWithSingBox
	}
	checker := opts.ReadyChecker
	if checker == nil {
		checker = verifyRuntimeReady
	}
	return &Runtime{
		workDir:         workDir,
		binaryPath:      opts.BinaryPath,
		runner:          runner,
		configValidator: validator,
		readyChecker:    checker,
		status: Status{
			BinaryPath: opts.BinaryPath,
		},
	}
}

func (r *Runtime) Start(node proxynode.Node, opts Options) (Status, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if opts.Mode == "" {
		opts.Mode = ModeTUN
	}
	if opts.MixedPort == 0 {
		opts.MixedPort = 7890
	}
	if opts.SocksPort == 0 {
		opts.SocksPort = 1080
	}
	if r.process != nil {
		r.restoreSystemProxyLocked()
		_ = r.process.Kill()
		_ = r.process.Wait()
		r.process = nil
	}
	resolvedPorts, err := resolveLocalPorts(opts.MixedPort, opts.SocksPort)
	if err != nil {
		r.status = Status{LastError: err.Error(), Mode: string(opts.Mode), MixedPort: opts.MixedPort, SocksPort: opts.SocksPort}
		return r.status, err
	}
	opts.MixedPort = resolvedPorts.MixedPort
	opts.SocksPort = resolvedPorts.SocksPort
	binaryPath, err := r.resolveBinaryPath()
	if err != nil {
		r.status = Status{BinaryPath: r.binaryPath, LastError: err.Error(), Mode: string(opts.Mode), MixedPort: opts.MixedPort, SocksPort: opts.SocksPort}
		return r.status, err
	}
	if err := os.MkdirAll(r.workDir, 0o755); err != nil {
		r.status = Status{BinaryPath: binaryPath, LastError: err.Error(), Mode: string(opts.Mode), MixedPort: opts.MixedPort, SocksPort: opts.SocksPort}
		return r.status, err
	}
	configText, err := BuildConfig(node, opts)
	if err != nil {
		r.status = Status{BinaryPath: binaryPath, LastError: err.Error(), Mode: string(opts.Mode), MixedPort: opts.MixedPort, SocksPort: opts.SocksPort, Node: &node}
		return r.status, err
	}
	configPath := filepath.Join(r.workDir, "sing-box-config.json")
	logPath := filepath.Join(r.workDir, "sing-box.log")
	if err := os.WriteFile(configPath, []byte(configText), 0o600); err != nil {
		r.status = Status{BinaryPath: binaryPath, LastError: err.Error(), Mode: string(opts.Mode), MixedPort: opts.MixedPort, SocksPort: opts.SocksPort, Node: &node}
		return r.status, err
	}
	if err := ensureFile(logPath); err != nil {
		r.status = Status{BinaryPath: binaryPath, LastError: err.Error(), Mode: string(opts.Mode), MixedPort: opts.MixedPort, SocksPort: opts.SocksPort, Node: &node, ConfigPath: configPath}
		return r.status, err
	}
	if err := truncateFile(logPath); err != nil {
		r.status = Status{BinaryPath: binaryPath, LastError: err.Error(), Mode: string(opts.Mode), MixedPort: opts.MixedPort, SocksPort: opts.SocksPort, Node: &node, ConfigPath: configPath, LogPath: logPath}
		return r.status, err
	}
	if err := r.configValidator(binaryPath, configPath); err != nil {
		msg := fmt.Sprintf("sing-box 配置检查失败: %v", err)
		r.status = Status{BinaryPath: binaryPath, LastError: msg, Mode: string(opts.Mode), MixedPort: opts.MixedPort, SocksPort: opts.SocksPort, Node: &node, ConfigPath: configPath, LogPath: logPath}
		return r.status, fmt.Errorf(msg)
	}
	proc, err := r.runner.Start(binaryPath, []string{"run", "-D", r.workDir, "-c", configPath}, logPath, logPath)
	if err != nil {
		r.status = Status{BinaryPath: binaryPath, LastError: err.Error(), Mode: string(opts.Mode), MixedPort: opts.MixedPort, SocksPort: opts.SocksPort, Node: &node, ConfigPath: configPath, LogPath: logPath}
		return r.status, err
	}
	r.process = proc
	nodeCopy := node
	candidate := Status{
		Connected:  true,
		Usable:     false,
		Mode:       string(opts.Mode),
		PID:        proc.PID(),
		BinaryPath: binaryPath,
		ConfigPath: configPath,
		LogPath:    logPath,
		MixedPort:  opts.MixedPort,
		SocksPort:  opts.SocksPort,
		Node:       &nodeCopy,
	}
	if err := r.readyChecker(candidate); err != nil {
		_ = r.process.Kill()
		_ = r.process.Wait()
		r.process = nil
		msg := strings.TrimSpace(err.Error())
		logTail := readLogTail(logPath, 1200)
		msg = humanizeRuntimeError(msg, logTail)
		if logTail != "" {
			msg = strings.TrimSpace(msg + " | log: " + logTail)
		}
		candidate.Connected = false
		candidate.Usable = false
		candidate.LastError = msg
		candidate.CheckMessage = "代理可用性检查失败"
		candidate.PID = 0
		r.status = candidate
		return r.status, fmt.Errorf(msg)
	}
	if err := r.applySystemProxyLocked(opts, &candidate); err != nil {
		_ = r.process.Kill()
		_ = r.process.Wait()
		r.process = nil
		candidate.Connected = false
		candidate.Usable = false
		candidate.PID = 0
		candidate.LastError = err.Error()
		candidate.CheckMessage = "设置系统代理失败"
		r.status = candidate
		return r.status, err
	}
	candidate.Usable = true
	candidate.CheckMessage = "本地端口已启动；浏览器风格 HTTPS CONNECT 隧道探测通过"
	if candidate.SystemProxy {
		candidate.CheckMessage += "；Windows 浏览器/系统代理已指向本地 mixed 端口"
	}
	candidate.LastError = ""
	r.status = candidate
	r.statusCheckAt = time.Now()
	r.statusChecking = false
	return r.status, nil
}

func (r *Runtime) Stop() (Status, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.stopLocked(150 * time.Millisecond)
}

func (r *Runtime) StopQuick() (Status, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.stopLocked(-1)
}

func (r *Runtime) stopLocked(waitTimeout time.Duration) (Status, error) {
	if r.process == nil {
		r.status.Connected = false
		r.status.Usable = false
		r.status.PID = 0
		r.statusChecking = false
		return r.status, nil
	}
	proc := r.process
	killErr := proc.Kill()
	r.process = nil
	r.restoreSystemProxyLocked()
	r.status.Connected = false
	r.status.Usable = false
	r.status.SystemProxy = false
	r.status.PID = 0
	r.status.CheckMessage = "代理已停止"
	r.statusChecking = false
	if killErr != nil && !errors.Is(killErr, os.ErrProcessDone) {
		r.status.LastError = killErr.Error()
		return r.status, killErr
	}
	waitErr, waitTimedOut := error(nil), false
	if waitTimeout < 0 {
		go func() {
			_ = proc.Wait()
		}()
		waitTimedOut = true
	} else {
		waitErr, waitTimedOut = waitForProcessExit(proc, waitTimeout)
	}
	if waitErr != nil && !errors.Is(waitErr, os.ErrProcessDone) && !isExpectedProcessExitAfterKill(waitErr) {
		r.status.LastError = waitErr.Error()
		return r.status, waitErr
	}
	r.status.LastError = ""
	if waitTimedOut {
		r.status.CheckMessage = "代理已停止，后台正在完成清理"
	}
	return r.status, nil
}

func (r *Runtime) Status() Status {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.process == nil || !r.status.Connected {
		return r.status
	}
	if !r.statusChecking && time.Since(r.statusCheckAt) >= 3*time.Second {
		r.statusChecking = true
		snapshot := r.status
		proc := r.process
		go r.refreshStatusAsync(proc, snapshot)
	}
	return r.status
}

func (r *Runtime) refreshStatusAsync(proc Process, snapshot Status) {
	err := r.readyChecker(snapshot)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.statusCheckAt = time.Now()
	r.statusChecking = false
	if r.process == nil || r.process != proc || !r.status.Connected {
		return
	}
	if err == nil {
		if !r.status.Usable {
			r.status.Usable = true
		}
		if strings.TrimSpace(r.status.CheckMessage) == "" {
			r.status.CheckMessage = "本地端口已启动；浏览器风格 HTTPS CONNECT 隧道探测通过"
			if r.status.SystemProxy {
				r.status.CheckMessage += "；Windows 浏览器/系统代理已指向本地 mixed 端口"
			}
		}
		return
	}
	msg := strings.TrimSpace(err.Error())
	logTail := readLogTail(snapshot.LogPath, 1200)
	msg = humanizeRuntimeError(msg, logTail)
	if logTail != "" {
		msg = strings.TrimSpace(msg + " | log: " + logTail)
	}
	r.status.Connected = false
	r.status.Usable = false
	r.status.SystemProxy = false
	r.status.PID = 0
	r.status.LastError = msg
	r.status.CheckMessage = "代理可用性检查失败"
	r.restoreSystemProxyLocked()
}

func isExpectedProcessExitAfterKill(err error) bool {
	if err == nil {
		return false
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return true
	}
	var errno syscall.Errno
	if errors.As(err, &errno) {
		return errno == syscall.ECHILD
	}
	return false
}

func (r *Runtime) applySystemProxyLocked(opts Options, status *Status) error {
	status.SystemProxy = false
	if opts.MixedPort <= 0 || !systemProxySupported() {
		return nil
	}
	listenAddress := fmt.Sprintf("127.0.0.1:%d", opts.MixedPort)
	state, err := applySystemProxy(listenAddress)
	if err != nil {
		return fmt.Errorf("设置 Windows 系统代理 %s 失败: %w", listenAddress, err)
	}
	r.proxyState = state
	r.proxyApplied = true
	status.SystemProxy = true
	return nil
}

func (r *Runtime) restoreSystemProxyLocked() {
	if !r.proxyApplied {
		return
	}
	_ = restoreSystemProxy(r.proxyState)
	r.proxyState = systemProxyState{}
	r.proxyApplied = false
}

func (r *Runtime) resolveBinaryPath() (string, error) {
	if r.binaryPath == "" {
		return "", fmt.Errorf("未配置 sing-box 可执行文件")
	}
	abs, err := filepath.Abs(r.binaryPath)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(abs); err != nil {
		return "", fmt.Errorf("找不到 sing-box 可执行文件: %s", abs)
	}
	r.binaryPath = abs
	return abs, nil
}

func defaultWorkDir() string {
	base, err := os.UserConfigDir()
	if err != nil || base == "" {
		base = os.TempDir()
	}
	return filepath.Join(base, "ProxyNodeStudio")
}

func ensureFile(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	return f.Close()
}

func truncateFile(path string) error {
	return os.WriteFile(path, nil, 0o600)
}

func resolveLocalPorts(preferredMixedPort int, preferredSocksPort int) (Options, error) {
	mixedPort, err := chooseLocalPort(preferredMixedPort, nil)
	if err != nil {
		return Options{}, fmt.Errorf("无法为 mixed 入口选择本地端口: %w", err)
	}
	socksPort, err := chooseLocalPort(preferredSocksPort, map[int]struct{}{mixedPort: {}})
	if err != nil {
		return Options{}, fmt.Errorf("无法为 socks 入口选择本地端口: %w", err)
	}
	return Options{MixedPort: mixedPort, SocksPort: socksPort}, nil
}

func chooseLocalPort(preferredPort int, forbidden map[int]struct{}) (int, error) {
	if preferredPort > 0 {
		if _, blocked := forbidden[preferredPort]; !blocked && localPortAvailable(preferredPort) {
			return preferredPort, nil
		}
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok || addr.Port <= 0 {
		return 0, fmt.Errorf("failed to determine allocated TCP port")
	}
	if _, blocked := forbidden[addr.Port]; blocked {
		return chooseLocalPort(0, forbidden)
	}
	return addr.Port, nil
}

func localPortAvailable(port int) bool {
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return false
	}
	_ = listener.Close()
	return true
}

func validateConfigWithSingBox(binaryPath string, configPath string) error {
	cmd := exec.Command(binaryPath, "check", "-c", configPath)
	output, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}
	message := strings.TrimSpace(string(output))
	if message == "" {
		message = err.Error()
	}
	return fmt.Errorf(message)
}

func verifyRuntimeReady(status Status) error {
	deadline := time.Now().Add(10 * time.Second)
	ports := []int{}
	if status.MixedPort > 0 {
		ports = append(ports, status.MixedPort)
	}
	if status.SocksPort > 0 {
		ports = append(ports, status.SocksPort)
	}
	for _, port := range ports {
		if err := waitForLocalPort(port, deadline); err != nil {
			return err
		}
	}
	if status.MixedPort == 0 {
		return nil
	}
	proxyURL, _ := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", status.MixedPort))
	client := &http.Client{
		Timeout: 8 * time.Second,
		Transport: &http.Transport{
			Proxy:       http.ProxyURL(proxyURL),
			DialContext: (&net.Dialer{Timeout: 5 * time.Second}).DialContext,
		},
	}
	_ = waitForHTTP204Probe(client, deadline)
	_ = waitForHTTPS204Probe(client, deadline)
	if err := waitForBrowserStyleHTTPSConnect(status.MixedPort, deadline); err != nil {
		return err
	}
	return nil
}

func waitForLocalPort(port int, deadline time.Time) error {
	address := fmt.Sprintf("127.0.0.1:%d", port)
	var lastErr error
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", address, 500*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		lastErr = err
		time.Sleep(250 * time.Millisecond)
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("timed out waiting for %s", address)
	}
	return fmt.Errorf("proxy port %s not ready: %w", address, lastErr)
}

func waitForHTTP204Probe(client *http.Client, deadline time.Time) error {
	return waitForProxyProbe(deadline, func() error {
		resp, err := client.Get("http://cp.cloudflare.com/generate_204")
		if err != nil {
			return fmt.Errorf("plain HTTP probe failed: %w", err)
		}
		defer resp.Body.Close()
		_, _ = io.Copy(io.Discard, resp.Body)
		if resp.StatusCode != http.StatusNoContent {
			return fmt.Errorf("plain HTTP probe expected 204 but got %s", resp.Status)
		}
		return nil
	})
}

func waitForHTTPS204Probe(client *http.Client, deadline time.Time) error {
	return waitForProxyProbe(deadline, func() error {
		resp, err := client.Get("https://cp.cloudflare.com/generate_204")
		if err != nil {
			return fmt.Errorf("HTTPS proxy probe failed: %w", err)
		}
		defer resp.Body.Close()
		_, _ = io.Copy(io.Discard, resp.Body)
		if resp.StatusCode != http.StatusNoContent {
			return fmt.Errorf("HTTPS proxy probe expected 204 but got %s", resp.Status)
		}
		return nil
	})
}

func waitForBrowserStyleHTTPSConnect(port int, deadline time.Time) error {
	proxyAddress := fmt.Sprintf("127.0.0.1:%d", port)
	targetAddress := "1.1.1.1:443"
	return waitForProxyProbe(deadline, func() error {
		conn, err := net.DialTimeout("tcp", proxyAddress, 5*time.Second)
		if err != nil {
			return fmt.Errorf("browser-style CONNECT probe could not reach local mixed port: %w", err)
		}
		defer conn.Close()
		_ = conn.SetDeadline(time.Now().Add(6 * time.Second))
		request := "CONNECT " + targetAddress + " HTTP/1.1\r\nHost: " + targetAddress + "\r\nProxy-Connection: Keep-Alive\r\n\r\n"
		if _, err := io.WriteString(conn, request); err != nil {
			return fmt.Errorf("browser-style CONNECT probe write failed: %w", err)
		}
		reader := bufio.NewReader(conn)
		statusLine, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("browser-style CONNECT probe read failed: %w", err)
		}
		if !strings.Contains(statusLine, " 200 ") {
			return fmt.Errorf("browser-style CONNECT probe expected HTTP 200 tunnel but got %s", strings.TrimSpace(statusLine))
		}
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("browser-style CONNECT probe header read failed: %w", err)
			}
			if line == "\r\n" {
				break
			}
		}
		tlsConn := tls.Client(conn, &tls.Config{ServerName: "one.one.one.one", MinVersion: tls.VersionTLS12})
		defer tlsConn.Close()
		if err := tlsConn.Handshake(); err != nil {
			return fmt.Errorf("browser-style HTTPS tunnel handshake failed: %w", err)
		}
		if _, err := io.WriteString(tlsConn, "GET / HTTP/1.1\r\nHost: one.one.one.one\r\nConnection: close\r\n\r\n"); err != nil {
			return fmt.Errorf("browser-style HTTPS request failed: %w", err)
		}
		respReader := bufio.NewReader(tlsConn)
		respLine, err := respReader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("browser-style HTTPS response read failed: %w", err)
		}
		if !strings.HasPrefix(respLine, "HTTP/1.1 2") && !strings.HasPrefix(respLine, "HTTP/1.1 3") && !strings.HasPrefix(respLine, "HTTP/2 2") && !strings.HasPrefix(respLine, "HTTP/2 3") {
			return fmt.Errorf("browser-style HTTPS request returned unexpected status %s", strings.TrimSpace(respLine))
		}
		return nil
	})
}

func waitForProxyProbe(deadline time.Time, fn func() error) error {
	var lastErr error
	for time.Now().Before(deadline) {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
		}
		time.Sleep(400 * time.Millisecond)
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("proxy probe timed out")
	}
	return lastErr
}

func waitForProcessExit(proc Process, timeout time.Duration) (error, bool) {
	if proc == nil {
		return nil, false
	}
	if timeout <= 0 {
		return proc.Wait(), false
	}
	result := make(chan error, 1)
	go func() {
		result <- proc.Wait()
	}()
	select {
	case err := <-result:
		return err, false
	case <-time.After(timeout):
		go func() {
			<-result
		}()
		return nil, true
	}
}

func readLogTail(path string, maxBytes int64) string {
	if maxBytes <= 0 {
		maxBytes = 1200
	}
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return ""
	}
	start := info.Size() - maxBytes
	if start < 0 {
		start = 0
	}
	if _, err := file.Seek(start, io.SeekStart); err != nil {
		return ""
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func humanizeRuntimeError(message string, logTail string) string {
	base := strings.TrimSpace(message)
	combined := strings.ToLower(strings.TrimSpace(message + "\n" + logTail))
	if strings.Contains(combined, "configure tun interface: access is denied") || strings.Contains(combined, "access is denied") {
		if !strings.Contains(strings.ToLower(base), "access is denied") {
			base = strings.TrimSpace(base + "; sing-box reported: Access is denied")
		}
		return strings.TrimSpace(base + " —— Windows 的 TUN 模式需要管理员权限。请完全退出程序后，右键 ProxyNodeStudio-wails.exe，选择“以管理员身份运行”。如果仍然失败，可能是其他 VPN 或安全软件拦截了 TUN 创建。")
	}
	if strings.Contains(combined, "only one usage of each socket address") || strings.Contains(combined, "address already in use") || strings.Contains(combined, "bind:") && strings.Contains(combined, "127.0.0.1:7890") {
		return strings.TrimSpace(base + " —— 本地 127.0.0.1:7890 端口已被其他程序占用（常见是 Clash、v2rayN、nekoray 或旧版程序实例）。关闭占用 7890 的程序后再试，或使用新版自动改用空闲端口。")
	}
	if strings.Contains(combined, "https proxy probe failed") || strings.Contains(combined, "browser-style connect probe") || strings.Contains(combined, "https tunnel handshake failed") {
		return strings.TrimSpace(base + " —— 节点虽然打开了本地代理端口，但没有通过浏览器风格的 HTTPS 隧道检测。这通常表示它不能稳定承载正常浏览器的 HTTPS 流量。")
	}
	if strings.Contains(combined, "plain http probe expected 204") || strings.Contains(combined, "https proxy probe expected 204") {
		return strings.TrimSpace(base + " —— 代理返回了异常的连通性检测结果，这个节点可能被劫持、被封锁，或行为不像正常互联网代理。")
	}
	return base
}

type execCommandRunner struct{}

type execProcess struct{ cmd *exec.Cmd }

func (p execProcess) PID() int {
	if p.cmd == nil || p.cmd.Process == nil {
		return 0
	}
	return p.cmd.Process.Pid
}

func (p execProcess) Kill() error {
	if p.cmd == nil || p.cmd.Process == nil {
		return nil
	}
	return p.cmd.Process.Kill()
}

func (p execProcess) Wait() error {
	if p.cmd == nil {
		return nil
	}
	return p.cmd.Wait()
}

func (execCommandRunner) Start(binary string, args []string, stdoutPath string, stderrPath string) (Process, error) {
	stdout, err := os.OpenFile(stdoutPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, err
	}
	stderr, err := os.OpenFile(stderrPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		_ = stdout.Close()
		return nil, err
	}
	cmd := exec.Command(binary, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	applyPlatformSysProcAttr(cmd)
	if err := cmd.Start(); err != nil {
		_ = stdout.Close()
		_ = stderr.Close()
		return nil, err
	}
	return execProcess{cmd: cmd}, nil
}
