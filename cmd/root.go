package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AmirhoseinMasoumi/lenv/internal/ui"
	"github.com/spf13/cobra"
)

var projectDir string
var plainOutput bool
var compactOutput bool

var rootCmd = &cobra.Command{
	Use:   "lenv",
	Short: "Instant, per-project Linux environments",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		ui.Configure(plainOutput, compactOutput)
	},
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
	rootCmd.PersistentFlags().BoolVar(&plainOutput, "plain", false, "disable ANSI visuals and use script-friendly output")
	rootCmd.PersistentFlags().BoolVar(&compactOutput, "compact", false, "use compact key=value output mode")
}

func absProjectDir() (string, error) { return filepath.Abs(projectDir) }
