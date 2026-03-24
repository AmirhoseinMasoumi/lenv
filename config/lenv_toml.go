package config

import (
"errors"
"fmt"
"os"
"path/filepath"

"github.com/BurntSushi/toml"
)

func Load(projectDir string) (*LenvToml, error) {
path := filepath.Join(projectDir, "lenv.toml")
var lt LenvToml
_, err := os.Stat(path)
if errors.Is(err, os.ErrNotExist) {
return &lt, nil
}
if err != nil {
return nil, fmt.Errorf("stat lenv.toml: %w", err)
}
if _, err := toml.DecodeFile(path, &lt); err != nil {
return nil, fmt.Errorf("parse lenv.toml: %w", err)
}
return &lt, nil
}

func WriteResolved(path string, cfg *Config) error {
f, err := os.Create(path)
if err != nil {
return fmt.Errorf("create resolved config: %w", err)
}
defer f.Close()
enc := toml.NewEncoder(f)
if err := enc.Encode(cfg); err != nil {
return fmt.Errorf("encode resolved config: %w", err)
}
return nil
}

