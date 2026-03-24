package cmd

import (
"fmt"

"github.com/AmirhoseinMasoumi/lenv/vm"
"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
Use:   "status",
Short: "Show running VM status",
RunE: func(cmd *cobra.Command, args []string) error {
	statuses, err := vm.ListRunningStatuses()
	if err != nil {
		return err
	}
	if len(statuses) == 0 {
		fmt.Println("No running lenv VMs found.")
		return nil
	}
	for _, s := range statuses {
		fmt.Printf("Instance: %s\nProject: %s\nRunning: %t\nPID: %d\nSSH Port: %d\nDistro: %s\nAccel: %s\n\n",
			s.Instance, s.ProjectDir, s.Running, s.PID, s.SSHPort, s.Distro, s.Accel)
	}
	return nil
},
}

func init() { rootCmd.AddCommand(statusCmd) }

