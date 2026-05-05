package globalproxy

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"hermes-agent/proxy-node-studio/internal/proxynode"
)

type fakeProcess struct {
	pid        int
	killCalled bool
	waitCalled bool
	waitErr    error
}

func (p *fakeProcess) PID() int { return p.pid }
func (p *fakeProcess) Kill() error {
	p.killCalled = true
	return nil
}
func (p *fakeProcess) Wait() error {
	p.waitCalled = true
	return p.waitErr
}

type slowWaitProcess struct {
	fakeProcess
	delay         time.Duration
	waitStartedAt time.Time
	waitDoneAt    time.Time
	killedAt      time.Time
	waitDone      chan struct{}
}

func (p *slowWaitProcess) Kill() error {
	p.killedAt = time.Now()
	return p.fakeProcess.Kill()
}

func (p *slowWaitProcess) Wait() error {
	p.waitCalled = true
	p.waitStartedAt = time.Now()
	time.Sleep(p.delay)
	p.waitDoneAt = time.Now()
	if p.waitDone != nil {
		close(p.waitDone)
	}
	return p.waitErr
}

func (p *slowWaitProcess) waitFinishedWithin(timeout time.Duration) bool {
	if p.waitDone == nil {
		return true
	}
	select {
	case <-p.waitDone:
		return true
	case <-time.After(timeout):
		return false
	}
}

type fakeRunner struct {
	lastBinary string
	lastArgs   []string
	lastStdout string
	lastStderr string
	process    Process
	err        error
	calls      int
}

func (r *fakeRunner) Start(binary string, args []string, stdoutPath string, stderrPath string) (Process, error) {
	r.calls++
	r.lastBinary = binary
	r.lastArgs = append([]string(nil), args...)
	r.lastStdout = stdoutPath
	r.lastStderr = stderrPath
	if r.err != nil {
		return nil, r.err
	}
	if r.process == nil {
		r.process = &fakeProcess{pid: 4321}
	}
	return r.process, nil
}

func testNode() proxynode.Node {
	return proxynode.Node{
		Protocol:   "trojan",
		Name:       "demo",
		Host:       "example.com",
		Port:       443,
		Credential: "secret",
		Network:    "tcp",
		TLS:        "tls",
		RawURI:     "trojan://secret@example.com:443?type=tcp#demo",
	}
}

func TestRuntimeStartWritesConfigAndLaunchesSingBox(t *testing.T) {
	workdir := t.TempDir()
	binary := filepath.Join(workdir, "sing-box.exe")
	if err := os.WriteFile(binary, []byte("fake-binary"), 0o755); err != nil {
		t.Fatalf("write binary: %v", err)
	}
	runner := &fakeRunner{}
	rt := NewRuntime(RuntimeOptions{
		WorkDir:         workdir,
		BinaryPath:      binary,
		Runner:          runner,
		ConfigValidator: func(_, _ string) error { return nil },
		ReadyChecker:    func(Status) error { return nil },
	})

	status, err := rt.Start(testNode(), Options{Mode: ModeTUN})
	if err != nil {
		t.Fatalf("Start error: %v", err)
	}
	if !status.Connected {
		t.Fatal("expected connected status")
	}
	if status.PID != 4321 {
		t.Fatalf("PID = %d, want 4321", status.PID)
	}
	if runner.lastBinary != binary {
		t.Fatalf("binary = %q, want %q", runner.lastBinary, binary)
	}
	if len(runner.lastArgs) < 4 || runner.lastArgs[0] != "run" {
		t.Fatalf("unexpected args: %#v", runner.lastArgs)
	}
	if filepath.Base(status.ConfigPath) != "sing-box-config.json" {
		t.Fatalf("config path = %q", status.ConfigPath)
	}
	if _, err := os.Stat(status.ConfigPath); err != nil {
		t.Fatalf("expected config file: %v", err)
	}
	if _, err := os.Stat(status.LogPath); err != nil {
		t.Fatalf("expected log file: %v", err)
	}
}

func TestRuntimeStartRequiresBinary(t *testing.T) {
	rt := NewRuntime(RuntimeOptions{WorkDir: t.TempDir(), BinaryPath: filepath.Join(t.TempDir(), "missing.exe"), Runner: &fakeRunner{}})
	_, err := rt.Start(testNode(), Options{Mode: ModeTUN})
	if err == nil {
		t.Fatal("expected missing binary error")
	}
}

