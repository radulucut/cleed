package storage

import (
	"encoding/json"
	"os"
)

const (
	configFile = "config.json"
)

type Config struct {
	Version string `json:"version"`
}

func (s *LocalStorage) LoadConfig() (*Config, error) {
	if s.config != nil {
		return s.config, nil
	}
	configPath, err := s.joinConfigDir(configFile)
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	s.config = &Config{}
	err = json.Unmarshal(b, s.config)
	if err != nil {
		return nil, err
	}
	return s.config, nil
}

func (s *LocalStorage) SaveConfig() error {
	if s.config == nil {
		return nil
	}
	configPath, err := s.joinConfigDir(configFile)
	if err != nil {
		return err
	}
	b, err := json.Marshal(s.config)
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, b, 0600)
}
