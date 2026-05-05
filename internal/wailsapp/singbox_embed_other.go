//go:build !windows

package wailsapp

func resolveBundledSingBoxBinaryPath() (string, error) {
	return defaultSingBoxBinaryPath(), nil
}
