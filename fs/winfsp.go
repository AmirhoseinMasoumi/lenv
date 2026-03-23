package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func CheckInstalled() error {
	if runtime.GOOS != "windows" {
		return nil
	}
	base := os.Getenv("LENV_WINFSP_DIR")
	if base == "" {
		base = os.Getenv("ProgramFiles(x86)")
		if base == "" {
			base = os.Getenv("ProgramFiles")
		}
		base = filepath.Join(base, "WinFsp", "bin")
	}

	candidates := []string{
		filepath.Join(base, "winfsp.dll"),
		filepath.Join(base, "winfsp-x64.dll"),
		filepath.Join(base, "winfsp-x86.dll"),
		filepath.Join(base, "winfsp-a64.dll"),
	}
	for _, dll := range candidates {
		if _, err := os.Stat(dll); err == nil {
			return nil
		}
	}

	return fmt.Errorf("WinFsp not found; set LENV_WINFSP_DIR to WinFsp bin dir or install from https://winfsp.dev")
}
