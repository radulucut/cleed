package cleed

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path"
	"testing"
	"time"

	"github.com/radulucut/cleed/internal"
	_storage "github.com/radulucut/cleed/internal/storage"
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
	storage := _storage.NewLocalStorage("cleed_test", timeMock)
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
	storage := _storage.NewLocalStorage("cleed_test", timeMock)
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
	storage := _storage.NewLocalStorage("cleed_test", timeMock)
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

func Test_List_Rename(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	out := new(bytes.Buffer)
	printer := internal.NewPrinter(nil, out, out)
	storage := _storage.NewLocalStorage("cleed_test", timeMock)
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
			defaultCurrentTime.Unix()+300, "https://test.com",
		),
		), 0600)
	if err != nil {
		t.Fatal(err)
	}

	feed := internal.NewTerminalFeed(timeMock, printer, storage)
	feed.SetAgent("cleed/test")

	root, err := NewRoot("0.1.0", timeMock, printer, storage, feed)
	assert.NoError(t, err)

	os.Args = []string{"cleed", "list", "test", "--rename", "newlist"}

	err = root.Cmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, "list test was renamed to newlist\n", out.String())

	lists, err := storage.LoadLists()
	assert.NoError(t, err)
	assert.Equal(t, []string{"newlist"}, lists)

	items, err := storage.GetFeedsFromList("newlist")
	assert.NoError(t, err)
	assert.Equal(t, []*_storage.ListItem{
		{AddedAt: time.Unix(defaultCurrentTime.Unix(), 0), Address: "https://example.com"},
		{AddedAt: time.Unix(defaultCurrentTime.Unix()+300, 0), Address: "https://test.com"},
	}, items)
}

func Test_List_Merge(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	out := new(bytes.Buffer)
	printer := internal.NewPrinter(nil, out, out)
	storage := _storage.NewLocalStorage("cleed_test", timeMock)
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
			defaultCurrentTime.Unix()+300, "https://test.com",
		),
		), 0600)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(path.Join(listsDir, "test2"),
		[]byte(fmt.Sprintf("%d %s\n%d %s\n%d %s\n",
			defaultCurrentTime.Unix()-300, "https://example0.com",
			defaultCurrentTime.Unix()+400, "https://test.com",
			defaultCurrentTime.Unix()+500, "https://example2.com",
		),
		), 0600)
	if err != nil {
		t.Fatal(err)
	}

	feed := internal.NewTerminalFeed(timeMock, printer, storage)
	feed.SetAgent("cleed/test")

	root, err := NewRoot("0.1.0", timeMock, printer, storage, feed)
	assert.NoError(t, err)

	os.Args = []string{"cleed", "list", "test", "--merge", "test2"}

	err = root.Cmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, "list test was merged with test2. test2 was removed\n", out.String())

	lists, err := storage.LoadLists()
	assert.NoError(t, err)
	assert.Equal(t, []string{"test"}, lists)

	items, err := storage.GetFeedsFromList("test")
	assert.NoError(t, err)
	assert.Equal(t, []*_storage.ListItem{
		{AddedAt: time.Unix(defaultCurrentTime.Unix()-300, 0), Address: "https://example0.com"},
		{AddedAt: time.Unix(defaultCurrentTime.Unix(), 0), Address: "https://example.com"},
		{AddedAt: time.Unix(defaultCurrentTime.Unix()+300, 0), Address: "https://test.com"},
		{AddedAt: time.Unix(defaultCurrentTime.Unix()+500, 0), Address: "https://example2.com"},
	}, items)
}

