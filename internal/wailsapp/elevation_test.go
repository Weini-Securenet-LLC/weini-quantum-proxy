package wailsapp

import (
	"errors"
	"testing"
)

type fakeElevationController struct {
	admin            bool
	isAdminErr       error
	relaunchErr      error
	relaunchExe      string
	relaunchArgs     []string
	relaunchAttempts int
}

func (f *fakeElevationController) IsAdmin() (bool, error) {
	if f.isAdminErr != nil {
		return false, f.isAdminErr
	}
	return f.admin, nil
}

func (f *fakeElevationController) RelaunchAsAdministrator(exe string, args []string) error {
	f.relaunchExe = exe
	f.relaunchArgs = append([]string(nil), args...)
	f.relaunchAttempts++
	return f.relaunchErr
}

func TestEnsureElevatedRelaunchesWhenNotAdmin(t *testing.T) {
	ctrl := &fakeElevationController{}
	relaunched, err := EnsureElevated("C:/app/维尼量子节点.exe", []string{"--flag", "value"}, ctrl)
	if err != nil {
		t.Fatalf("EnsureElevated error: %v", err)
	}
	if !relaunched {
		t.Fatal("expected relaunch when not admin")
	}
	if ctrl.relaunchAttempts != 1 {
		t.Fatalf("relaunch attempts = %d, want 1", ctrl.relaunchAttempts)
	}
	if ctrl.relaunchExe != "C:/app/维尼量子节点.exe" {
		t.Fatalf("relaunch exe = %q", ctrl.relaunchExe)
	}
}

func TestEnsureElevatedSkipsRelaunchWhenAlreadyAdmin(t *testing.T) {
	ctrl := &fakeElevationController{admin: true}
	relaunched, err := EnsureElevated("C:/app/维尼量子节点.exe", nil, ctrl)
	if err != nil {
		t.Fatalf("EnsureElevated error: %v", err)
	}
	if relaunched {
		t.Fatal("expected no relaunch when already admin")
	}
	if ctrl.relaunchAttempts != 0 {
		t.Fatalf("relaunch attempts = %d, want 0", ctrl.relaunchAttempts)
	}
}

func TestEnsureElevatedReturnsRelaunchError(t *testing.T) {
	ctrl := &fakeElevationController{relaunchErr: errors.New("shell execute failed")}
	relaunched, err := EnsureElevated("C:/app/维尼量子节点.exe", nil, ctrl)
	if err == nil {
		t.Fatal("expected relaunch error")
	}
	if relaunched {
		t.Fatal("expected relaunched=false on error")
	}
}
