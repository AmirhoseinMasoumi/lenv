package config

import (
	"fmt"
	"regexp"
	"strings"
)

var memRe = regexp.MustCompile(`^[1-9][0-9]*[MG]$`)

func Validate(lt *LenvToml) error {
	d := strings.TrimSpace(lt.Env.Distro)
	if d != "" && d != "alpine" && d != "ubuntu" && d != "debian" && d != "arch" {
		return fmt.Errorf("env.distro must be one of alpine|ubuntu|debian|arch")
	}
	if lt.Env.CPUs < 0 {
		return fmt.Errorf("env.cpus must be greater than 0")
	}
	if lt.Env.Memory != "" && !memRe.MatchString(strings.ToUpper(lt.Env.Memory)) {
		return fmt.Errorf("env.memory must look like 512M, 2G, or 4G")
	}
	if lt.Mount.Workspace != "" && !strings.HasPrefix(lt.Mount.Workspace, "/") {
		return fmt.Errorf("mount.workspace must be an absolute Linux path, e.g. /workspace")
	}
	for i, pkg := range lt.Packages.Install {
		if strings.TrimSpace(pkg) == "" {
			return fmt.Errorf("packages.install[%d] cannot be empty", i)
		}
	}
	return nil
}
