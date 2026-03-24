package cmd

import (
	"fmt"
	"strings"

	"github.com/AmirhoseinMasoumi/lenv/config"
	"github.com/AmirhoseinMasoumi/lenv/internal/ui"
	"github.com/AmirhoseinMasoumi/lenv/vm"
	"github.com/spf13/cobra"
)

var kernelCmd = &cobra.Command{
	Use:   "kernel",
	Short: "Kernel profile configuration commands",
}

var kernelRebuildCmd = &cobra.Command{
	Use:   "rebuild",
	Short: "Rebuild/apply kernel profile configuration using configured build command",
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.Title("lenv kernel rebuild")
		dir, err := absProjectDir()
		if err != nil {
			return err
		}
		ui.KV("Project", dir)
		lt, err := config.Load(dir)
		if err != nil {
			return err
		}
		cfg, err := config.Resolve(lt)
		if err != nil {
			return err
		}
		if err := config.ApplyProfiles(cfg, cfg.Profiles); err != nil {
			return err
		}
		ui.KV("Profiles", strings.Join(cfg.Profiles, ", "))
		ui.KV("Kernel config entries", fmt.Sprintf("%d", len(cfg.KernelConfig)))
		ui.Divider()
		if err := vm.RebuildKernelProfileConfig(cfg, dir); err != nil {
			return err
		}
		ui.Success("Kernel profile configuration applied.")
		return nil
	},
}

func init() {
	kernelCmd.AddCommand(kernelRebuildCmd)
	rootCmd.AddCommand(kernelCmd)
}
