package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

"github.com/AmirhoseinMasoumi/lenv/internal/ui"
lssh "github.com/AmirhoseinMasoumi/lenv/ssh"
"github.com/AmirhoseinMasoumi/lenv/vm"
"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
Use:   "run <command>",
Short: "Run command in VM",
Args:  cobra.MinimumNArgs(1),
RunE: func(cmd *cobra.Command, args []string) error {
dir, err := absProjectDir()
if err != nil {
return err
}
st := vm.GetStatus(dir)
if st.SSHPort == 0 {
return fmt.Errorf("VM not initialized; run `lenv init` first")
}
	if !st.Running {
		return fmt.Errorf("VM is not running; run `lenv init` first")
	}
	client, err := lssh.WaitAndConnect(st.SSHPort, runSSHTimeout())
if err != nil {
return fmt.Errorf("connect SSH: %w", err)
}
defer client.Close()
command := strings.Join(args, " ")
ui.Info("Running: " + command)
exitCode, err := lssh.Exec(client, command)
if err != nil {
return err
}
if exitCode != 0 {
os.Exit(exitCode)
}
return nil
},
}

func init() { rootCmd.AddCommand(runCmd) }

func runSSHTimeout() time.Duration {
	if raw := os.Getenv("LENV_SSH_WAIT_TIMEOUT_SECONDS"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			return time.Duration(n) * time.Second
		}
	}
	return 120 * time.Second
}