func TestRuntimeStopKillsProcessAndClearsConnection(t *testing.T) {
	workdir := t.TempDir()
	binary := filepath.Join(workdir, "sing-box.exe")
	if err := os.WriteFile(binary, []byte("fake-binary"), 0o755); err != nil {
		t.Fatalf("write binary: %v", err)
	}
	process := &fakeProcess{pid: 9001}
	rt := NewRuntime(RuntimeOptions{
		WorkDir:         workdir,
		BinaryPath:      binary,
		Runner:          &fakeRunner{process: process},
		ConfigValidator: func(_, _ string) error { return nil },
		ReadyChecker:    func(Status) error { return nil },
	})
	if _, err := rt.Start(testNode(), Options{Mode: ModeTUN}); err != nil {
		t.Fatalf("Start error: %v", err)
	}
	status, err := rt.Stop()
	if err != nil {
		t.Fatalf("Stop error: %v", err)
	}
	if status.Connected {
		t.Fatal("expected disconnected status")
	}
	if !process.killCalled {
		t.Fatal("expected process kill")
	}
}

func TestRuntimeStopIgnoresExitErrorAfterKill(t *testing.T) {
	workdir := t.TempDir()
	binary := filepath.Join(workdir, "sing-box.exe")
	if err := os.WriteFile(binary, []byte("fake-binary"), 0o755); err != nil {
		t.Fatalf("write binary: %v", err)
	}
	process := &fakeProcess{pid: 9002, waitErr: &exec.ExitError{}}
	rt := NewRuntime(RuntimeOptions{
		WorkDir:         workdir,
		BinaryPath:      binary,
		Runner:          &fakeRunner{process: process},
		ConfigValidator: func(_, _ string) error { return nil },
		ReadyChecker:    func(Status) error { return nil },
	})
	if _, err := rt.Start(testNode(), Options{Mode: ModeTUN}); err != nil {
		t.Fatalf("Start error: %v", err)
	}
	status, err := rt.Stop()
	if err != nil {
		t.Fatalf("Stop error: %v", err)
	}
	if status.Connected {
		t.Fatal("expected disconnected status")
	}
	if status.LastError != "" {
		t.Fatalf("expected no last error, got %q", status.LastError)
	}
	if !process.killCalled || !process.waitCalled {
		t.Fatal("expected process kill and wait")
	}
}

