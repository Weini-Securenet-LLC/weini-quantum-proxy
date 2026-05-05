//go:build windows

package globalproxy

import (
	"os/exec"
	"syscall"
)

func applyPlatformSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}
