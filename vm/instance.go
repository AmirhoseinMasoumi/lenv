package vm

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"
)

func InstanceName(projectDir string) string {
	hash := sha256.Sum256([]byte(projectDir))
	base := filepath.Base(projectDir)
	return fmt.Sprintf("lenv-%s-%x", base, hash[:4])
}
