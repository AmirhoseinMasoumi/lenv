package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var forwardCmd = &cobra.Command{
	Use:   "forward",
	Short: "Manage forward mappings metadata",
}

var forwardAddCmd = &cobra.Command{
	Use:   "add <host:guest>",
	Short: "Record port forward mapping",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := absProjectDir()
		if err != nil {
			return err
		}
		if !strings.Contains(args[0], ":") {
			return fmt.Errorf("mapping must be host:guest")
		}
		path := filepath.Join(dir, ".lenv", "forwards")
		f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = f.WriteString(args[0] + "\n")
		return err
	},
}

var forwardListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recorded port forwards",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := absProjectDir()
		if err != nil {
			return err
		}
		path := filepath.Join(dir, ".lenv", "forwards")
		b, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			fmt.Println("No forwards configured.")
			return nil
		}
		if err != nil {
			return err
		}
		fmt.Print(string(b))
		return nil
	},
}

var forwardStopCmd = &cobra.Command{
	Use:   "stop <host-port>",
	Short: "Remove recorded host port forward",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := absProjectDir()
		if err != nil {
			return err
		}
		path := filepath.Join(dir, ".lenv", "forwards")
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		lines := strings.Split(string(b), "\n")
		out := []string{}
		prefix := strings.TrimSpace(args[0]) + ":"
		for _, ln := range lines {
			if strings.TrimSpace(ln) == "" {
				continue
			}
			if strings.HasPrefix(strings.TrimSpace(ln), prefix) {
				continue
			}
			out = append(out, ln)
		}
		return os.WriteFile(path, []byte(strings.Join(out, "\n")+"\n"), 0o644)
	},
}

func init() {
	rootCmd.AddCommand(forwardCmd)
	forwardCmd.AddCommand(forwardAddCmd)
	forwardCmd.AddCommand(forwardListCmd)
	forwardCmd.AddCommand(forwardStopCmd)
}
