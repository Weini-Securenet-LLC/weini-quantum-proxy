package wailsapp

import (
	"context"
	"errors"
	"testing"
	"time"

	"hermes-agent/proxy-node-studio/internal/proxynode"
)

type fakeHostPinger struct {
	latency time.Duration
	err     error
	hosts   []string
}

func (f *fakeHostPinger) Ping(ctx context.Context, host string) (time.Duration, error) {
	f.hosts = append(f.hosts, host)
	if f.err != nil {
		return 0, f.err
	}
	return f.latency, nil
}

func TestRuntimeNodeProberReturnsPingLatency(t *testing.T) {
	pinger := &fakeHostPinger{latency: 42 * time.Millisecond}
	prober := runtimeNodeProber{pinger: pinger}
	node := proxynode.Node{Protocol: "trojan", Host: "example.com", Port: 443, RawURI: "trojan://secret@example.com:443#demo"}

	result := prober.Probe(node)

	if !result.Usable {
		t.Fatalf("expected usable result, got %#v", result)
	}
	if result.LatencyMS != 42 {
		t.Fatalf("latency_ms = %d, want 42", result.LatencyMS)
	}
	if len(pinger.hosts) != 1 || pinger.hosts[0] != "example.com" {
		t.Fatalf("ping hosts = %#v, want [example.com]", pinger.hosts)
	}
}

func TestRuntimeNodeProberReturnsPingError(t *testing.T) {
	pinger := &fakeHostPinger{err: errors.New("ping timeout")}
	prober := runtimeNodeProber{pinger: pinger}
	node := proxynode.Node{Protocol: "ss", Host: "timeout.example.com", Port: 8388, RawURI: "ss://demo"}

	result := prober.Probe(node)

	if result.Usable {
		t.Fatalf("expected unusable result, got %#v", result)
	}
	if result.Error != "ping timeout" {
		t.Fatalf("error = %q, want ping timeout", result.Error)
	}
}
