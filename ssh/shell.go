package ssh

import (
"fmt"
"os"

gssh "golang.org/x/crypto/ssh"
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
modes := gssh.TerminalModes{gssh.ECHO: 1, gssh.TTY_OP_ISPEED: 14400, gssh.TTY_OP_OSPEED: 14400}
if err := sess.RequestPty("xterm", 80, 40, modes); err != nil {
return fmt.Errorf("request pty: %w", err)
}
if err := sess.Shell(); err != nil {
return fmt.Errorf("start shell: %w", err)
}
if err := sess.Wait(); err != nil {
return fmt.Errorf("wait shell: %w", err)
}
return nil
}
