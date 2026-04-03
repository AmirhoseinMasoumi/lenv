//go:build !windows

package progress

func enableVirtualTerminal() {
	// No-op on non-Windows platforms
}
