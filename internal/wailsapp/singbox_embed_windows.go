//go:build windows

package wailsapp

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:embed assets/windows/sing-box.exe
var embeddedSingBoxWindows []byte

func resolveBundledSingBoxBinaryPath() (string, error) {
	if env := strings.TrimSpace(os.Getenv("SING_BOX_PATH")); env != "" {
		return env, nil
	}
	workDir := defaultProxyWorkDir()
	bundleDir := filepath.Join(workDir, "bundled-runtime")
	binaryPath := filepath.Join(bundleDir, binaryName())
	if err := os.MkdirAll(bundleDir, 0o755); err != nil {
		return "", fmt.Errorf("创建内置运行时目录失败: %w", err)
	}
	needsWrite := true
	if info, err := os.Stat(binaryPath); err == nil && info.Size() == int64(len(embeddedSingBoxWindows)) {
		if currentHash, hashErr := fileSHA256(binaryPath); hashErr == nil {
			if currentHash == bytesSHA256(embeddedSingBoxWindows) {
				needsWrite = false
			}
		}
	}
	if needsWrite {
		tmpPath := binaryPath + ".tmp"
		if err := os.WriteFile(tmpPath, embeddedSingBoxWindows, 0o755); err != nil {
			return "", fmt.Errorf("写入内置 sing-box 失败: %w", err)
		}
		if err := os.Rename(tmpPath, binaryPath); err != nil {
			_ = os.Remove(tmpPath)
			return "", fmt.Errorf("替换内置 sing-box 失败: %w", err)
		}
	}
	return binaryPath, nil
}

func bytesSHA256(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func fileSHA256(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return bytesSHA256(data), nil
}
