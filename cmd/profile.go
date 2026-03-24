package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/AmirhoseinMasoumi/lenv/config"
	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage lenv profiles",
}

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		names := []string{"minimal", "usb", "audio", "embedded", "gpu", "full"}
		for _, name := range names {
			if p, ok := config.BuiltInProfile(name); ok {
				fmt.Printf("%s %s (built-in)\n", p.Profile.Name, p.Profile.Version)
			}
		}
		dir, err := config.ProfileDir()
		if err != nil {
			return err
		}
		entries, err := os.ReadDir(dir)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		custom := make([]string, 0, len(entries))
		for _, e := range entries {
			if e.IsDir() || filepath.Ext(e.Name()) != ".toml" {
				continue
			}
			custom = append(custom, strings.TrimSuffix(e.Name(), ".toml"))
		}
		sort.Strings(custom)
		for _, n := range custom {
			pf, err := config.LoadProfile(n)
			if err != nil {
				continue
			}
			fmt.Printf("%s %s (local)\n", pf.Profile.Name, pf.Profile.Version)
		}
		return nil
	},
}

var profileInstallCmd = &cobra.Command{
	Use:   "install <source>",
	Short: "Install a community profile from a local TOML file or GitHub URL",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		src := strings.TrimSpace(args[0])
		if src == "" {
			return fmt.Errorf("profile source cannot be empty")
		}
		dir, err := config.ProfileDir()
		if err != nil {
			return err
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		if strings.HasPrefix(src, "github.com/") {
			parts := strings.Split(src, "/")
			if len(parts) < 3 {
				return fmt.Errorf("invalid GitHub profile source %q", src)
			}
			name := strings.TrimPrefix(parts[len(parts)-1], "lenv-profile-")
			if name == "" {
				return fmt.Errorf("cannot infer profile name from %q", src)
			}
			url := "https://" + src + "/raw/main/profile.toml"
			out := filepath.Join(dir, name+".toml")
			if err := downloadAsset(url, out); err != nil {
				return fmt.Errorf("download profile: %w", err)
			}
			pf, err := config.LoadProfile(name)
			if err != nil {
				return fmt.Errorf("installed profile invalid: %w", err)
			}
			fmt.Printf("Installed profile %s %s from %s\n", pf.Profile.Name, pf.Profile.Version, src)
			return nil
		}
		if filepath.Ext(src) != ".toml" {
			return fmt.Errorf("unsupported source %q: provide a .toml file or github.com/... URL", src)
		}
		b, err := os.ReadFile(src)
		if err != nil {
			return err
		}
		base := strings.TrimSuffix(filepath.Base(src), ".toml")
		dst := filepath.Join(dir, base+".toml")
		if err := os.WriteFile(dst, b, 0o644); err != nil {
			return err
		}
		pf, err := config.LoadProfile(base)
		if err != nil {
			return fmt.Errorf("installed profile invalid: %w", err)
		}
		fmt.Printf("Installed profile %s %s from %s\n", pf.Profile.Name, pf.Profile.Version, src)
		return nil
	},
}

func completeProfiles(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	out := []string{"minimal", "usb", "audio", "embedded", "gpu", "full"}
	dir, err := config.ProfileDir()
	if err != nil {
		return out, cobra.ShellCompDirectiveNoFileComp
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return out, cobra.ShellCompDirectiveNoFileComp
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".toml" {
			continue
		}
		out = append(out, strings.TrimSuffix(e.Name(), ".toml"))
	}
	return out, cobra.ShellCompDirectiveNoFileComp
}

func init() {
	profileCmd.AddCommand(profileListCmd)
	profileCmd.AddCommand(profileInstallCmd)
	rootCmd.AddCommand(profileCmd)
}
