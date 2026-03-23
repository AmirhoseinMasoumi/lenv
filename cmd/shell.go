package cmd

import (
"fmt"
"time"

"github.com/AmirhoseinMasoumi/lenv/internal/ui"
lssh "github.com/AmirhoseinMasoumi/lenv/ssh"
"github.com/AmirhoseinMasoumi/lenv/vm"
"github.com/spf13/cobra"
)

var shellCmd = &cobra.Command{
Use:   "shell",
Short: "Open an interactive shell in the VM",
RunE: func(cmd *cobra.Command, args []string) error {
dir, err := absProjectDir()
if err != nil {
return err
}
st := vm.GetStatus(dir)
if st.SSHPort == 0 {
return fmt.Errorf("VM not initialized; run `lenv init` first")
}
ui.Step("Connecting SSH")
client, err := lssh.WaitAndConnect(st.SSHPort, 30*time.Second)
if err != nil {
return fmt.Errorf("connect SSH: %w", err)
}
defer client.Close()
ui.Done("connected")
return lssh.OpenShell(client)
},
}

func init() { rootCmd.AddCommand(shellCmd) }
