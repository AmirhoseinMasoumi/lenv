package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

const currentVersion = "v0.4.0-dev"

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for updates and replace binary in-place when available",
	RunE: func(cmd *cobra.Command, args []string) error {
		release, err := latestRelease()
		if err != nil {
			return err
		}
		if release.TagName == "" || release.TagName == currentVersion {
			fmt.Println("Already up to date.")
			return nil
		}
		fmt.Printf("New version available: %s\n", release.TagName)
		asset := pickAsset(release.Assets)
		if asset.URL == "" {
			return fmt.Errorf("no compatible binary asset found for %s/%s", runtime.GOOS, runtime.GOARCH)
		}
		exe, err := os.Executable()
		if err != nil {
			return err
		}
		tmp := exe + ".new"
		if err := downloadAsset(asset.URL, tmp); err != nil {
			return err
		}
		if err := os.Chmod(tmp, 0o755); err != nil && runtime.GOOS != "windows" {
			return err
		}
		backup := exe + ".bak"
		_ = os.Remove(backup)
		_ = os.Rename(exe, backup)
		if err := os.Rename(tmp, exe); err != nil {
			_ = os.Rename(backup, exe)
			return err
		}
		fmt.Printf("Updated to %s\n", release.TagName)
		return nil
	},
}

type ghRelease struct {
	TagName string    `json:"tag_name"`
	Assets  []ghAsset `json:"assets"`
}
type ghAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	URL                string
}

func latestRelease() (*ghRelease, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.github.com/repos/AmirhoseinMasoumi/lenv/releases/latest", nil)
	if err != nil {
		return nil, fmt.Errorf("create release request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("release API failed: %s", resp.Status)
	}
	var rel ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}
	for i := range rel.Assets {
		rel.Assets[i].URL = rel.Assets[i].BrowserDownloadURL
	}
	return &rel, nil
}

func pickAsset(assets []ghAsset) ghAsset {
	key := runtime.GOOS + "_" + runtime.GOARCH
	for _, a := range assets {
		if filepath.Ext(a.Name) == ".zip" || filepath.Ext(a.Name) == ".tar.gz" {
			continue
		}
		if containsAll(a.Name, []string{runtime.GOOS, runtime.GOARCH}) || containsAll(a.Name, []string{key}) {
			return a
		}
	}
	return ghAsset{}
}

func containsAll(s string, parts []string) bool {
	for _, p := range parts {
		if p == "" {
			continue
		}
		if !containsFold(s, p) {
			return false
		}
	}
	return true
}

func containsFold(s, sub string) bool {
	return len(s) >= len(sub) && (indexFold(s, sub) >= 0)
}

func indexFold(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if equalFoldASCII(s[i:i+len(sub)], sub) {
			return i
		}
	}
	return -1
}

func equalFoldASCII(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}

func downloadAsset(url, out string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("download failed: %s", resp.Status)
	}
	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
