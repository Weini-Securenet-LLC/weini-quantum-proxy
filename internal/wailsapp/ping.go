package wailsapp

import (
	"context"
	"fmt"
	"math"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var pingLatencyPattern = regexp.MustCompile(`(?i)time\s*[=<]\s*([0-9]+(?:\.[0-9]+)?)\s*ms`)

type commandHostPinger struct{}

func (commandHostPinger) Ping(ctx context.Context, host string) (time.Duration, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return 0, fmt.Errorf("host is empty")
	}
	args := pingArgs(host)
	cmd := exec.CommandContext(ctx, "ping", args...)
	output, err := cmd.CombinedOutput()
	latency, parseErr := parsePingLatency(string(output))
	if parseErr == nil {
		return latency, nil
	}
	if err != nil {
		msg := strings.TrimSpace(string(output))
		if msg == "" {
			msg = err.Error()
		}
		return 0, fmt.Errorf("%s", msg)
	}
	return 0, parseErr
}

func pingArgs(host string) []string {
	if runtime.GOOS == "windows" {
		return []string{"-n", "1", "-w", "3000", host}
	}
	return []string{"-c", "1", "-W", "3", host}
}

func parsePingLatency(output string) (time.Duration, error) {
	match := pingLatencyPattern.FindStringSubmatch(output)
	if len(match) < 2 {
		return 0, fmt.Errorf("unable to parse ping latency")
	}
	value, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		return 0, fmt.Errorf("parse ping latency: %w", err)
	}
	return time.Duration(math.Round(value * float64(time.Millisecond))), nil
}
