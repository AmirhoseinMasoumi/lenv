package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/AmirhoseinMasoumi/lenv/vm"
	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return writeCompletion(os.Stdout, args[0])
	},
}

func writeCompletion(w io.Writer, shell string) error {
	switch shell {
	case "bash":
		return rootCmd.GenBashCompletion(w)
	case "zsh":
		return rootCmd.GenZshCompletion(w)
	case "fish":
		return rootCmd.GenFishCompletion(w, true)
	case "powershell":
		return rootCmd.GenPowerShellCompletionWithDesc(w)
	default:
		return fmt.Errorf("unsupported shell %q", shell)
	}
}

func completeDistros(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return []string{"alpine", "ubuntu", "debian", "arch"}, cobra.ShellCompDirectiveNoFileComp
}

func completeSnapshots(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	dir := home + string(os.PathSeparator) + ".lenv" + string(os.PathSeparator) + "snapshots"
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	out := []string{}
	for _, e := range ents {
		n := e.Name()
		if len(n) > 6 && n[len(n)-6:] == ".qcow2" {
			out = append(out, n[:len(n)-6])
		}
	}
	return out, cobra.ShellCompDirectiveNoFileComp
}

func completeInstances(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	st, err := vm.ListRunningStatuses()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	out := make([]string, 0, len(st))
	for _, s := range st {
		out = append(out, s.Instance)
	}
	return out, cobra.ShellCompDirectiveNoFileComp
}

func init() {
	rootCmd.AddCommand(completionCmd)
	_ = initCmd.RegisterFlagCompletionFunc("distro", completeDistros)
	_ = snapshotRestoreCmd.ValidArgsFunction
	_ = snapshotDeleteCmd.ValidArgsFunction
	snapshotRestoreCmd.ValidArgsFunction = completeSnapshots
	snapshotDeleteCmd.ValidArgsFunction = completeSnapshots
}
