package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AmirhoseinMasoumi/lenv/config"
	"github.com/spf13/cobra"
)

var provenanceCmd = &cobra.Command{
	Use:   "provenance",
	Short: "Show provenance for runtime and installed profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("[runtime]")
		_ = runtimeProvenanceCmd.RunE(cmd, args)
		fmt.Println()
		fmt.Println("[profiles]")
		dir, err := config.ProfileDir()
		if err != nil {
			return err
		}
		entries, err := os.ReadDir(dir)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if os.IsNotExist(err) || len(entries) == 0 {
			fmt.Println("none")
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
			fmt.Printf("%s: %s\n", base, source)
		}
		fmt.Println()
		fmt.Println("[profile_trust_catalog]")
		catalog := filepath.Join(dir, "trusted-sources.txt")
		fmt.Println("built-in:")
		for prefix := range trustedProfileSources {
			fmt.Printf("- %s\n", prefix)
		}
		fmt.Println("local:")
		if _, err := os.Stat(catalog); err != nil {
			fmt.Println("none")
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
			fmt.Println(line)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(provenanceCmd)
}
