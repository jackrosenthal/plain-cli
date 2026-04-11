package config

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
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
		cfg.ClientID, err = randomUUID()
		if err != nil {
			return Config{}, err
		}
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

func randomUUID() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}

	raw[6] = (raw[6] & 0x0f) | 0x40
	raw[8] = (raw[8] & 0x3f) | 0x80

	var dst [36]byte
	hex.Encode(dst[0:8], raw[0:4])
	dst[8] = '-'
	hex.Encode(dst[9:13], raw[4:6])
	dst[13] = '-'
	hex.Encode(dst[14:18], raw[6:8])
	dst[18] = '-'
	hex.Encode(dst[19:23], raw[8:10])
	dst[23] = '-'
	hex.Encode(dst[24:36], raw[10:16])

	return string(dst[:]), nil
}
