package cmd

import (
"fmt"

"github.com/AmirhoseinMasoumi/lenv/vm"
"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
Use:   "status",
Short: "Show VM status for current project",
RunE: func(cmd *cobra.Command, args []string) error {
dir, err := absProjectDir()
if err != nil {
return err
}
s := vm.GetStatus(dir)
fmt.Printf("Instance: %s\nRunning: %t\nPID: %d\nSSH Port: %d\n", s.Instance, s.Running, s.PID, s.SSHPort)
return nil
},
}

func init() { rootCmd.AddCommand(statusCmd) }
