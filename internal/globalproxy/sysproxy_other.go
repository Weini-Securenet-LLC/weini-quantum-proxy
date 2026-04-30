//go:build !windows

package globalproxy

type systemProxyState struct{}

func applySystemProxy(string) (systemProxyState, error) {
	return systemProxyState{}, nil
}

func restoreSystemProxy(systemProxyState) error {
	return nil
}

func systemProxySupported() bool {
	return false
}

func systemProxyMatches(string) (bool, error) {
	return false, nil
}
