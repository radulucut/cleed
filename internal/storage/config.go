package storage

import (
	"encoding/json"
	"os"
	"time"
)

const (
	configFile = "config.json"
)

type Config struct {
	Version  string          `json:"version"`
	LastRun  time.Time       `json:"lastRun"`
	Styling  uint8           `json:"styling"` // 0: default, 1: enabled, 2: disabled
	Summary  uint8           `json:"summary"` // 0: disabled, 1: enabled
	ColorMap map[uint8]uint8 `json:"colorMap"`
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
	if s.config.ColorMap == nil {
		s.config.ColorMap = make(map[uint8]uint8)
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
