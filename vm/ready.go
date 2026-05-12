package vm

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ReadyMarker is the sentinel emitted by the guest cloud-init / OpenRC local
// service to /dev/console once the VM has finished booting. lenv tails QEMU's
// serial-redirected log file for this string instead of blindly polling SSH.
const ReadyMarker = "<<<LENV_VM_READY>>>"

// QEMULogPath is where startQEMU on Windows captures stdout/stderr (which
// includes -nographic serial output). Linux/macOS daemonize and have no
// equivalent file; on those platforms callers should fall back to SSH probing.
func QEMULogPath(projectDir string) string {
	return filepath.Join(StateDir(projectDir), "qemu.log")
}

// WaitForReadyMarker tails the QEMU serial log until the ready marker appears
// or the deadline elapses. Returns nil on success. If the log file does not
// exist (non-Windows daemonized launch), returns os.ErrNotExist immediately so
// callers can fall back to SSH probing.
func WaitForReadyMarker(ctx context.Context, projectDir string, timeout time.Duration) error {
	logPath := QEMULogPath(projectDir)
	deadline := time.Now().Add(timeout)

	var f *os.File
	for {
		var err error
		f, err = os.Open(logPath)
		if err == nil {
			break
		}
		if !os.IsNotExist(err) {
			return err
		}
		if time.Now().After(deadline) {
			return os.ErrNotExist
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(200 * time.Millisecond):
		}
	}
	defer f.Close()

	buf := make([]byte, 8192)
	var carry string
	for {
		n, err := f.Read(buf)
		if n > 0 {
			chunk := carry + string(buf[:n])
			if strings.Contains(chunk, ReadyMarker) {
				return nil
			}
			// Keep a tail in case the marker straddles a read boundary.
			if len(chunk) > len(ReadyMarker) {
				carry = chunk[len(chunk)-len(ReadyMarker):]
			} else {
				carry = chunk
			}
		}
		if err == io.EOF || n == 0 {
			if time.Now().After(deadline) {
				return fmt.Errorf("ready marker not seen within %s", timeout)
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(250 * time.Millisecond):
			}
			continue
		}
		if err != nil {
			return err
		}
	}
}
