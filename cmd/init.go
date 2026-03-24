package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/AmirhoseinMasoumi/lenv/config"
	"github.com/AmirhoseinMasoumi/lenv/fs"
	"github.com/AmirhoseinMasoumi/lenv/internal/ui"
	lssh "github.com/AmirhoseinMasoumi/lenv/ssh"
	"github.com/AmirhoseinMasoumi/lenv/vm"
	"github.com/spf13/cobra"
)

var noStart bool
var initDistro string
var saveConfig bool
var initProfiles []string

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Linux environment in current directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := absProjectDir()
		if err != nil {
			return fmt.Errorf("resolve project dir: %w", err)
		}
		if err := vm.EnsureState(dir); err != nil {
			return fmt.Errorf("prepare state: %w", err)
		}
		lt, err := config.Load(dir)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		if saveConfig {
			if initDistro == "" {
				initDistro = "alpine"
			}
			lt.Env.Distro = initDistro
			if len(initProfiles) > 0 {
				lt.Env.Profiles = append([]string{}, initProfiles...)
			}
			if err := config.Save(filepath.Join(dir, "lenv.toml"), lt); err != nil {
				return fmt.Errorf("save lenv.toml: %w", err)
			}
			ui.Done("Saved lenv.toml")
		}
		if initDistro != "" {
			lt.Env.Distro = initDistro
		}
		selectedProfiles := lt.Env.Profiles
		if len(initProfiles) > 0 {
			selectedProfiles = initProfiles
		}
		cfg, err := config.Resolve(lt)
		if err != nil {
			return fmt.Errorf("resolve config: %w", err)
		}
		basePackages := append([]string{}, cfg.Packages...)
		if err := config.ApplyProfiles(cfg, selectedProfiles); err != nil {
			return fmt.Errorf("apply profiles: %w", err)
		}
		if err := vm.HandleKernelProfileConfig(cfg, dir); err != nil {
			return fmt.Errorf("apply profile kernel config: %w", err)
		}
		if err := vm.EnsureDisk(cfg, dir); err != nil {
			return fmt.Errorf("prepare rootfs disk: %w", err)
		}
		if err := vm.RestoreBootSnapshot(dir); err != nil {
			ui.Warn("snapshot restore skipped: " + err.Error())
		}
		vm.ResolveKernelPath(cfg, dir)
		if err := config.WriteResolved(vm.ConfigPath(dir), cfg); err != nil {
			return fmt.Errorf("write resolved config: %w", err)
		}
		st := vm.GetStatus(dir)
		if st.Running && st.SSHPort > 0 {
			if client, err := lssh.WaitAndConnect(st.SSHPort, 10*time.Second); err == nil {
				_ = client.Close()
				ui.Success("Already running. Use `lenv shell` or `lenv run <cmd>`")
				return nil
			}
		}
		port, err := vm.EnsurePort(dir)
		if err != nil {
			return fmt.Errorf("prepare ssh port: %w", err)
		}
		if noStart {
			ui.Success("Initialized .lenv state")
			return nil
		}

		useVirtioFS := fs.Available() || runtime.GOOS != "windows"
		if useVirtioFS {
			if err := fs.CheckInstalled(); err != nil {
				return err
			}
		}
		if useVirtioFS {
			ui.Step("Starting virtiofsd")
			if err := fs.Start(dir); err != nil {
				return fmt.Errorf("start virtiofsd: %w", err)
			}
			ui.Done("virtiofsd running")
		} else {
			ui.Warn("virtiofsd unavailable; continuing without host shared-folder integration")
		}

		ui.Step("Starting QEMU")
		if err := vm.Start(cfg, dir, port); err != nil {
			return fmt.Errorf("starting VM: %w", err)
		}
		ui.Done("VM started")

		ui.Step("Waiting for SSH")
		sshTimeout := initSSHTimeout(cfg.Accel)
		client, err := lssh.WaitAndConnect(port, sshTimeout)
		if err != nil {
			return fmt.Errorf("waiting for VM readiness: %w", err)
		}
		defer client.Close()
		profilePackages := diffPackages(cfg.Packages, basePackages)
		if len(profilePackages) > 0 {
			pkgs := strings.Join(profilePackages, " ")
			installLine := packageInstallCommand(cfg.PkgManager, pkgs)
			exitCode, err := lssh.Exec(client, installLine)
			if err != nil {
				return fmt.Errorf("apply profile packages: %w", err)
			}
			if exitCode != 0 {
				return fmt.Errorf("profile package install failed with exit code %d", exitCode)
			}
			cfg.InstalledPackages = mergePackages(cfg.InstalledPackages, profilePackages)
			if err := config.WriteResolved(vm.ConfigPath(dir), cfg); err != nil {
				return fmt.Errorf("persist resolved config: %w", err)
			}
		}
		if err := vm.EnsureBootSnapshot(dir); err != nil {
			ui.Warn("snapshot save skipped: " + err.Error())
		}
		ui.Done("VM ready")
		ui.Success("Ready. Run `lenv shell` or `lenv run <cmd>`")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVar(&noStart, "no-start", false, "only prepare .lenv metadata and config")
	initCmd.Flags().StringVar(&initDistro, "distro", "", "override distro for this init (alpine|ubuntu|debian|arch)")
	initCmd.Flags().BoolVar(&saveConfig, "save", false, "write selected settings to lenv.toml")
	initCmd.Flags().StringArrayVar(&initProfiles, "profile", nil, "activate profile(s), e.g. --profile usb --profile audio")
}

func initSSHTimeout(_ string) time.Duration {
	if raw := os.Getenv("LENV_SSH_WAIT_TIMEOUT_SECONDS"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			return time.Duration(n) * time.Second
		}
	}
	return 120 * time.Second
}

func diffPackages(all, base []string) []string {
	baseSet := map[string]bool{}
	for _, p := range base {
		baseSet[strings.TrimSpace(p)] = true
	}
	out := []string{}
	for _, p := range all {
		p = strings.TrimSpace(p)
		if p == "" || baseSet[p] {
			continue
		}
		out = append(out, p)
	}
	return out
}
