//go:build windows

package progress

import (
	"os"
	"sync"

	"golang.org/x/sys/windows"
)

var vtEnabled sync.Once

func enableVirtualTerminal() {
	vtEnabled.Do(func() {
		stdout := windows.Handle(os.Stdout.Fd())
		var mode uint32
		if err := windows.GetConsoleMode(stdout, &mode); err == nil {
			mode |= windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
			_ = windows.SetConsoleMode(stdout, mode)
		}
	})
}
