package vm

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/AmirhoseinMasoumi/Lenv/config"
	"github.com/BurntSushi/toml"
)

type instanceRecord struct {
	Instance   string `toml:"instance"`
	ProjectDir string `toml:"project_dir"`
	PID        int    `toml:"pid"`
	SSHPort    int    `toml:"ssh_port"`
	Accel      string `toml:"accel"`
	Distro     string `toml:"distro"`
}

func instanceStoreDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".lenv", "instances"), nil
}

func writeInstanceRecord(projectDir string, cfg *config.Config, sshPort, pid int) error {
	dir, err := instanceStoreDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create instance registry dir: %w", err)
	}
	rec := instanceRecord{
		Instance:   InstanceName(projectDir),
		ProjectDir: projectDir,
		PID:        pid,
		SSHPort:    sshPort,
		Accel:      cfg.Accel,
		Distro:     cfg.Distro,
	}
	f, err := os.Create(filepath.Join(dir, rec.Instance+".toml"))
	if err != nil {
		return fmt.Errorf("create instance record: %w", err)
	}
	defer f.Close()
	if err := toml.NewEncoder(f).Encode(rec); err != nil {
		return fmt.Errorf("encode instance record: %w", err)
	}
	return nil
}

func removeInstanceRecord(projectDir string) error {
	dir, err := instanceStoreDir()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, InstanceName(projectDir)+".toml")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove instance record: %w", err)
	}
	return nil
}

func RemoveInstance(projectDir string) error {
	return removeInstanceRecord(projectDir)
}

func ListRunningStatuses() ([]Status, error) {
	dir, err := instanceStoreDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("list instances: %w", err)
	}
	statuses := make([]Status, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".toml" {
			continue
		}
		path := filepath.Join(dir, e.Name())
		var rec instanceRecord
		if _, err := toml.DecodeFile(path, &rec); err != nil {
			continue
		}
		running := processRunning(rec.PID)
		if !running {
			_ = os.Remove(path)
			continue
		}
		statuses = append(statuses, Status{
			Instance:   rec.Instance,
			ProjectDir: rec.ProjectDir,
			Accel:      rec.Accel,
			Distro:     rec.Distro,
			PID:        rec.PID,
			SSHPort:    rec.SSHPort,
			Running:    true,
		})
	}
	sort.Slice(statuses, func(i, j int) bool { return statuses[i].Instance < statuses[j].Instance })
	return statuses, nil
}

