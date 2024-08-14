package cleed

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/radulucut/cleed/internal"
	"github.com/radulucut/cleed/internal/storage"
	"github.com/radulucut/cleed/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_Follow_Default(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	out := new(bytes.Buffer)
	printer := internal.NewPrinter(nil, out, out)
	storage := storage.NewLocalStorage("cleed_test", timeMock)
	defer localStorageCleanup(t, storage)

	feed := internal.NewTerminalFeed(timeMock, printer, storage)
	feed.SetAgent("cleed/test")

	root, err := NewRoot("0.1.0", timeMock, printer, storage, feed)
	assert.NoError(t, err)

	os.Args = []string{"cleed", "follow", "https://example.com", "https://test.com"}

	err = root.Cmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, "added 2 feeds to list: default\n", out.String())

	configDir, err := os.UserConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	listsDir := path.Join(configDir, "cleed_test", "lists")
	files, err := os.ReadDir(listsDir)
	if err != nil {
		t.Fatal(err)
	}
	assert.Len(t, files, 1)
	file := files[0]
	assert.Equal(t, "default", file.Name())
	b, err := os.ReadFile(path.Join(listsDir, file.Name()))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, fmt.Sprintf("%d %s\n%d %s\n",
		defaultCurrentTime.Unix(), "https://example.com",
		defaultCurrentTime.Unix(), "https://test.com",
	), string(b))
}

func Test_Follow_Custom_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	out := new(bytes.Buffer)
	printer := internal.NewPrinter(nil, out, out)
	storage := storage.NewLocalStorage("cleed_test", timeMock)
	defer localStorageCleanup(t, storage)

	feed := internal.NewTerminalFeed(timeMock, printer, storage)
	feed.SetAgent("cleed/test")

	root, err := NewRoot("0.1.0", timeMock, printer, storage, feed)
	assert.NoError(t, err)

	os.Args = []string{"cleed", "follow", "--list", "test", "https://example.com"}

	err = root.Cmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, "added 1 feed to list: test\n", out.String())

	configDir, err := os.UserConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	listsDir := path.Join(configDir, "cleed_test", "lists")
	files, err := os.ReadDir(listsDir)
	if err != nil {
		t.Fatal(err)
	}
	assert.Len(t, files, 1)
	file := files[0]
	assert.Equal(t, "test", file.Name())
	b, err := os.ReadFile(path.Join(listsDir, file.Name()))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, fmt.Sprintf("%d %s\n", defaultCurrentTime.Unix(), "https://example.com"), string(b))
}

func Test_Follow_Invalid_URL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	out := new(bytes.Buffer)
	printer := internal.NewPrinter(nil, out, out)
	storage := storage.NewLocalStorage("cleed_test", timeMock)
	defer localStorageCleanup(t, storage)

	feed := internal.NewTerminalFeed(timeMock, printer, storage)
	feed.SetAgent("cleed/test")

	root, err := NewRoot("0.1.0", timeMock, printer, storage, feed)
	assert.NoError(t, err)

	os.Args = []string{"cleed", "follow", "invalid", "https://test.com"}

	err = root.Cmd.Execute()
	assert.EqualError(t, err, "failed to parse URL: invalid")

	configDir, err := os.UserConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	listsDir := path.Join(configDir, "cleed_test", "lists")
	files, err := os.ReadDir(listsDir)
	if err != nil {
		t.Fatal(err)
	}
	assert.Len(t, files, 0)
}
