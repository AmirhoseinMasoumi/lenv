package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AmirhoseinMasoumi/lenv/internal/logger"
	"github.com/AmirhoseinMasoumi/lenv/internal/ui"
	"github.com/spf13/cobra"
)

// Version information (set via ldflags during build)
var (
	Version   = "dev"
	BuildTime = "unknown"
)

var projectDir string
var plainOutput bool
var compactOutput bool
var verboseOutput bool

var rootCmd = &cobra.Command{
	Use:   "lenv",
	Short: "Instant, per-project Linux environments",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		ui.Configure(plainOutput, compactOutput)
		// Set log level based on verbose flag
		if verboseOutput {
			os.Setenv("LENV_LOG_LEVEL", "DEBUG")
		}
		// Initialize logger with .lenv directory
		stateDir := filepath.Join(projectDir, ".lenv")
		if err := logger.Init(stateDir); err != nil {
			// Non-fatal: continue without file logging
			fmt.Fprintf(os.Stderr, "Warning: logger init failed: %v\n", err)
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		logger.Close()
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
	rootCmd.PersistentFlags().BoolVarP(&verboseOutput, "verbose", "v", false, "enable verbose debug output")

	// Add version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("lenv version %s\n", Version)
			fmt.Printf("Built: %s\n", BuildTime)
		},
	})
}

func absProjectDir() (string, error) { return filepath.Abs(projectDir) }
