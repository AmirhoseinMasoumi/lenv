package cmd

import (
	"fmt"
	"strings"
	"time"

	lssh "github.com/AmirhoseinMasoumi/lenv/ssh"
	"github.com/AmirhoseinMasoumi/lenv/vm"
	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec <instance-name> <command>",
	Short: "Run command in a named running instance",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		instanceName := args[0]
		command := strings.Join(args[1:], " ")
		statuses, err := vm.ListRunningStatuses()
		if err != nil {
			return err
		}
		for _, s := range statuses {
			if s.Instance != instanceName {
				continue
			}
			client, err := lssh.WaitAndConnect(s.SSHPort, 30*time.Second)
			if err != nil {
				return err
			}
			defer client.Close()
			_, err = lssh.Exec(client, command)
			return err
		}
		return fmt.Errorf("instance not found: %s", instanceName)
	},
}

func init() {
	rootCmd.AddCommand(execCmd)
	execCmd.ValidArgsFunction = completeInstances
}
