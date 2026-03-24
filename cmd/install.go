package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/AmirhoseinMasoumi/lenv/config"
	lssh "github.com/AmirhoseinMasoumi/lenv/ssh"
	"github.com/AmirhoseinMasoumi/lenv/vm"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install <packages...>",
	Short: "Install packages inside VM",
	Args:  cobra.MinimumNArgs(1),
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
		st := vm.GetStatus(dir)
		if st.SSHPort == 0 {
			return fmt.Errorf("VM not initialized; run `lenv init` first")
		}
		if !st.Running {
			return fmt.Errorf("VM is not running; run `lenv init` first")
		}
		client, err := lssh.WaitAndConnect(st.SSHPort, 30*time.Second)
		if err != nil {
			return err
		}
		defer client.Close()
		pkgs := strings.Join(args, " ")
		installLine := packageInstallCommand(cfg.PkgManager, pkgs)
		exitCode, err := lssh.Exec(client, installLine)
		if err != nil {
			return err
		}
		if exitCode != 0 {
			return fmt.Errorf("package install failed with exit code %d", exitCode)
		}
		cfg.InstalledPackages = mergePackages(cfg.InstalledPackages, args)
		if err := config.WriteResolved(vm.ConfigPath(dir), cfg); err != nil {
			// Don't fail the command, but warn the user that config wasn't persisted
			fmt.Fprintf(os.Stderr, "warning: failed to persist package list: %v\n", err)
		}
		return nil
	},
}

func packageInstallCommand(manager, pkgs string) string {
	switch manager {
	case "apk":
		return "apk add --no-cache " + pkgs
	case "apt":
		return "apt update && apt install -y " + pkgs
	case "pacman":
		return "pacman -Sy --noconfirm --needed " + pkgs
	default:
		return "echo unsupported package manager && false"
	}
}

func init() { rootCmd.AddCommand(installCmd) }

func mergePackages(existing, added []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, p := range existing {
		p = strings.TrimSpace(p)
		if p == "" || seen[p] {
			continue
		}
		seen[p] = true
		out = append(out, p)
	}
	for _, p := range added {
		p = strings.TrimSpace(p)
		if p == "" || seen[p] {
			continue
		}
		seen[p] = true
		out = append(out, p)
	}
	return out
}
