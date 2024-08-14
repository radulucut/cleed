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

func Test_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	out := new(bytes.Buffer)
	printer := internal.NewPrinter(nil, out, out)
	storage := storage.NewLocalStorage("cleed_test", timeMock)
	defer localStorageCleanup(t, storage)

	configDir, err := os.UserConfigDir()
	if err != nil {
		t.Fatal(err)
	}

	listsDir := path.Join(configDir, "cleed_test", "lists")
	err = os.MkdirAll(listsDir, 0700)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(path.Join(listsDir, "default"), []byte{}, 0600)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(path.Join(listsDir, "mylist"), []byte{}, 0600)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(path.Join(listsDir, "abc"), []byte{}, 0600)
	if err != nil {
		t.Fatal(err)
	}

	feed := internal.NewTerminalFeed(timeMock, printer, storage)
	feed.SetAgent("cleed/test")

	root, err := NewRoot("0.1.0", timeMock, printer, storage, feed)
	assert.NoError(t, err)

	os.Args = []string{"cleed", "list"}

	err = root.Cmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, "abc\ndefault\nmylist\n", out.String())
}

func Test_List_No_List(t *testing.T) {
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

	os.Args = []string{"cleed", "list"}

	err = root.Cmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, "default\n", out.String())
}

func Test_List_Feeds(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	out := new(bytes.Buffer)
	printer := internal.NewPrinter(nil, out, out)
	storage := storage.NewLocalStorage("cleed_test", timeMock)
	defer localStorageCleanup(t, storage)

	configDir, err := os.UserConfigDir()
	if err != nil {
		t.Fatal(err)
	}

	listsDir := path.Join(configDir, "cleed_test", "lists")
	err = os.MkdirAll(listsDir, 0700)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(path.Join(listsDir, "test"),
		[]byte(fmt.Sprintf("%d %s\n%d %s\n",
			defaultCurrentTime.Unix(), "https://example.com",
			defaultCurrentTime.Unix(), "https://test.com",
		),
		), 0600)
	if err != nil {
		t.Fatal(err)
	}

	feed := internal.NewTerminalFeed(timeMock, printer, storage)
	feed.SetAgent("cleed/test")

	root, err := NewRoot("0.1.0", timeMock, printer, storage, feed)
	assert.NoError(t, err)

	os.Args = []string{"cleed", "list", "test"}

	err = root.Cmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%s  %s\n%s  %s\n%s\n",
		defaultCurrentTime.Local().Format("2006-01-02 15:04:05"),
		"https://example.com",
		defaultCurrentTime.Local().Format("2006-01-02 15:04:05"),
		"https://test.com",
		"Total: 2 feeds",
	), out.String())
}
