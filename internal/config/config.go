package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/google/uuid"
)

type Config struct {
	Host     string `toml:"host"`
	Token    string `toml:"token"`
	ClientID string `toml:"client_id"`
}

func Load() (Config, error) {
	path, err := configPath()
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return Config{}, err
		}
	}

	if cfg.ClientID == "" {
		u, err := uuid.NewRandom()
		if err != nil {
			return Config{}, err
		}
		cfg.ClientID = u.String()
		if err := cfg.Save(); err != nil {
			return Config{}, err
		}
	}

	return cfg, nil
}

func (c Config) Save() error {
	path, err := configPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		closeErr := file.Close()
		if err == nil {
			err = closeErr
		}
	}()

	if err := toml.NewEncoder(file).Encode(c); err != nil {
		return err
	}

	return err
}

func configPath() (string, error) {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "plain", "config.toml"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".config", "plain", "config.toml"), nil
}
