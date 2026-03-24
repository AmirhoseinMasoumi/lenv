package vm

import (
	"fmt"
	"math/rand"
	"net"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func FindFreePort() (int, error) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	excluded := map[int]bool{}
	if runtime.GOOS == "windows" {
		ports, err := windowsExcludedPorts()
		if err == nil {
			excluded = ports
		}
	}

	for i := 0; i < 100; i++ {
		p := 3000 + rng.Intn(1000)
		if excluded[p] {
			continue
		}
		l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", p))
		if err != nil {
			continue
		}
		_ = l.Close()
		return p, nil
	}
	return 0, fmt.Errorf("failed to allocate free SSH port in 3000-3999")
}

func windowsExcludedPorts() (map[int]bool, error) {
	out, err := exec.Command("netsh", "int", "ipv4", "show", "excludedportrange", "protocol=tcp").CombinedOutput()
	if err != nil {
		return nil, err
	}
	re := regexp.MustCompile(`^\s*(\d+)\s+(\d+)\s*$`)
	excluded := map[int]bool{}
	for _, line := range strings.Split(string(out), "\n") {
		m := re.FindStringSubmatch(strings.TrimSpace(line))
		if len(m) != 3 {
			continue
		}
		start, _ := strconv.Atoi(m[1])
		end, _ := strconv.Atoi(m[2])
		for p := start; p <= end; p++ {
			excluded[p] = true
		}
	}
	return excluded, nil
}

