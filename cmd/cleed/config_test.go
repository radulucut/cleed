package cleed

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/radulucut/cleed/internal"
	_storage "github.com/radulucut/cleed/internal/storage"
	"github.com/radulucut/cleed/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_Config(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	out := new(bytes.Buffer)
	printer := internal.NewPrinter(nil, out, out)
	storage := _storage.NewLocalStorage("cleed_test", timeMock)
	defer localStorageCleanup(t, storage)

	feed := internal.NewTerminalFeed(timeMock, printer, storage)
	feed.SetAgent("cleed/test")

	root, err := NewRoot("0.1.0", timeMock, printer, storage, feed)
	assert.NoError(t, err)

	os.Args = []string{"cleed", "config"}

	err = root.Cmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, `Styling: enabled
Color map:
Summary: disabled
`, out.String())

	config, err := storage.LoadConfig()
	assert.NoError(t, err)
	expectedConfig := &_storage.Config{
		Version:  "0.1.0",
		LastRun:  time.Time{},
		Styling:  0,
		ColorMap: make(map[uint8]uint8),
	}
	assert.Equal(t, expectedConfig, config)
}

func Test_Config_Styling(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	out := new(bytes.Buffer)
	printer := internal.NewPrinter(nil, out, out)
	storage := _storage.NewLocalStorage("cleed_test", timeMock)
	defer localStorageCleanup(t, storage)

	feed := internal.NewTerminalFeed(timeMock, printer, storage)
	feed.SetAgent("cleed/test")

	root, err := NewRoot("0.1.0", timeMock, printer, storage, feed)
	assert.NoError(t, err)

	os.Args = []string{"cleed", "config", "--styling", "2"}

	err = root.Cmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, "styling was updated\n", out.String())

	config, err := storage.LoadConfig()
	assert.NoError(t, err)
	expectedConfig := &_storage.Config{
		Version:  "0.1.0",
		LastRun:  time.Time{},
		Styling:  2,
		ColorMap: make(map[uint8]uint8),
	}
	assert.Equal(t, expectedConfig, config)
}

func Test_Config_Summary(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	out := new(bytes.Buffer)
	printer := internal.NewPrinter(nil, out, out)
	storage := _storage.NewLocalStorage("cleed_test", timeMock)
	defer localStorageCleanup(t, storage)

	feed := internal.NewTerminalFeed(timeMock, printer, storage)
	feed.SetAgent("cleed/test")

	root, err := NewRoot("0.1.0", timeMock, printer, storage, feed)
	assert.NoError(t, err)

	os.Args = []string{"cleed", "config", "--summary", "1"}

	err = root.Cmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, "summary was updated\n", out.String())

	config, err := storage.LoadConfig()
	assert.NoError(t, err)
	expectedConfig := &_storage.Config{
		Version:  "0.1.0",
		LastRun:  time.Time{},
		Styling:  0,
		Summary:  1,
		ColorMap: make(map[uint8]uint8),
	}
	assert.Equal(t, expectedConfig, config)
}

func Test_Config_MapColors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	out := new(bytes.Buffer)
	printer := internal.NewPrinter(nil, out, out)
	storage := _storage.NewLocalStorage("cleed_test", timeMock)
	defer localStorageCleanup(t, storage)

	feed := internal.NewTerminalFeed(timeMock, printer, storage)
	feed.SetAgent("cleed/test")

	root, err := NewRoot("0.1.0", timeMock, printer, storage, feed)
	assert.NoError(t, err)

	os.Args = []string{"cleed", "config", "--map-colors", "1:2,3:4"}

	err = root.Cmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, "color map updated\n", out.String())

	config, err := storage.LoadConfig()
	assert.NoError(t, err)
	expectedConfig := &_storage.Config{
		Version:  "0.1.0",
		LastRun:  time.Time{},
		Styling:  0,
		ColorMap: map[uint8]uint8{1: 2, 3: 4},
	}
	assert.Equal(t, expectedConfig, config)
}

func Test_Config_MapColors_RemoveColorMapping(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	out := new(bytes.Buffer)
	printer := internal.NewPrinter(nil, out, out)
	storage := _storage.NewLocalStorage("cleed_test", timeMock)
	defer localStorageCleanup(t, storage)
	storage.Init("0.1.0")

	config, err := storage.LoadConfig()
	assert.NoError(t, err)
	config.ColorMap = map[uint8]uint8{1: 2, 3: 4}
	err = storage.SaveConfig()
	assert.NoError(t, err)

	feed := internal.NewTerminalFeed(timeMock, printer, storage)
	feed.SetAgent("cleed/test")

	root, err := NewRoot("0.1.0", timeMock, printer, storage, feed)
	assert.NoError(t, err)

	os.Args = []string{"cleed", "config", "--map-colors", "1:"}

	err = root.Cmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, "color map updated\n", out.String())

	config, err = storage.LoadConfig()
	assert.NoError(t, err)
	expectedConfig := &_storage.Config{
		Version:  "0.1.0",
		LastRun:  time.Time{},
		Styling:  0,
		ColorMap: map[uint8]uint8{3: 4},
	}
	assert.Equal(t, expectedConfig, config)
}

func Test_Config_MapColors_ClearColorMapping(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	out := new(bytes.Buffer)
	printer := internal.NewPrinter(nil, out, out)
	storage := _storage.NewLocalStorage("cleed_test", timeMock)
	defer localStorageCleanup(t, storage)
	storage.Init("0.1.0")

	config, err := storage.LoadConfig()
	assert.NoError(t, err)
	config.ColorMap = map[uint8]uint8{1: 2, 3: 4}
	err = storage.SaveConfig()
	assert.NoError(t, err)

	feed := internal.NewTerminalFeed(timeMock, printer, storage)
	feed.SetAgent("cleed/test")

	root, err := NewRoot("0.1.0", timeMock, printer, storage, feed)
	assert.NoError(t, err)

	os.Args = []string{"cleed", "config", "--map-colors="}

	err = root.Cmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, "color map updated\n", out.String())

	config, err = storage.LoadConfig()
	assert.NoError(t, err)
	expectedConfig := &_storage.Config{
		Version:  "0.1.0",
		LastRun:  time.Time{},
		Styling:  0,
		ColorMap: map[uint8]uint8{},
	}
	assert.Equal(t, expectedConfig, config)
}

func Test_Config_ColorRange(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	out := new(bytes.Buffer)
	printer := internal.NewPrinter(nil, out, out)
	storage := _storage.NewLocalStorage("cleed_test", timeMock)
	defer localStorageCleanup(t, storage)
	storage.Init("0.1.0")

	config, err := storage.LoadConfig()
	assert.NoError(t, err)
	config.ColorMap = map[uint8]uint8{1: 2, 3: 4}
	err = storage.SaveConfig()
	assert.NoError(t, err)

	feed := internal.NewTerminalFeed(timeMock, printer, storage)
	feed.SetAgent("cleed/test")

	root, err := NewRoot("0.1.0", timeMock, printer, storage, feed)
	assert.NoError(t, err)

	os.Args = []string{"cleed", "config", "--color-range"}

	err = root.Cmd.Execute()
	assert.NoError(t, err)

	expectedOutput := ""
	for i := 0; i < 256; i++ {
		expectedOutput += fmt.Sprintf("\033[38;5;%dm%d \033[0m", i, i)
	}
	expectedOutput += "\n"
	assert.Equal(t, expectedOutput, out.String())
}
