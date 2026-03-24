package vm

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/AmirhoseinMasoumi/lenv/internal/ui"
)

func DetectAccel() string {
	if override := strings.ToLower(strings.TrimSpace(os.Getenv("LENV_ACCEL"))); override != "" {
		return override
	}
	supported := qemuAccelSet()
	switch runtime.GOOS {
	case "windows":
		if supported["whpx"] {
			return "whpx"
		}
	case "darwin":
		if supported["hvf"] {
			return "hvf"
		}
	case "linux":
		if _, err := os.Stat("/dev/kvm"); err == nil && supported["kvm"] {
			return "kvm"
		}
	}
	ui.Warn("Hardware acceleration unavailable; falling back to TCG (slower boot).")
	return "tcg"
}

func qemuAccelSet() map[string]bool {
	set := map[string]bool{}
	qemu, err := resolveQEMUPath()
	if err != nil {
		return set
	}
	out, err := exec.Command(qemu, "-accel", "help").CombinedOutput()
	if err != nil {
		return set
	}
	for _, ln := range strings.Split(strings.ToLower(string(out)), "\n") {
		f := strings.Fields(strings.TrimSpace(ln))
		if len(f) == 0 {
			continue
		}
		name := f[0]
		set[name] = true
	}
	return set
}

func DetectAccelWithReason() (string, string) {
	accel := DetectAccel()
	switch accel {
	case "whpx":
		return accel, "Windows Hypervisor Platform detected"
	case "kvm":
		return accel, "/dev/kvm present and supported by QEMU"
	case "hvf":
		return accel, "Hypervisor.framework supported by QEMU"
	default:
		return accel, fmt.Sprintf("fallback selected: %s", accel)
	}
}