func Test_List_Remove(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	out := new(bytes.Buffer)
	printer := internal.NewPrinter(nil, out, out)
	storage := _storage.NewLocalStorage("cleed_test", timeMock)
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

	err = os.WriteFile(path.Join(listsDir, "default"),
		[]byte(fmt.Sprintf("%d %s\n%d %s\n",
			defaultCurrentTime.Unix(), "https://example.com",
			defaultCurrentTime.Unix(), "https://test.com",
		),
		), 0600)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(path.Join(listsDir, "test"),
		[]byte(fmt.Sprintf("%d %s\n%d %s\n",
			defaultCurrentTime.Unix(), "https://test.com",
			defaultCurrentTime.Unix(), "https://example2.com",
		),
		), 0600)
	if err != nil {
		t.Fatal(err)
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		t.Fatal(err)
	}
	cacheDir = path.Join(cacheDir, "cleed_test")
	err = os.MkdirAll(cacheDir, 0700)
	if err != nil {
		t.Fatal(err)
	}

	err = storage.SaveCacheInfo(map[string]*_storage.CacheInfoItem{
		"https://example.com": {
			URL:        "https://example.com",
			LastFetch:  defaultCurrentTime,
			ETag:       "etag",
			FetchAfter: time.Unix(0, 0),
		},
		"https://test.com": {
			URL:        "https://test.com",
			LastFetch:  defaultCurrentTime,
			ETag:       "etag",
			FetchAfter: time.Unix(0, 0),
		},
		"https://example2.com": {
			URL:        "https://example2.com",
			LastFetch:  defaultCurrentTime,
			ETag:       "etag",
			FetchAfter: time.Unix(0, 0),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	err = storage.SaveFeedCache(bytes.NewBuffer([]byte("example")), "https://example.com")
	if err != nil {
		t.Fatal(err)
	}

	err = storage.SaveFeedCache(bytes.NewBuffer([]byte("test")), "https://test.com")
	if err != nil {
		t.Fatal(err)
	}

	err = storage.SaveFeedCache(bytes.NewBuffer([]byte("example2")), "https://example2.com")
	if err != nil {
		t.Fatal(err)
	}

	feed := internal.NewTerminalFeed(timeMock, printer, storage)
	feed.SetAgent("cleed/test")

	root, err := NewRoot("0.1.0", timeMock, printer, storage, feed)
	assert.NoError(t, err)

	os.Args = []string{"cleed", "list", "test", "--remove"}

	err = root.Cmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, "list test was removed\n", out.String())

	files, err := os.ReadDir(cacheDir)
	if err != nil {
		t.Fatal(err)
	}
	assert.Len(t, files, 3)

	cacheInfo, err := storage.LoadCacheInfo()
	if err != nil {
		t.Fatal(err)
	}
	assert.Len(t, cacheInfo, 2)
	assert.Equal(t, &_storage.CacheInfoItem{
		URL:        "https://example.com",
		LastFetch:  time.Unix(defaultCurrentTime.Unix(), 0),
		ETag:       "etag",
		FetchAfter: time.Unix(0, 0),
	}, cacheInfo["https://example.com"])
	assert.Equal(t, &_storage.CacheInfoItem{
		URL:        "https://test.com",
		LastFetch:  time.Unix(defaultCurrentTime.Unix(), 0),
		ETag:       "etag",
		FetchAfter: time.Unix(0, 0),
	}, cacheInfo["https://test.com"])

	assert.NoFileExists(t, path.Join(cacheDir, "feed_"+url.QueryEscape("https://example2.com")))

	b, err := os.ReadFile(path.Join(cacheDir, "feed_"+url.QueryEscape("https://example.com")))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "example", string(b))

	b, err = os.ReadFile(path.Join(cacheDir, "feed_"+url.QueryEscape("https://test.com")))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "test", string(b))
}

func Test_List_ImportFromFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	out := new(bytes.Buffer)
	printer := internal.NewPrinter(nil, out, out)
	storage := _storage.NewLocalStorage("cleed_test", timeMock)
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
			defaultCurrentTime.Unix()+300, "https://test.com",
		),
		), 0600)
	if err != nil {
		t.Fatal(err)
	}

	importFilePath := path.Join(listsDir, "import")
	err = os.WriteFile(importFilePath,
		[]byte(`https://example0.com
 https://test.com
# comment
https://example2.com`), 0600)
	if err != nil {
		t.Fatal(err)
	}

	feed := internal.NewTerminalFeed(timeMock, printer, storage)
	feed.SetAgent("cleed/test")

	root, err := NewRoot("0.1.0", timeMock, printer, storage, feed)
	assert.NoError(t, err)

	os.Args = []string{"cleed", "list", "test", "--import-from-file", importFilePath}

	err = root.Cmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, "added 3 feeds to list: test\n", out.String())

	items, err := storage.GetFeedsFromList("test")
	assert.NoError(t, err)
	assert.Equal(t, []*_storage.ListItem{
		{AddedAt: time.Unix(defaultCurrentTime.Unix(), 0), Address: "https://example.com"},
		{AddedAt: time.Unix(defaultCurrentTime.Unix()+300, 0), Address: "https://test.com"},
		{AddedAt: time.Unix(defaultCurrentTime.Unix(), 0), Address: "https://example0.com"},
		{AddedAt: time.Unix(defaultCurrentTime.Unix(), 0), Address: "https://example2.com"},
	}, items)
}

