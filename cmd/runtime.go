package cmd

import (
	"fmt"
	"os"

	"github.com/AmirhoseinMasoumi/lenv/internal/ui"
	"github.com/AmirhoseinMasoumi/lenv/vm"
	"github.com/spf13/cobra"
)

var runtimeCmd = &cobra.Command{
	Use:   "runtime",
	Short: "Manage managed runtime assets",
}

var runtimeStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show managed runtime status",
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.Title("lenv runtime status")
		st, err := vm.GetRuntimeStatus()
		if err != nil {
			return err
		}
		ui.KV("Runtime root", st.RootDir)
		ui.KV("Managed dir", st.ManagedDir)
		ui.KV("Managed ready", fmt.Sprintf("%v", st.ManagedReady))
		if st.QEMUPath != "" {
			ui.KV("qemu-system-x86_64", st.QEMUPath)
		}
		if st.QEMUImgPath != "" {
			ui.KV("qemu-img", st.QEMUImgPath)
		}
		ui.Divider()
		return nil
	},
}

var runtimeVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify managed runtime integrity, completeness, and trust policy",
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.Title("lenv runtime verify")
		if err := vm.VerifyManagedRuntime(); err != nil {
			return err
		}
		ui.Success("Managed runtime verified.")
		return nil
	},
}

var runtimeCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove managed runtime cache for current OS/arch",
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.Title("lenv runtime clean")
		if err := vm.ClearManagedRuntime(); err != nil {
			return err
		}
		ui.Success("Managed runtime cache removed.")
		return nil
	},
}

var runtimeProvenanceCmd = &cobra.Command{
	Use:   "provenance",
	Short: "Show managed runtime source and trust policy inputs",
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.Title("lenv runtime provenance")
		ui.KV("runtime_url", getenvOr("<default>", "LENV_QEMU_RUNTIME_URL"))
		ui.KV("runtime_sha256_url", getenvOr("<default>", "LENV_QEMU_RUNTIME_SHA256_URL"))
		ui.KV("manifest_url", getenvOr("<default>", "LENV_QEMU_RUNTIME_MANIFEST_URL"))
		ui.KV("manifest_sig_url", getenvOr("<default>", "LENV_QEMU_RUNTIME_MANIFEST_SIG_URL"))
		pub := getenvOr("", "LENV_RUNTIME_MANIFEST_PUBKEY")
		if pub == "" {
			ui.KV("manifest_pubkey", "<unset>")
		} else {
			ui.KV("manifest_pubkey", "<set>")
		}
		ui.KV("manifest_required", getenvOr("0", "LENV_RUNTIME_MANIFEST_REQUIRED"))
		ui.Divider()
		return nil
	},
}

func init() {
	runtimeCmd.AddCommand(runtimeStatusCmd)
	runtimeCmd.AddCommand(runtimeVerifyCmd)
	runtimeCmd.AddCommand(runtimeCleanCmd)
	runtimeCmd.AddCommand(runtimeProvenanceCmd)
	rootCmd.AddCommand(runtimeCmd)
}

func getenvOr(def, key string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
