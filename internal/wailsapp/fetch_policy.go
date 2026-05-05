package wailsapp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	maxLocalSampleNodes = 10
	maxShareURIs        = 5
	fetchCooldown       = time.Hour
)

type fetchPolicyFile struct {
	LastFetchUnix int64           `json:"last_fetch_unix"`
	SampleURIs    []string        `json:"sample_uris"`
	ProbeDone     map[string]bool `json:"probe_done,omitempty"`
	ProbeUsable   map[string]bool `json:"probe_usable,omitempty"`
}

func (a *App) fetchPolicyPath() string {
	return filepath.Join(defaultProxyWorkDir(), "fetch_policy.json")
}

func (a *App) loadFetchPolicy() {
	a.fetchMu.Lock()
	defer a.fetchMu.Unlock()
	data, err := os.ReadFile(a.fetchPolicyPath())
	if err != nil || len(data) == 0 {
		return
	}
	var st fetchPolicyFile
	if err := json.Unmarshal(data, &st); err != nil {
		return
	}
	if st.LastFetchUnix > 0 {
		a.lastFetchAt = time.Unix(st.LastFetchUnix, 0)
	}
	a.lastSampleURIs = append([]string(nil), st.SampleURIs...)
	if st.ProbeDone != nil {
		a.probeDone = st.ProbeDone
	} else {
		a.probeDone = map[string]bool{}
	}
	if st.ProbeUsable != nil {
		a.probeUsable = st.ProbeUsable
	} else {
		a.probeUsable = map[string]bool{}
	}
}

func (a *App) saveFetchPolicyLocked() {
	st := fetchPolicyFile{
		LastFetchUnix: 0,
		SampleURIs:    append([]string(nil), a.lastSampleURIs...),
		ProbeDone:     a.probeDone,
		ProbeUsable:   a.probeUsable,
	}
	if !a.lastFetchAt.IsZero() {
		st.LastFetchUnix = a.lastFetchAt.Unix()
	}
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return
	}
	dir := filepath.Dir(a.fetchPolicyPath())
	_ = os.MkdirAll(dir, 0o700)
	_ = os.WriteFile(a.fetchPolicyPath(), data, 0o600)
}

// checkFetchAllowedLocked enforces 1h cooldown unless majority of locally cached nodes were probed unusable.
func (a *App) checkFetchAllowedLocked(skip bool) error {
	if skip {
		return nil
	}
	if a.lastFetchAt.IsZero() {
		return nil
	}
	if time.Since(a.lastFetchAt) >= fetchCooldown {
		return nil
	}
	n := len(a.lastSampleURIs)
	if n == 0 {
		return nil
	}
	allProbed := true
	for _, u := range a.lastSampleURIs {
		if u == "" {
			continue
		}
		if !a.probeDone[u] {
			allProbed = false
			break
		}
	}
	if !allProbed {
		return fmt.Errorf("量子抓取冷却中（约 1 小时），且需先对本页全部节点完成测速后，若半数以上不可用方可再次抓取")
	}
	unusable := 0
	for _, u := range a.lastSampleURIs {
		if u == "" {
			continue
		}
		if a.probeDone[u] && !a.probeUsable[u] {
			unusable++
		}
	}
	if unusable > n/2 {
		return nil
	}
	left := fetchCooldown - time.Since(a.lastFetchAt)
	mins := int(left.Round(time.Minute) / time.Minute)
	if mins < 1 {
		mins = 1
	}
	return fmt.Errorf("量子抓取冷却中，约 %d 分钟后可再次尝试", mins)
}
