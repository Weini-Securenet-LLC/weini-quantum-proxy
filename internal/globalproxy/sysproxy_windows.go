//go:build windows

package globalproxy

import (
	"fmt"
	"strings"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

const (
	internetOptionRefresh         = 37
	internetOptionSettingsChanged = 39
	internetSettingsKeyPath       = `Software\Microsoft\Windows\CurrentVersion\Internet Settings`
)

var (
	wininetDLL             = syscall.NewLazyDLL("wininet.dll")
	internetSetOptionWProc = wininetDLL.NewProc("InternetSetOptionW")
)

type systemProxyState struct {
	Valid     bool
	Enabled   bool
	Server    string
	Override  string
	AutoConfig string
}

func applySystemProxy(listenAddress string) (systemProxyState, error) {
	previous, err := readSystemProxyState()
	if err != nil {
		return systemProxyState{}, err
	}
	key, err := registry.OpenKey(registry.CURRENT_USER, internetSettingsKeyPath, registry.SET_VALUE)
	if err != nil {
		return systemProxyState{}, fmt.Errorf("open internet settings: %w", err)
	}
	defer key.Close()
	if err := key.SetDWordValue("ProxyEnable", 1); err != nil {
		return systemProxyState{}, fmt.Errorf("set ProxyEnable: %w", err)
	}
	if err := key.SetStringValue("ProxyServer", listenAddress); err != nil {
		return systemProxyState{}, fmt.Errorf("set ProxyServer: %w", err)
	}
	if err := key.SetStringValue("ProxyOverride", "localhost;127.*;<local>"); err != nil {
		return systemProxyState{}, fmt.Errorf("set ProxyOverride: %w", err)
	}
	_ = key.SetStringValue("AutoConfigURL", "")
	if err := notifySystemProxyChanged(); err != nil {
		return systemProxyState{}, err
	}
	matched, err := systemProxyMatches(listenAddress)
	if err != nil {
		return systemProxyState{}, err
	}
	if !matched {
		return systemProxyState{}, fmt.Errorf("Windows reported a different ProxyServer after update; expected %s", listenAddress)
	}
	return previous, nil
}

func restoreSystemProxy(previous systemProxyState) error {
	if !previous.Valid {
		return nil
	}
	key, err := registry.OpenKey(registry.CURRENT_USER, internetSettingsKeyPath, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("open internet settings: %w", err)
	}
	defer key.Close()
	enabled := uint32(0)
	if previous.Enabled {
		enabled = 1
	}
	if err := key.SetDWordValue("ProxyEnable", enabled); err != nil {
		return fmt.Errorf("restore ProxyEnable: %w", err)
	}
	if err := key.SetStringValue("ProxyServer", previous.Server); err != nil {
		return fmt.Errorf("restore ProxyServer: %w", err)
	}
	if err := key.SetStringValue("ProxyOverride", previous.Override); err != nil {
		return fmt.Errorf("restore ProxyOverride: %w", err)
	}
	if err := key.SetStringValue("AutoConfigURL", previous.AutoConfig); err != nil {
		return fmt.Errorf("restore AutoConfigURL: %w", err)
	}
	return notifySystemProxyChanged()
}

func readSystemProxyState() (systemProxyState, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, internetSettingsKeyPath, registry.QUERY_VALUE)
	if err != nil {
		return systemProxyState{}, fmt.Errorf("open internet settings: %w", err)
	}
	defer key.Close()
	state := systemProxyState{Valid: true}
	if enabled, _, err := key.GetIntegerValue("ProxyEnable"); err == nil {
		state.Enabled = enabled != 0
	}
	if server, _, err := key.GetStringValue("ProxyServer"); err == nil {
		state.Server = strings.TrimSpace(server)
	}
	if override, _, err := key.GetStringValue("ProxyOverride"); err == nil {
		state.Override = strings.TrimSpace(override)
	}
	if autoConfig, _, err := key.GetStringValue("AutoConfigURL"); err == nil {
		state.AutoConfig = strings.TrimSpace(autoConfig)
	}
	return state, nil
}

func systemProxyMatches(listenAddress string) (bool, error) {
	state, err := readSystemProxyState()
	if err != nil {
		return false, err
	}
	return state.Enabled && strings.EqualFold(strings.TrimSpace(state.Server), strings.TrimSpace(listenAddress)), nil
}

func notifySystemProxyChanged() error {
	if err := internetSetOption(internetOptionSettingsChanged); err != nil {
		return err
	}
	return internetSetOption(internetOptionRefresh)
}

func internetSetOption(option uintptr) error {
	ret, _, callErr := internetSetOptionWProc.Call(0, option, 0, 0)
	if ret != 0 {
		return nil
	}
	if callErr != syscall.Errno(0) {
		return fmt.Errorf("InternetSetOptionW(%d): %w", option, callErr)
	}
	return fmt.Errorf("InternetSetOptionW(%d) failed", option)
}

func systemProxySupported() bool {
	return true
}
