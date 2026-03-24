package cmd

import (
	"fmt"
	"os"

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
		st, err := vm.GetRuntimeStatus()
		if err != nil {
			return err
		}
		fmt.Printf("Runtime root: %s\n", st.RootDir)
		fmt.Printf("Managed dir: %s\n", st.ManagedDir)
		fmt.Printf("Managed ready: %v\n", st.ManagedReady)
		if st.QEMUPath != "" {
			fmt.Printf("qemu-system-x86_64: %s\n", st.QEMUPath)
		}
		if st.QEMUImgPath != "" {
			fmt.Printf("qemu-img: %s\n", st.QEMUImgPath)
		}
		return nil
	},
}

var runtimeVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify managed runtime integrity, completeness, and trust policy",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := vm.VerifyManagedRuntime(); err != nil {
			return err
		}
		fmt.Println("Managed runtime verified.")
		return nil
	},
}

var runtimeCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove managed runtime cache for current OS/arch",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := vm.ClearManagedRuntime(); err != nil {
			return err
		}
		fmt.Println("Managed runtime cache removed.")
		return nil
	},
}

var runtimeProvenanceCmd = &cobra.Command{
	Use:   "provenance",
	Short: "Show managed runtime source and trust policy inputs",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("runtime_url=%s\n", getenvOr("<default>", "LENV_QEMU_RUNTIME_URL"))
		fmt.Printf("runtime_sha256_url=%s\n", getenvOr("<default>", "LENV_QEMU_RUNTIME_SHA256_URL"))
		fmt.Printf("manifest_url=%s\n", getenvOr("<default>", "LENV_QEMU_RUNTIME_MANIFEST_URL"))
		fmt.Printf("manifest_sig_url=%s\n", getenvOr("<default>", "LENV_QEMU_RUNTIME_MANIFEST_SIG_URL"))
		pub := getenvOr("", "LENV_RUNTIME_MANIFEST_PUBKEY")
		if pub == "" {
			fmt.Println("manifest_pubkey=<unset>")
		} else {
			fmt.Println("manifest_pubkey=<set>")
		}
		fmt.Printf("manifest_required=%s\n", getenvOr("0", "LENV_RUNTIME_MANIFEST_REQUIRED"))
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