func Test_List_ExportToFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	out := new(bytes.Buffer)
	printer := internal.NewPrinter(nil, out, out)
	storage := _storage.NewLocalStorage("cleed_test", timeMock)
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
			defaultCurrentTime.Unix()+300, "https://test.com",
		),
		), 0600)
	if err != nil {
		t.Fatal(err)
	}

	feed := internal.NewTerminalFeed(timeMock, printer, storage)
	feed.SetAgent("cleed/test")

	root, err := NewRoot("0.1.0", timeMock, printer, storage, feed)
	assert.NoError(t, err)

	exportPath := path.Join(listsDir, "export")
	os.Args = []string{"cleed", "list", "test", "--export-to-file", exportPath}

	err = root.Cmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("exported 2 feeds to %s\n", exportPath), out.String())

	b, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, `https://example.com
https://test.com
`, string(b))
}

func Test_List_ImportFromOPML(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	out := new(bytes.Buffer)
	printer := internal.NewPrinter(nil, out, out)
	storage := _storage.NewLocalStorage("cleed_test", timeMock)
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
			defaultCurrentTime.Unix()+300, "https://test.com",
		),
		), 0600)
	if err != nil {
		t.Fatal(err)
	}

	importFilePath := path.Join(listsDir, "import.opml")
	err = os.WriteFile(importFilePath,
		[]byte(`<?xml version="1.0" encoding="UTF-8"?>
<opml version="1.0">
  <head>
    <title>test</title>
  </head>
  <body>
	<outline text="test">
      <outline xmlUrl="https://example1.com"/>
      <outline xmlUrl="https://example2.com"/>
      <outline xmlUrl="https://example3.com"/>
	</outline>
  </body>
</opml>`), 0600)
	if err != nil {
		t.Fatal(err)
	}

	feed := internal.NewTerminalFeed(timeMock, printer, storage)
	feed.SetAgent("cleed/test")

	root, err := NewRoot("0.1.0", timeMock, printer, storage, feed)
	assert.NoError(t, err)

	os.Args = []string{"cleed", "list", "test", "--import-from-opml", importFilePath}

	err = root.Cmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, "added 3 feeds to list: test\n", out.String())

	items, err := storage.GetFeedsFromList("test")
	assert.NoError(t, err)
	assert.Equal(t, []*_storage.ListItem{
		{AddedAt: time.Unix(defaultCurrentTime.Unix(), 0), Address: "https://example.com"},
		{AddedAt: time.Unix(defaultCurrentTime.Unix()+300, 0), Address: "https://test.com"},
		{AddedAt: time.Unix(defaultCurrentTime.Unix(), 0), Address: "https://example1.com"},
		{AddedAt: time.Unix(defaultCurrentTime.Unix(), 0), Address: "https://example2.com"},
		{AddedAt: time.Unix(defaultCurrentTime.Unix(), 0), Address: "https://example3.com"},
	}, items)
}

func Test_List_ExportToOPML(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	out := new(bytes.Buffer)
	printer := internal.NewPrinter(nil, out, out)
	storage := _storage.NewLocalStorage("cleed_test", timeMock)
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
			defaultCurrentTime.Unix()+300, "https://test.com",
		),
		), 0600)
	if err != nil {
		t.Fatal(err)
	}

	feed := internal.NewTerminalFeed(timeMock, printer, storage)
	feed.SetAgent("cleed/test")

	root, err := NewRoot("0.1.0", timeMock, printer, storage, feed)
	assert.NoError(t, err)

	exportPath := path.Join(listsDir, "export.opml")
	os.Args = []string{"cleed", "list", "test", "--export-to-opml", exportPath}

	err = root.Cmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("exported 2 feeds to %s\n", exportPath), out.String())

	b, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, `<?xml version="1.0" encoding="UTF-8"?>
<opml version="1.0">
  <head>
    <title>test</title>
  </head>
  <body>
	<outline text="test">
      <outline xmlUrl="https://example.com"/>
      <outline xmlUrl="https://test.com"/>
	</outline>
  </body>
</opml>`, string(b))
}
