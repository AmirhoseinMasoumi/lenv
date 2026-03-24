package cmd

import (
	"fmt"

	"github.com/AmirhoseinMasoumi/lenv/internal/ui"
	"github.com/AmirhoseinMasoumi/lenv/vm"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show running VM status",
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.Title("lenv status")
		statuses, err := vm.ListRunningStatuses()
		if err != nil {
			return err
		}
		if len(statuses) == 0 {
			ui.Warn("No running lenv VMs found.")
			return nil
		}
		for _, s := range statuses {
			ui.Divider()
			ui.KV("Instance", s.Instance)
			ui.KV("Project", s.ProjectDir)
			ui.KV("Running", fmt.Sprintf("%t", s.Running))
			ui.KV("PID", fmt.Sprintf("%d", s.PID))
			ui.KV("SSH Port", fmt.Sprintf("%d", s.SSHPort))
			ui.KV("Distro", s.Distro)
			ui.KV("Accel", s.Accel)
		}
		ui.Divider()
		return nil
	},
}

func init() { rootCmd.AddCommand(statusCmd) }
