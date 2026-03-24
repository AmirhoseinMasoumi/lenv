package cmd

import (
"fmt"
"os"
"path/filepath"

"github.com/spf13/cobra"
)

var projectDir string

var rootCmd = &cobra.Command{
Use:   "lenv",
Short: "Instant, per-project Linux environments",
}

func Execute() {
if err := rootCmd.Execute(); err != nil {
fmt.Fprintln(os.Stderr, "Error:", err)
os.Exit(1)
}
}

func init() {
cwd, _ := os.Getwd()
rootCmd.PersistentFlags().StringVar(&projectDir, "project-dir", cwd, "project directory")
}

func absProjectDir() (string, error) { return filepath.Abs(projectDir) }

