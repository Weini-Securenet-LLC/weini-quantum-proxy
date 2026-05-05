//go:build windows

package wailsapp

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

var (
	shell32DLL           = syscall.NewLazyDLL("shell32.dll")
	isUserAnAdminProc    = shell32DLL.NewProc("IsUserAnAdmin")
	shellExecuteWProc    = shell32DLL.NewProc("ShellExecuteW")
)

type windowsElevationController struct{}

func MaybeRelaunchElevated() (bool, error) {
	exe, err := os.Executable()
	if err != nil {
		return false, fmt.Errorf("resolve executable path: %w", err)
	}
	return EnsureElevated(exe, os.Args[1:], windowsElevationController{})
}

func (windowsElevationController) IsAdmin() (bool, error) {
	ret, _, callErr := isUserAnAdminProc.Call()
	if ret == 0 {
		if callErr != syscall.Errno(0) {
			return false, fmt.Errorf("IsUserAnAdmin failed: %w", callErr)
		}
		return false, nil
	}
	return true, nil
}

func (windowsElevationController) RelaunchAsAdministrator(exe string, args []string) error {
	verb := syscall.StringToUTF16Ptr("runas")
	file := syscall.StringToUTF16Ptr(exe)
	params := syscall.StringToUTF16Ptr(joinWindowsCommandLine(args))
	ret, _, callErr := shellExecuteWProc.Call(
		0,
		uintptr(unsafe.Pointer(verb)),
		uintptr(unsafe.Pointer(file)),
		uintptr(unsafe.Pointer(params)),
		0,
		1,
	)
	if ret <= 32 {
		if callErr != syscall.Errno(0) {
			return fmt.Errorf("ShellExecuteW runas failed: %w", callErr)
		}
		return fmt.Errorf("ShellExecuteW runas failed with code %d", ret)
	}
	return nil
}
