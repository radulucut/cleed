package storage

import (
	"os"
	"path"

	"github.com/radulucut/cleed/internal/utils"
)

const (
	listsDir = "lists"
)

type LocalStorage struct {
	name string
	time utils.Time

	config *Config
}

func NewLocalStorage(
	name string,
	time utils.Time,
) *LocalStorage {
	return &LocalStorage{
		name: name,
		time: time,
	}
}

func (s *LocalStorage) Init(version string) error {
	configDir, err := s.JoinConfigDir("")
	if err != nil {
		return err
	}
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll(path.Join(configDir, listsDir), 0755)
	if err != nil {
		return err
	}
	cacheDir, err := s.JoinCacheDir("")
	if err != nil {
		return err
	}
	err = os.MkdirAll(cacheDir, 0755)
	if err != nil {
		return err
	}
	_, err = s.LoadConfig()
	if err != nil {
		if os.IsNotExist(err) {
			s.config = &Config{
				Version:  version,
				ColorMap: make(map[uint8]uint8),
			}
			err = s.SaveConfig()
		}
		return err
	}
	return s.Migrate()
}

func (s *LocalStorage) Migrate() error {
	// handle migration here
	return nil
}

func (s *LocalStorage) ClearAll() error {
	configDir, err := s.JoinConfigDir("")
	if err != nil {
		return err
	}
	err = os.RemoveAll(configDir)
	if err != nil {
		return err
	}
	cacheDir, err := s.JoinCacheDir("")
	if err != nil {
		return err
	}
	err = os.RemoveAll(cacheDir)
	if err != nil {
		return err
	}
	return nil
}

func (s *LocalStorage) JoinCacheDir(file string) (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return path.Join(base, s.name, file), nil
}

func (s *LocalStorage) JoinConfigDir(file string) (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return path.Join(base, s.name, file), nil
}

func (s *LocalStorage) joinListsDir(file string) (string, error) {
	base, err := s.JoinConfigDir(listsDir)
	if err != nil {
		return "", err
	}
	return path.Join(base, file), nil
}
