package cmd

import (
	"fmt"

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
	Short: "Verify managed runtime integrity and completeness",
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

func init() {
	runtimeCmd.AddCommand(runtimeStatusCmd)
	runtimeCmd.AddCommand(runtimeVerifyCmd)
	runtimeCmd.AddCommand(runtimeCleanCmd)
	rootCmd.AddCommand(runtimeCmd)
}
