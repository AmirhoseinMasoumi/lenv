package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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

var snapshotListCmd = &cobra.Command{
	Use:   "list",
	Short: "List snapshots",
	RunE: func(cmd *cobra.Command, args []string) error {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		dir := filepath.Join(home, ".lenv", "snapshots")
		entries, err := os.ReadDir(dir)
		if os.IsNotExist(err) {
			fmt.Println("No snapshots found.")
			return nil
		}
		if err != nil {
			return err
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			if filepath.Ext(e.Name()) != ".qcow2" {
				continue
			}
			fmt.Println(strings.TrimSuffix(e.Name(), ".qcow2"))
		}
		return nil
	},
}

var snapshotDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete snapshot",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		path := filepath.Join(home, ".lenv", "snapshots", args[0]+".qcow2")
		if err := os.Remove(path); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("snapshot not found: %s", args[0])
			}
			return err
		}
		fmt.Printf("Deleted snapshot %s\n", args[0])
		return nil
	},
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

func init() {
	snapshotCmd.AddCommand(snapshotSaveCmd)
	snapshotCmd.AddCommand(snapshotRestoreCmd)
	snapshotCmd.AddCommand(snapshotListCmd)
	snapshotCmd.AddCommand(snapshotDeleteCmd)
	rootCmd.AddCommand(snapshotCmd)
}
