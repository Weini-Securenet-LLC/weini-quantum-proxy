//go:build !windows

package wailsapp

func MaybeRelaunchElevated() (bool, error) {
	return false, nil
}
