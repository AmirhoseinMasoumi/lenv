package integration

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitRunDestroyAlpine(t *testing.T) {
	if os.Getenv("LENV_INTEGRATION") != "1" {
		t.Skip("set LENV_INTEGRATION=1 to run integration tests")
	}

	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	projectDir := t.TempDir()

	env := append(os.Environ(),
		"LENV_ACCEL=tcg",
		"LENV_SSH_WAIT_TIMEOUT_SECONDS=120",
	)

	defer runGoCommand(t, repoRoot, env, "--project-dir", projectDir, "destroy")

	initOut := runGoCommand(t, repoRoot, env, "--project-dir", projectDir, "init", "--distro", "alpine")
	if !strings.Contains(initOut, "Ready") && !strings.Contains(initOut, "Already running") {
		t.Fatalf("init output missing readiness marker:\n%s", initOut)
	}

	unameOut := runGoCommand(t, repoRoot, env, "--project-dir", projectDir, "run", "uname -a")
	if !strings.Contains(unameOut, "Linux") {
		t.Fatalf("expected uname output to contain Linux, got:\n%s", unameOut)
	}

	_ = runGoCommand(t, repoRoot, env, "--project-dir", projectDir, "destroy")
}

func runGoCommand(t *testing.T, repoRoot string, env []string, args ...string) string {
	t.Helper()
	cmdArgs := append([]string{"run", "."}, args...)
	cmd := exec.Command("go", cmdArgs...)
	cmd.Dir = repoRoot
	cmd.Env = env
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("command failed (%v): %v\n%s", cmdArgs, err, out.String())
	}
	return out.String()
}
