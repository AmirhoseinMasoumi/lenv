package cmd

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/AmirhoseinMasoumi/lenv/config"
	"github.com/AmirhoseinMasoumi/lenv/internal/ui"
	"github.com/spf13/cobra"
)

var provenanceCmd = &cobra.Command{
	Use:   "provenance",
	Short: "Show provenance for runtime and installed profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.Title("lenv provenance")
		ui.KV("Section", "runtime")
		ui.Divider()
		_ = runtimeProvenanceCmd.RunE(cmd, args)
		ui.Divider()
		ui.KV("Section", "profiles")
		dir, err := config.ProfileDir()
		if err != nil {
			return err
		}
		entries, err := os.ReadDir(dir)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if os.IsNotExist(err) || len(entries) == 0 {
			ui.Warn("No installed profiles.")
			return nil
		}
		for _, e := range entries {
			if e.IsDir() || filepath.Ext(e.Name()) != ".toml" {
				continue
			}
			base := strings.TrimSuffix(e.Name(), ".toml")
			sourcePath := filepath.Join(dir, e.Name()+".source")
			source := "<unknown>"
			if b, err := os.ReadFile(sourcePath); err == nil {
				source = strings.TrimSpace(string(b))
			}
			ui.KV(base, source)
		}
		ui.Divider()
		ui.KV("Section", "profile_trust_catalog")
		catalog := filepath.Join(dir, "trusted-sources.txt")
		ui.KV("Built-in", "")
		for prefix := range trustedProfileSources {
			ui.KV("Trusted source", prefix)
		}
		ui.KV("Local", "")
		if _, err := os.Stat(catalog); err != nil {
			ui.Info("none")
			return nil
		}
		f, err := os.Open(catalog)
		if err != nil {
			return err
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			ui.KV("Trusted source", line)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(provenanceCmd)
}
