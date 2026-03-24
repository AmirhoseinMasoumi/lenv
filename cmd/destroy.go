package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/AmirhoseinMasoumi/lenv/fs"
	"github.com/AmirhoseinMasoumi/lenv/internal/ui"
	"github.com/AmirhoseinMasoumi/lenv/vm"
	"github.com/spf13/cobra"
)

var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy VM state for current project",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := absProjectDir()
		if err != nil {
			return err
		}
		_ = fs.Stop(dir)
		_ = vm.Stop(dir)
		var rmErr error
		for i := 0; i < 20; i++ {
			rmErr = os.RemoveAll(vm.StateDir(dir))
			if rmErr == nil || os.IsNotExist(rmErr) {
				rmErr = nil
				break
			}
			time.Sleep(150 * time.Millisecond)
		}
		if rmErr != nil {
			return fmt.Errorf("remove .lenv: %w", rmErr)
		}
		_ = vm.RemoveInstance(dir)
		ui.Success("Destroyed .lenv state")
		return nil
	},
}

func init() { rootCmd.AddCommand(destroyCmd) }

