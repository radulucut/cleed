package cleed

import (
	"bytes"
	"os"
	"testing"

	"github.com/radulucut/cleed/internal"
	"github.com/radulucut/cleed/internal/storage"
	"github.com/radulucut/cleed/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_Version(t *testing.T) {
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

	os.Args = []string{"cleed", "version"}

	err = root.Cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "cleed version 0.1.0\n", out.String())
}
