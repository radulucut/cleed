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
	configDir, err := s.joinConfigDir("")
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
	cacheDir, err := s.joinCacheDir("")
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
				Version: version,
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
	configDir, err := s.joinConfigDir("")
	if err != nil {
		return err
	}
	err = os.RemoveAll(configDir)
	if err != nil {
		return err
	}
	cacheDir, err := s.joinCacheDir("")
	if err != nil {
		return err
	}
	err = os.RemoveAll(cacheDir)
	if err != nil {
		return err
	}
	return nil
}

func (s *LocalStorage) joinCacheDir(file string) (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return path.Join(base, s.name, file), nil
}

func (s *LocalStorage) joinConfigDir(file string) (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return path.Join(base, s.name, file), nil
}

func (s *LocalStorage) joinListsDir(file string) (string, error) {
	base, err := s.joinConfigDir(listsDir)
	if err != nil {
		return "", err
	}
	return path.Join(base, file), nil
}
