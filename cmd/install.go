package cmd

import (
"fmt"
"strings"
"time"

"github.com/AmirhoseinMasoumi/Lenv/config"
lssh "github.com/AmirhoseinMasoumi/Lenv/ssh"
"github.com/AmirhoseinMasoumi/Lenv/vm"
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

