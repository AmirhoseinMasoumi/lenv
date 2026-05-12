package cmd

import (
	"context"
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
		ui.Title("lenv init")
		ui.KV("Project", dir)
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
		ui.KV("Distro", cfg.Distro)
		if len(selectedProfiles) > 0 {
			ui.KV("Profiles", strings.Join(selectedProfiles, ", "))
		}
		ui.Divider()
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

		sshTimeout := initSSHTimeout(cfg.Accel)
		ui.Step("Waiting for guest boot")
		markerErr := waitForGuestBoot(dir, sshTimeout)
		if markerErr != nil {
			ui.Warn("readiness marker missed, falling back to SSH probe: " + markerErr.Error())
		} else {
			ui.Done("Guest signalled ready")
		}
		ui.Step("Waiting for SSH")
		sshWait := sshTimeout
		if markerErr == nil {
			sshWait = 45 * time.Second
		}
		client, err := lssh.WaitAndConnect(port, sshWait)
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
		vmWasStopped, err := vm.EnsureBootSnapshot(dir)
		if err != nil {
			ui.Warn("snapshot save skipped: " + err.Error())
		}
		
		// Restart VM if it was stopped for snapshot
		if vmWasStopped {
			ui.Step("Restarting VM after snapshot")
			if err := vm.Start(cfg, dir, port); err != nil {
				return fmt.Errorf("restart VM after snapshot: %w", err)
			}
			ui.Done("VM restarted")
			
			restartTimeout := initSSHTimeout(cfg.Accel)
			ui.Step("Waiting for guest boot")
			markerErr2 := waitForGuestBoot(dir, restartTimeout)
			if markerErr2 != nil {
				ui.Warn("readiness marker missed, falling back to SSH probe: " + markerErr2.Error())
			} else {
				ui.Done("Guest signalled ready")
			}
			ui.Step("Waiting for SSH")
			sshWait2 := restartTimeout
			if markerErr2 == nil {
				sshWait2 = 45 * time.Second
			}
			client, err := lssh.WaitAndConnect(port, sshWait2)
			if err != nil {
				return fmt.Errorf("waiting for VM readiness after restart: %w", err)
			}
			client.Close()
			ui.Done("VM ready")
		} else {
			ui.Done("VM ready")
		}
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

// waitForGuestBoot truncates the QEMU serial log (to ignore prior boots) and
// waits for the readiness marker emitted by cloud-init / OpenRC local services.
// Returns nil on success, or an error explaining why marker-based wait was
// skipped/timed out; callers fall back to SSH probing in that case.
func waitForGuestBoot(projectDir string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return vm.WaitForReadyMarker(ctx, projectDir, timeout)
}

func initSSHTimeout(_ string) time.Duration {
	if raw := os.Getenv("LENV_SSH_WAIT_TIMEOUT_SECONDS"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			return time.Duration(n) * time.Second
		}
	}
	return 240 * time.Second
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
