package ssh

import (
	"fmt"
	"os"
	"strconv"
	"time"

gssh "golang.org/x/crypto/ssh"
)

func WaitAndConnect(port int, timeout time.Duration) (*gssh.Client, error) {
deadline := time.Now().Add(timeout)
for time.Now().Before(deadline) {
c, err := tryConnect(port)
if err == nil {
return c, nil
}
time.Sleep(200 * time.Millisecond)
}
return nil, fmt.Errorf("VM did not become ready within %s", timeout)
}

func tryConnect(port int) (*gssh.Client, error) {
	user := getenvDefault("LENV_SSH_USER", "root")
	password := getenvDefault("LENV_SSH_PASSWORD", "lenv")
	timeoutMS := getenvIntDefault("LENV_SSH_DIAL_TIMEOUT_MS", 500)
	return gssh.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port), &gssh.ClientConfig{
		User:            user,
		Auth:            []gssh.AuthMethod{gssh.Password(password)},
		HostKeyCallback: gssh.InsecureIgnoreHostKey(),
		Timeout:         time.Duration(timeoutMS) * time.Millisecond,
	})
}

func getenvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getenvIntDefault(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

