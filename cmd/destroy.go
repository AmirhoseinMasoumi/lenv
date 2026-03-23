package cmd

import (
"fmt"
"os"

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
if err := os.RemoveAll(vm.StateDir(dir)); err != nil {
return fmt.Errorf("remove .lenv: %w", err)
}
ui.Success("Destroyed .lenv state")
return nil
},
}

func init() { rootCmd.AddCommand(destroyCmd) }
