package ssh

import (
"fmt"
"os"

gssh "golang.org/x/crypto/ssh"
)

func Exec(client *gssh.Client, command string) (int, error) {
sess, err := client.NewSession()
if err != nil {
return 1, fmt.Errorf("create session: %w", err)
}
defer sess.Close()
sess.Stdout = os.Stdout
sess.Stderr = os.Stderr
if err := sess.Run(command); err != nil {
if ee, ok := err.(*gssh.ExitError); ok {
return ee.ExitStatus(), nil
}
return 1, fmt.Errorf("run command: %w", err)
}
return 0, nil
}

