package cmd

import (
	"fmt"

	"github.com/AmirhoseinMasoumi/lenv/config"
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
		dir, err := absProjectDir()
		if err != nil {
			return err
		}
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
		if err := vm.RebuildKernelProfileConfig(cfg, dir); err != nil {
			return err
		}
		fmt.Println("Kernel profile configuration applied.")
		return nil
	},
}

func init() {
	kernelCmd.AddCommand(kernelRebuildCmd)
	rootCmd.AddCommand(kernelCmd)
}
