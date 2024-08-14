package cleed

import (
	"testing"
	"time"

	"github.com/radulucut/cleed/internal/storage"
)

var (
	defaultCurrentTime = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
)

func localStorageCleanup(t *testing.T, storage *storage.LocalStorage) {
	err := storage.ClearAll()
	if err != nil {
		t.Fatal(err)
	}
}
