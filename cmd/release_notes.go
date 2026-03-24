package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var releaseVersion string

var releaseNotesCmd = &cobra.Command{
	Use:   "release-notes",
	Short: "Generate a release notes template",
	RunE: func(cmd *cobra.Command, args []string) error {
		v := strings.TrimSpace(releaseVersion)
		if v == "" {
			v = "vNEXT"
		}
		fmt.Printf("## lenv %s\n\n", v)
		fmt.Println("### What's new")
		fmt.Println("- Runtime and trust policy improvements")
		fmt.Println("- Profile ecosystem and provenance enhancements")
		fmt.Println("- Kernel profile pipeline updates")
		fmt.Println()
		fmt.Println("### Install")
		fmt.Println("```bash")
		fmt.Println("go install github.com/AmirhoseinMasoumi/lenv@latest")
		fmt.Println("```")
		fmt.Println()
		fmt.Println("### Validation")
		fmt.Println("- go test ./...")
		fmt.Println("- go build ./...")
		fmt.Println("- go vet ./...")
		return nil
	},
}

func init() {
	releaseNotesCmd.Flags().StringVar(&releaseVersion, "version", "", "release version label")
	rootCmd.AddCommand(releaseNotesCmd)
}
