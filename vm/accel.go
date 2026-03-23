package vm

import (
	"os"
	"runtime"
	"strings"
)

func DetectAccel() string {
	if override := strings.ToLower(strings.TrimSpace(os.Getenv("LENV_ACCEL"))); override != "" {
		return override
	}
	switch runtime.GOOS {
	case "windows":
		return "whpx"
	case "darwin":
		return "hvf"
case "linux":
if _, err := os.Stat("/dev/kvm"); err == nil {
return "kvm"
}
}
return "tcg"
}
