package cmd

import (
"fmt"
"io"
"os"
"path/filepath"

"github.com/AmirhoseinMasoumi/lenv/vm"
"github.com/spf13/cobra"
)

var snapshotCmd = &cobra.Command{Use: "snapshot", Short: "Save/restore VM snapshots"}

var snapshotSaveCmd = &cobra.Command{
Use:   "save <name>",
Short: "Save snapshot",
Args:  cobra.ExactArgs(1),
RunE: func(cmd *cobra.Command, args []string) error {
dir, _ := absProjectDir()
home, err := os.UserHomeDir()
if err != nil {
return err
}
if err := os.MkdirAll(filepath.Join(home, ".lenv", "snapshots"), 0o755); err != nil {
return err
}
src := vm.DiskPath(dir)
dst := filepath.Join(home, ".lenv", "snapshots", args[0]+".qcow2")
return copyFile(src, dst)
},
}

var snapshotRestoreCmd = &cobra.Command{
Use:   "restore <name>",
Short: "Restore snapshot",
Args:  cobra.ExactArgs(1),
RunE: func(cmd *cobra.Command, args []string) error {
dir, _ := absProjectDir()
home, err := os.UserHomeDir()
if err != nil {
return err
}
src := filepath.Join(home, ".lenv", "snapshots", args[0]+".qcow2")
dst := vm.DiskPath(dir)
if _, err := os.Stat(src); err != nil {
return fmt.Errorf("snapshot not found: %s", args[0])
}
return copyFile(src, dst)
},
}

func copyFile(src, dst string) error {
in, err := os.Open(src)
if err != nil { return err }
defer in.Close()
if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil { return err }
out, err := os.Create(dst)
if err != nil { return err }
defer out.Close()
if _, err := io.Copy(out, in); err != nil { return err }
return out.Close()
}

func init() {
snapshotCmd.AddCommand(snapshotSaveCmd)
snapshotCmd.AddCommand(snapshotRestoreCmd)
rootCmd.AddCommand(snapshotCmd)
}
