package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/AmirhoseinMasoumi/lenv/config"
	"github.com/AmirhoseinMasoumi/lenv/fs"
	"github.com/AmirhoseinMasoumi/lenv/internal/ui"
	lssh "github.com/AmirhoseinMasoumi/lenv/ssh"
	"github.com/AmirhoseinMasoumi/lenv/vm"
	"github.com/spf13/cobra"
)

var benchmarkRuns int

var benchmarkCmd = &cobra.Command{
	Use:   "benchmark",
	Short: "Benchmark init/run/destroy cycle",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := absProjectDir()
		if err != nil {
			return err
		}
		if benchmarkRuns <= 0 {
			benchmarkRuns = 3
		}

		var bootSum, sshSum, rttSum time.Duration
		for i := 0; i < benchmarkRuns; i++ {
			_ = fs.Stop(dir)
			_ = vm.Stop(dir)
			_ = vm.RemoveInstance(dir)
			_ = vm.EnsureState(dir)
			lt, err := config.Load(dir)
			if err != nil {
				return err
			}
			cfg, err := config.Resolve(lt)
			if err != nil {
				return err
			}
			if err := vm.EnsureDisk(cfg, dir); err != nil {
				return err
			}
			_ = vm.RestoreBootSnapshot(dir)
			vm.ResolveKernelPath(cfg, dir)
			port, err := vm.EnsurePort(dir)
			if err != nil {
				return err
			}
			useVirtioFS := fs.Available()
			if useVirtioFS {
				_ = fs.CheckInstalled()
				_ = fs.Start(dir)
			}

			t0 := time.Now()
			if err := vm.Start(cfg, dir, port); err != nil {
				return err
			}
			bootDur := time.Since(t0)
			bootSum += bootDur

			t1 := time.Now()
			client, err := lssh.WaitAndConnect(port, initSSHTimeout(cfg.Accel))
			if err != nil {
				return err
			}
			sshDur := time.Since(t1)
			sshSum += sshDur

			cmdStart := time.Now()
			_, err = lssh.Exec(client, "uname -a")
			if err != nil {
				_ = client.Close()
				return err
			}
			rttDur := time.Since(cmdStart)
			rttSum += rttDur
			_ = client.Close()

			_ = vm.EnsureBootSnapshot(dir)
			_ = fs.Stop(dir)
			_ = vm.Stop(dir)
			ui.Info(fmt.Sprintf("run %d complete: boot=%s ssh=%s rtt=%s", i+1, bootDur, sshDur, rttDur))
		}

		avgBoot := bootSum / time.Duration(benchmarkRuns)
		avgSSH := sshSum / time.Duration(benchmarkRuns)
		avgRTT := rttSum / time.Duration(benchmarkRuns)
		fmt.Printf("Benchmark (%d runs)\n", benchmarkRuns)
		fmt.Printf("avg boot time: %s\n", avgBoot)
		fmt.Printf("avg SSH connect time: %s\n", avgSSH)
		fmt.Printf("avg command RTT: %s\n", avgRTT)
		fmt.Printf("summary: %s\n", strings.Join([]string{avgBoot.String(), avgSSH.String(), avgRTT.String()}, " | "))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(benchmarkCmd)
	benchmarkCmd.Flags().IntVar(&benchmarkRuns, "runs", 3, "number of benchmark runs")
}
