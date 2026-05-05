//go:build !windows

package globalproxy

import "os/exec"

func applyPlatformSysProcAttr(cmd *exec.Cmd) {}
