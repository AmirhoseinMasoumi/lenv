package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AmirhoseinMasoumi/lenv/vm"
	"github.com/spf13/cobra"
)

var vscodeCmd = &cobra.Command{
	Use:   "vscode",
	Short: "Write VS Code Remote SSH settings for the running VM",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := absProjectDir()
		if err != nil {
			return err
		}
		st := vm.GetStatus(dir)
		if !st.Running || st.SSHPort == 0 {
			return fmt.Errorf("VM is not running; run `lenv init` first")
		}
		settings := map[string]any{
			"remote.SSH.configFile": filepath.Join(dir, ".vscode", "ssh_config"),
		}
		vsdir := filepath.Join(dir, ".vscode")
		if err := os.MkdirAll(vsdir, 0o755); err != nil {
			return err
		}
		b, _ := json.MarshalIndent(settings, "", "  ")
		if err := os.WriteFile(filepath.Join(vsdir, "settings.json"), b, 0o644); err != nil {
			return err
		}
		sshCfg := fmt.Sprintf("Host lenv-%s\n  HostName 127.0.0.1\n  Port %d\n  User root\n", st.Instance, st.SSHPort)
		if err := os.WriteFile(filepath.Join(vsdir, "ssh_config"), []byte(sshCfg), 0o644); err != nil {
			return err
		}
		fmt.Println("Wrote .vscode/settings.json and .vscode/ssh_config")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(vscodeCmd)
}
