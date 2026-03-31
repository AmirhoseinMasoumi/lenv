package cmd

import (
	"github.com/AmirhoseinMasoumi/lenv/fs"
	"github.com/AmirhoseinMasoumi/lenv/internal/ui"
	"github.com/AmirhoseinMasoumi/lenv/vm"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the running VM for the current project",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := absProjectDir()
		if err != nil {
			return err
		}
		ui.Title("lenv stop")
		ui.KV("Project", dir)
		ui.Divider()
		st := vm.GetStatus(dir)
		if !st.Running {
			ui.Warn("VM is not running.")
			return nil
		}
		ui.Step("Stopping virtiofsd")
		_ = fs.Stop(dir)
		ui.Done("virtiofsd stopped")
		ui.Step("Stopping VM")
		if err := vm.Stop(dir); err != nil {
			return err
		}
		ui.Done("VM stopped")
		ui.Success("Stopped. Run `lenv init` to restart.")
		return nil
	},
}

func init() { rootCmd.AddCommand(stopCmd) }