func TestRuntimeStopReturnsQuicklyWhenWaitDragsOn(t *testing.T) {
	workdir := t.TempDir()
	binary := filepath.Join(workdir, "sing-box.exe")
	if err := os.WriteFile(binary, []byte("fake-binary"), 0o755); err != nil {
		t.Fatalf("write binary: %v", err)
	}
	process := &slowWaitProcess{
		fakeProcess: fakeProcess{pid: 9003},
		delay:       2 * time.Second,
		waitDone:    make(chan struct{}),
	}
	rt := NewRuntime(RuntimeOptions{
		WorkDir:         workdir,
		BinaryPath:      binary,
		Runner:          &fakeRunner{},
		ConfigValidator: func(_, _ string) error { return nil },
		ReadyChecker:    func(Status) error { return nil },
	})
	rt.process = process
	rt.status = Status{Connected: true, PID: process.pid}

	start := time.Now()
	status, err := rt.Stop()
	if err != nil {
		t.Fatalf("Stop error: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 800*time.Millisecond {
		t.Fatalf("Stop took %s, expected bounded wait", elapsed)
	}
	if status.Connected {
		t.Fatal("expected disconnected status")
	}
	if status.PID != 0 {
		t.Fatalf("expected cleared pid, got %d", status.PID)
	}
	if !strings.Contains(status.CheckMessage, "后台") {
		t.Fatalf("expected background cleanup message, got %q", status.CheckMessage)
	}
	if !process.killCalled || !process.waitCalled {
		t.Fatal("expected process kill and async wait")
	}
	if process.killedAt.IsZero() || process.waitStartedAt.IsZero() {
		t.Fatal("expected kill and wait timestamps")
	}
	if process.waitStartedAt.Before(process.killedAt) {
		t.Fatal("wait should start after kill")
	}
	if !process.waitFinishedWithin(3 * time.Second) {
		t.Fatal("background wait did not finish")
	}
}

func TestRuntimeStopQuickReturnsImmediately(t *testing.T) {
	workdir := t.TempDir()
	binary := filepath.Join(workdir, "sing-box.exe")
	if err := os.WriteFile(binary, []byte("fake-binary"), 0o755); err != nil {
		t.Fatalf("write binary: %v", err)
	}
	process := &slowWaitProcess{
		fakeProcess: fakeProcess{pid: 9004},
		delay:       2 * time.Second,
		waitDone:    make(chan struct{}),
	}
	rt := NewRuntime(RuntimeOptions{
		WorkDir:         workdir,
		BinaryPath:      binary,
		Runner:          &fakeRunner{},
		ConfigValidator: func(_, _ string) error { return nil },
		ReadyChecker:    func(Status) error { return nil },
	})
	rt.process = process
	rt.status = Status{Connected: true, PID: process.pid}

	start := time.Now()
	status, err := rt.StopQuick()
	if err != nil {
		t.Fatalf("StopQuick error: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 250*time.Millisecond {
		t.Fatalf("StopQuick took %s, expected near-immediate return", elapsed)
	}
	if status.Connected {
		t.Fatal("expected disconnected status")
	}
	if !strings.Contains(status.CheckMessage, "后台") {
		t.Fatalf("expected background cleanup message, got %q", status.CheckMessage)
	}
	if !process.waitFinishedWithin(3 * time.Second) {
		t.Fatal("background wait did not finish")
	}
}

func TestRuntimeStatusDoesNotBlockStopDuringSlowHealthCheck(t *testing.T) {
	workdir := t.TempDir()
	binary := filepath.Join(workdir, "sing-box.exe")
	if err := os.WriteFile(binary, []byte("fake-binary"), 0o755); err != nil {
		t.Fatalf("write binary: %v", err)
	}
	process := &slowWaitProcess{
		fakeProcess: fakeProcess{pid: 9005},
		delay:       2 * time.Second,
		waitDone:    make(chan struct{}),
	}
	checkerStarted := make(chan struct{}, 1)
	checkerRelease := make(chan struct{})
	rt := NewRuntime(RuntimeOptions{
		WorkDir:         workdir,
		BinaryPath:      binary,
		Runner:          &fakeRunner{},
		ConfigValidator: func(_, _ string) error { return nil },
		ReadyChecker: func(Status) error {
			checkerStarted <- struct{}{}
			<-checkerRelease
			return nil
		},
	})
	rt.process = process
	rt.status = Status{Connected: true, Usable: true, PID: process.pid, MixedPort: 7890, SocksPort: 1080}
	rt.statusCheckAt = time.Now().Add(-4 * time.Second)

	statusReturned := make(chan struct{})
	go func() {
		_ = rt.Status()
		close(statusReturned)
	}()

	select {
	case <-checkerStarted:
	case <-time.After(time.Second):
		t.Fatal("expected async health check to start")
	}

	select {
	case <-statusReturned:
	case <-time.After(250 * time.Millisecond):
		t.Fatal("Status should return immediately even while health check is still running")
	}

	start := time.Now()
	status, err := rt.Stop()
	if err != nil {
		t.Fatalf("Stop error: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 800*time.Millisecond {
		t.Fatalf("Stop took %s while health check was in flight, expected bounded return", elapsed)
	}
	if status.Connected {
		t.Fatal("expected disconnected status")
	}
	close(checkerRelease)
	if !process.waitFinishedWithin(3 * time.Second) {
		t.Fatal("background wait did not finish")
	}
}

func TestRuntimeStatusReportsLastError(t *testing.T) {
	workdir := t.TempDir()
	binary := filepath.Join(workdir, "sing-box.exe")
	if err := os.WriteFile(binary, []byte("fake-binary"), 0o755); err != nil {
		t.Fatalf("write binary: %v", err)
	}
	runner := &fakeRunner{err: errors.New("launch failed")}
	rt := NewRuntime(RuntimeOptions{
		WorkDir:         workdir,
		BinaryPath:      binary,
		Runner:          runner,
		ConfigValidator: func(_, _ string) error { return nil },
		ReadyChecker:    func(Status) error { return nil },
	})
	_, err := rt.Start(testNode(), Options{Mode: ModeTUN})
	if err == nil {
		t.Fatal("expected launch error")
	}
	status := rt.Status()
	if status.LastError == "" {
		t.Fatal("expected last error in status")
	}
	if status.Connected {
		t.Fatal("expected disconnected status")
	}
}

func TestRuntimeStartFailsValidationBeforeLaunch(t *testing.T) {
	workdir := t.TempDir()
	binary := filepath.Join(workdir, "sing-box.exe")
	if err := os.WriteFile(binary, []byte("fake-binary"), 0o755); err != nil {
		t.Fatalf("write binary: %v", err)
	}
	runner := &fakeRunner{}
	rt := NewRuntime(RuntimeOptions{
		WorkDir:         workdir,
		BinaryPath:      binary,
		Runner:          runner,
		ConfigValidator: func(_, _ string) error { return fmt.Errorf("invalid config") },
		ReadyChecker:    func(Status) error { return nil },
	})
	status, err := rt.Start(testNode(), Options{Mode: ModeTUN})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if runner.calls != 0 {
		t.Fatalf("runner should not start on invalid config, got %d calls", runner.calls)
	}
	if status.Connected {
		t.Fatal("expected disconnected status")
	}
	if status.LastError == "" {
		t.Fatal("expected last error")
	}
}

func TestRuntimeStartFailsReadinessCheckAndKillsProcess(t *testing.T) {
	workdir := t.TempDir()
	binary := filepath.Join(workdir, "sing-box.exe")
	if err := os.WriteFile(binary, []byte("fake-binary"), 0o755); err != nil {
		t.Fatalf("write binary: %v", err)
	}
	proc := &fakeProcess{pid: 777}
	rt := NewRuntime(RuntimeOptions{
		WorkDir:         workdir,
		BinaryPath:      binary,
		Runner:          &fakeRunner{process: proc},
		ConfigValidator: func(_, _ string) error { return nil },
		ReadyChecker:    func(Status) error { return fmt.Errorf("proxy port not listening") },
	})
	status, err := rt.Start(testNode(), Options{Mode: ModeTUN})
	if err == nil {
		t.Fatal("expected readiness error")
	}
	if !proc.killCalled {
		t.Fatal("expected process kill after readiness failure")
	}
	if status.Connected {
		t.Fatal("expected disconnected status")
	}
	if status.LastError == "" {
		t.Fatal("expected readiness error message")
	}
}

func TestResolveLocalPortsFallsBackWhenPreferredPortsBusy(t *testing.T) {
	mixedBusy, err := net.Listen("tcp", "127.0.0.1:7890")
	if err != nil {
		t.Fatalf("listen 7890: %v", err)
	}
	defer mixedBusy.Close()
	socksBusy, err := net.Listen("tcp", "127.0.0.1:1080")
	if err != nil {
		t.Fatalf("listen 1080: %v", err)
	}
	defer socksBusy.Close()

	ports, err := resolveLocalPorts(7890, 1080)
	if err != nil {
		t.Fatalf("resolveLocalPorts error: %v", err)
	}
	if ports.MixedPort == 7890 {
		t.Fatalf("expected fallback mixed port, still got %d", ports.MixedPort)
	}
	if ports.SocksPort == 1080 {
		t.Fatalf("expected fallback socks port, still got %d", ports.SocksPort)
	}
	if ports.MixedPort == ports.SocksPort {
		t.Fatalf("expected distinct ports, got %d", ports.MixedPort)
	}
}

func TestRuntimeStartRewritesBusyPreferredPortsBeforeLaunch(t *testing.T) {
	workdir := t.TempDir()
	binary := filepath.Join(workdir, "sing-box.exe")
	if err := os.WriteFile(binary, []byte("fake-binary"), 0o755); err != nil {
		t.Fatalf("write binary: %v", err)
	}
	mixedBusy, err := net.Listen("tcp", "127.0.0.1:7890")
	if err != nil {
		t.Fatalf("listen 7890: %v", err)
	}
	defer mixedBusy.Close()
	socksBusy, err := net.Listen("tcp", "127.0.0.1:1080")
	if err != nil {
		t.Fatalf("listen 1080: %v", err)
	}
	defer socksBusy.Close()

	rt := NewRuntime(RuntimeOptions{
		WorkDir:         workdir,
		BinaryPath:      binary,
		Runner:          &fakeRunner{},
		ConfigValidator: func(_, _ string) error { return nil },
		ReadyChecker:    func(Status) error { return nil },
	})
	status, err := rt.Start(testNode(), Options{Mode: ModeTUN, MixedPort: 7890, SocksPort: 1080})
	if err != nil {
		t.Fatalf("Start error: %v", err)
	}
	if status.MixedPort == 7890 {
		t.Fatalf("expected fallback mixed port, still got %d", status.MixedPort)
	}
	if status.SocksPort == 1080 {
		t.Fatalf("expected fallback socks port, still got %d", status.SocksPort)
	}
	if status.MixedPort == status.SocksPort {
		t.Fatalf("expected distinct ports, got %d", status.MixedPort)
	}
	configBytes, err := os.ReadFile(status.ConfigPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	configText := string(configBytes)
	if strings.Contains(configText, `"listen_port": 7890`) {
		t.Fatalf("config should not keep busy mixed port: %s", configText)
	}
	if strings.Contains(configText, `"listen_port": 1080`) {
		t.Fatalf("config should not keep busy socks port: %s", configText)
	}
}

func TestRuntimeStatusRechecksUsability(t *testing.T) {
	workdir := t.TempDir()
	binary := filepath.Join(workdir, "sing-box.exe")
	if err := os.WriteFile(binary, []byte("fake-binary"), 0o755); err != nil {
		t.Fatalf("write binary: %v", err)
	}
	checkerShouldFail := false
	rt := NewRuntime(RuntimeOptions{
		WorkDir:         workdir,
		BinaryPath:      binary,
		Runner:          &fakeRunner{process: &fakeProcess{pid: 888}},
		ConfigValidator: func(_, _ string) error { return nil },
		ReadyChecker: func(Status) error {
			if checkerShouldFail {
				return fmt.Errorf("probe failed")
			}
			return nil
		},
	})
	if _, err := rt.Start(testNode(), Options{Mode: ModeTUN}); err != nil {
		t.Fatalf("Start error: %v", err)
	}
	checkerShouldFail = true
	rt.statusCheckAt = time.Now().Add(-4 * time.Second)
	status := rt.Status()
	if !status.Connected || !status.Usable {
		t.Fatalf("expected immediate cached status while async recheck starts, got %#v", status)
	}
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		status = rt.Status()
		if !status.Connected && !status.Usable {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if status.Connected || status.Usable {
		t.Fatal("expected disconnected unusable status after failed recheck")
	}
	if status.LastError == "" {
		t.Fatal("expected recheck error message")
	}
}

func TestHumanizeRuntimeErrorExplainsWindowsAdminRequirement(t *testing.T) {
	msg := humanizeRuntimeError("proxy port 127.0.0.1:7890 not ready", "+0800 2026-04-28 14:49:33 INFO network updated\nFATAL start inbound/tun[tun-in]: configure tun interface: Access is denied.")
	if msg == "" {
		t.Fatal("expected message")
	}
	if !strings.Contains(msg, "TUN 模式需要管理员权限") {
		t.Fatalf("expected admin guidance, got: %s", msg)
	}
	if !strings.Contains(msg, "Access is denied") {
		t.Fatalf("expected original access denied detail, got: %s", msg)
	}
}

func TestRuntimeStartUsesUpdatedBrowserStyleSuccessMessage(t *testing.T) {
	workdir := t.TempDir()
	binary := filepath.Join(workdir, "sing-box.exe")
	if err := os.WriteFile(binary, []byte("fake-binary"), 0o755); err != nil {
		t.Fatalf("write binary: %v", err)
	}
	rt := NewRuntime(RuntimeOptions{
		WorkDir:         workdir,
		BinaryPath:      binary,
		Runner:          &fakeRunner{process: &fakeProcess{pid: 1234}},
		ConfigValidator: func(_, _ string) error { return nil },
		ReadyChecker:    func(Status) error { return nil },
	})
	status, err := rt.Start(testNode(), Options{Mode: ModeTUN})
	if err != nil {
		t.Fatalf("Start error: %v", err)
	}
	if !status.Connected || !status.Usable {
		t.Fatalf("expected connected usable status, got %#v", status)
	}
	if !strings.Contains(status.CheckMessage, "浏览器风格 HTTPS CONNECT") {
		t.Fatalf("unexpected check message: %q", status.CheckMessage)
	}
	if strings.Contains(status.CheckMessage, "HTTPS 与浏览器 CONNECT") {
		t.Fatalf("old strict probe wording should be removed: %q", status.CheckMessage)
	}
}
