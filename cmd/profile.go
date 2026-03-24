package cmd

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
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

var profileTrustCmd = &cobra.Command{
	Use:   "trust",
	Short: "Manage trusted profile source catalog",
}

var profileTrustListCmd = &cobra.Command{
	Use:   "list",
	Short: "List trusted profile source prefixes",
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, p := range mergeTrustedSources(readTrustedSourceCatalog()) {
			fmt.Println(p)
		}
		return nil
	},
}

var profileTrustAddCmd = &cobra.Command{
	Use:   "add <prefix>",
	Short: "Add trusted profile source prefix",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		prefix := strings.TrimSpace(args[0])
		if prefix == "" {
			return fmt.Errorf("trusted source prefix cannot be empty")
		}
		items := readTrustedSourceCatalog()
		for _, v := range items {
			if v == prefix {
				fmt.Println("Trusted source already present.")
				return nil
			}
		}
		items = append(items, prefix)
		if err := writeTrustedSourceCatalog(items); err != nil {
			return err
		}
		fmt.Printf("Added trusted source %s\n", prefix)
		return nil
	},
}

var profileTrustRemoveCmd = &cobra.Command{
	Use:   "remove <prefix>",
	Short: "Remove trusted profile source prefix from local catalog",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		prefix := strings.TrimSpace(args[0])
		if prefix == "" {
			return fmt.Errorf("trusted source prefix cannot be empty")
		}
		items := readTrustedSourceCatalog()
		next := make([]string, 0, len(items))
		removed := false
		for _, v := range items {
			if v == prefix {
				removed = true
				continue
			}
			next = append(next, v)
		}
		if !removed {
			return fmt.Errorf("trusted source %q not found in local catalog", prefix)
		}
		if err := writeTrustedSourceCatalog(next); err != nil {
			return err
		}
		fmt.Printf("Removed trusted source %s\n", prefix)
		return nil
	},
}

var trustedProfileSources = map[string]bool{
	"github.com/AmirhoseinMasoumi/": true,
	"github.com/someone/":           true,
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
			if !isTrustedProfileSource(src) {
				return fmt.Errorf("profile source %q is not trusted by policy; set LENV_PROFILE_TRUST_MODE=permissive to allow", src)
			}
			parts := strings.Split(src, "/")
			if len(parts) < 3 {
				return fmt.Errorf("invalid GitHub profile source %q", src)
			}
			name := strings.TrimPrefix(parts[len(parts)-1], "lenv-profile-")
			if name == "" {
				return fmt.Errorf("cannot infer profile name from %q", src)
			}
			url := "https://" + src + "/raw/main/profile.toml"
			sumURL := "https://" + src + "/raw/main/profile.toml.sha256"
			out := filepath.Join(dir, name+".toml")
			if err := downloadAsset(url, out); err != nil {
				return fmt.Errorf("download profile: %w", err)
			}
			if err := writeProfileChecksumFile(out, sumURL); err != nil {
				return err
			}
			if err := writeProfileProvenanceFile(out, "github", src); err != nil {
				return err
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
		if err := writeLocalProfileChecksumFile(dst); err != nil {
			return err
		}
		if err := writeProfileProvenanceFile(dst, "local", src); err != nil {
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

var profileRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove an installed local/community profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := strings.TrimSpace(args[0])
		if name == "" {
			return fmt.Errorf("profile name cannot be empty")
		}
		if _, ok := config.BuiltInProfile(name); ok {
			return fmt.Errorf("cannot remove built-in profile %q", name)
		}
		dir, err := config.ProfileDir()
		if err != nil {
			return err
		}
		profilePath := filepath.Join(dir, name+".toml")
		if err := os.Remove(profilePath); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("profile %q not found", name)
			}
			return err
		}
		_ = os.Remove(profilePath + ".sha256")
		fmt.Printf("Removed profile %s\n", name)
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
	profileCmd.AddCommand(profileRemoveCmd)
	profileTrustCmd.AddCommand(profileTrustListCmd)
	profileTrustCmd.AddCommand(profileTrustAddCmd)
	profileTrustCmd.AddCommand(profileTrustRemoveCmd)
	profileCmd.AddCommand(profileTrustCmd)
	rootCmd.AddCommand(profileCmd)
}

func writeProfileChecksumFile(profilePath, checksumURL string) error {
	resp, err := http.Get(checksumURL)
	if err != nil {
		return fmt.Errorf("fetch profile checksum: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("fetch profile checksum failed: %s", resp.Status)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fields := strings.Fields(string(b))
	if len(fields) == 0 {
		return fmt.Errorf("invalid profile checksum payload")
	}
	return os.WriteFile(profilePath+".sha256", []byte(fields[0]+"\n"), 0o644)
}

func writeLocalProfileChecksumFile(profilePath string) error {
	b, err := os.ReadFile(profilePath)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(b)
	return os.WriteFile(profilePath+".sha256", []byte(hex.EncodeToString(sum[:])+"\n"), 0o644)
}

func writeProfileProvenanceFile(profilePath, kind, source string) error {
	body := "kind=" + kind + "\nsource=" + source + "\n"
	return os.WriteFile(profilePath+".source", []byte(body), 0o644)
}

func isTrustedProfileSource(src string) bool {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("LENV_PROFILE_TRUST_MODE")), "permissive") {
		return true
	}
	for _, prefix := range mergeTrustedSources(readTrustedSourceCatalog()) {
		if strings.HasPrefix(src, prefix) {
			return true
		}
	}
	return false
}

func mergeTrustedSources(local []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for prefix := range trustedProfileSources {
		if strings.TrimSpace(prefix) == "" || seen[prefix] {
			continue
		}
		seen[prefix] = true
		out = append(out, prefix)
	}
	for _, prefix := range local {
		prefix = strings.TrimSpace(prefix)
		if prefix == "" || seen[prefix] {
			continue
		}
		seen[prefix] = true
		out = append(out, prefix)
	}
	sort.Strings(out)
	return out
}

func readTrustedSourceCatalog() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	path := filepath.Join(home, ".lenv", "profiles", "trusted-sources.txt")
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	out := []string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out = append(out, line)
	}
	return out
}

func writeTrustedSourceCatalog(items []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, ".lenv", "profiles")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, "trusted-sources.txt")
	normalized := []string{}
	for _, v := range items {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		normalized = append(normalized, v)
	}
	sort.Strings(normalized)
	body := strings.Join(normalized, "\n")
	if body != "" {
		body += "\n"
	}
	return os.WriteFile(path, []byte(body), 0o644)
}
