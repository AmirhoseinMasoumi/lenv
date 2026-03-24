package ssh

import (
	"fmt"
	"os"
	"strconv"

	gssh "golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

func OpenShell(client *gssh.Client) error {
	sess, err := client.NewSession()
if err != nil {
return fmt.Errorf("create session: %w", err)
}
	defer sess.Close()
	sess.Stdout = os.Stdout
	sess.Stderr = os.Stderr
	sess.Stdin = os.Stdin
	width, height := terminalSize()
	modes := gssh.TerminalModes{
		gssh.ECHO:          1,
		gssh.TTY_OP_ISPEED: 14400,
		gssh.TTY_OP_OSPEED: 14400,
	}
	if err := sess.RequestPty("xterm", height, width, modes); err != nil {
		return fmt.Errorf("request pty: %w", err)
	}
	if err := sess.Start(`if command -v bash >/dev/null 2>&1; then exec bash -l; else exec sh -l; fi`); err != nil {
		return fmt.Errorf("start interactive shell: %w", err)
	}
	if err := sess.Wait(); err != nil {
		return fmt.Errorf("wait shell: %w", err)
	}
	return nil
}

func terminalSize() (int, int) {
	fd := int(os.Stdout.Fd())
	if !term.IsTerminal(fd) {
		return 80, 24
	}
	w, h, err := term.GetSize(fd)
	if err != nil || w <= 0 || h <= 0 {
		return 80, 24
	}
	if cols := os.Getenv("COLUMNS"); cols != "" {
		if n, err := strconv.Atoi(cols); err == nil && n > 0 {
			w = n
		}
	}
	if rows := os.Getenv("LINES"); rows != "" {
		if n, err := strconv.Atoi(rows); err == nil && n > 0 {
			h = n
		}
	}
	return w, h
}

