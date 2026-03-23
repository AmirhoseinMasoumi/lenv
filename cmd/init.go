package cmd

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

"github.com/AmirhoseinMasoumi/lenv/config"
"github.com/AmirhoseinMasoumi/lenv/fs"
"github.com/AmirhoseinMasoumi/lenv/internal/ui"
lssh "github.com/AmirhoseinMasoumi/lenv/ssh"
"github.com/AmirhoseinMasoumi/lenv/vm"
"github.com/spf13/cobra"
)

var noStart bool

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
	cfg, err := config.Resolve(lt)
	if err != nil {
		return fmt.Errorf("resolve config: %w", err)
	}
	vm.ResolveKernelPath(cfg, dir)
	if err := config.WriteResolved(vm.ConfigPath(dir), cfg); err != nil {
		return fmt.Errorf("write resolved config: %w", err)
	}
port, err := vm.EnsurePort(dir)
if err != nil {
return fmt.Errorf("prepare ssh port: %w", err)
}
if noStart {
ui.Success("Initialized .lenv state")
return nil
}

	if err := fs.CheckInstalled(); err != nil {
		return err
	}
	useVirtioFS := fs.Available() || runtime.GOOS != "windows"
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
_ = client.Close()
ui.Done("VM ready")
ui.Success("Ready. Run `lenv shell` or `lenv run <cmd>`")
return nil
},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVar(&noStart, "no-start", false, "only prepare .lenv metadata and config")
}

func initSSHTimeout(_ string) time.Duration {
	if raw := os.Getenv("LENV_SSH_WAIT_TIMEOUT_SECONDS"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			return time.Duration(n) * time.Second
		}
	}
	return 120 * time.Second
}
