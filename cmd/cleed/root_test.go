package cleed

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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

func Test_Feed(t *testing.T) {
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

	rss := createDefaultRSS()
	atom := createDefaultAtom()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rss" {
			w.Header().Set("ETag", "123")
			w.Write([]byte(rss))
		} else if r.URL.Path == "/atom" {
			w.Write([]byte(atom))
		}
	}))
	defer server.Close()

	err = os.WriteFile(path.Join(listsDir, "default"),
		[]byte(fmt.Sprintf("%d %s\n",
			defaultCurrentTime.Unix(), server.URL+"/rss",
		),
		), 0600)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(path.Join(listsDir, "test"),
		[]byte(fmt.Sprintf("%d %s\n",
			defaultCurrentTime.Unix(), server.URL+"/atom",
		),
		), 0600)
	if err != nil {
		t.Fatal(err)
	}

	feed := internal.NewTerminalFeed(timeMock, printer, storage)
	feed.SetAgent("cleed/test")

	root, err := NewRoot("0.1.0", timeMock, printer, storage, feed)
	assert.NoError(t, err)

	os.Args = []string{"cleed"}

	err = root.Cmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, `RSS Feed        • Item 2
1688 days ago   https://rss-feed.com/item-2/

Atom Feed       • Item 2
1594 days ago   https://atom-feed.com/item-2/

Atom Feed       • Item 1
18 hours ago    https://atom-feed.com/item-1/

RSS Feed        • Item 1
15 minutes ago  https://rss-feed.com/item-1/

`, out.String())

	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		t.Fatal(err)
	}
	cacheDir := path.Join(userCacheDir, "cleed_test")
	files, err := os.ReadDir(cacheDir)
	if err != nil {
		t.Fatal(err)
	}
	assert.Len(t, files, 3)

	cacheInfo, err := storage.LoadCacheInfo()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(cacheInfo))
	assert.Equal(t, &_storage.CacheInfoItem{
		URL:        server.URL + "/rss",
		LastCheck:  time.Unix(defaultCurrentTime.Unix(), 0),
		ETag:       "123",
		FetchAfter: time.Unix(defaultCurrentTime.Unix()+60, 0),
	}, cacheInfo[server.URL+"/rss"])
	assert.Equal(t, &_storage.CacheInfoItem{
		URL:        server.URL + "/atom",
		LastCheck:  time.Unix(defaultCurrentTime.Unix(), 0),
		ETag:       "",
		FetchAfter: time.Unix(defaultCurrentTime.Unix()+60, 0),
	}, cacheInfo[server.URL+"/atom"])

	b, err := os.ReadFile(path.Join(cacheDir, "feed_"+url.QueryEscape(server.URL+"/rss")))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, rss, string(b))

	b, err = os.ReadFile(path.Join(cacheDir, "feed_"+url.QueryEscape(server.URL+"/atom")))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, atom, string(b))
}

func Test_Feed_With_Summary(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timeMock := mocks.NewMockTime(ctrl)
	timeMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	out := new(bytes.Buffer)
	printer := internal.NewPrinter(nil, out, out)
	storage := _storage.NewLocalStorage("cleed_test", timeMock)
	defer localStorageCleanup(t, storage)
	storage.Init("0.1.0")

	configDir, err := os.UserConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	listsDir := path.Join(configDir, "cleed_test", "lists")
	err = os.MkdirAll(listsDir, 0700)
	if err != nil {
		t.Fatal(err)
	}

	config, err := storage.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	config.Summary = 1
	err = storage.SaveConfig()
	if err != nil {
		t.Fatal(err)
	}

	rss := createDefaultRSS()
	atom := createDefaultAtom()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rss" {
			w.Header().Set("ETag", "123")
			w.Write([]byte(rss))
		} else if r.URL.Path == "/atom" {
			w.Write([]byte(atom))
		}
	}))
	defer server.Close()

	err = os.WriteFile(path.Join(listsDir, "default"),
		[]byte(fmt.Sprintf("%d %s\n",
			defaultCurrentTime.Unix(), server.URL+"/rss",
		),
		), 0600)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(path.Join(listsDir, "test"),
		[]byte(fmt.Sprintf("%d %s\n",
			defaultCurrentTime.Unix(), server.URL+"/atom",
		),
		), 0600)
	if err != nil {
		t.Fatal(err)
	}

	feed := internal.NewTerminalFeed(timeMock, printer, storage)
	feed.SetAgent("cleed/test")

	root, err := NewRoot("0.1.0", timeMock, printer, storage, feed)
	assert.NoError(t, err)

	os.Args = []string{"cleed"}

	err = root.Cmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, `RSS Feed        • Item 2
1688 days ago   https://rss-feed.com/item-2/

Atom Feed       • Item 2
1594 days ago   https://atom-feed.com/item-2/

Atom Feed       • Item 1
18 hours ago    https://atom-feed.com/item-1/

RSS Feed        • Item 1
15 minutes ago  https://rss-feed.com/item-1/

Displayed 4 items from 2 feeds (0 cached, 2 fetched) with 4 items in 0.00s
`, out.String())

	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		t.Fatal(err)
	}
	cacheDir := path.Join(userCacheDir, "cleed_test")
	files, err := os.ReadDir(cacheDir)
	if err != nil {
		t.Fatal(err)
	}
	assert.Len(t, files, 3)

	cacheInfo, err := storage.LoadCacheInfo()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(cacheInfo))
	assert.Equal(t, &_storage.CacheInfoItem{
		URL:        server.URL + "/rss",
		LastCheck:  time.Unix(defaultCurrentTime.Unix(), 0),
		ETag:       "123",
		FetchAfter: time.Unix(defaultCurrentTime.Unix()+60, 0),
	}, cacheInfo[server.URL+"/rss"])
	assert.Equal(t, &_storage.CacheInfoItem{
		URL:        server.URL + "/atom",
		LastCheck:  time.Unix(defaultCurrentTime.Unix(), 0),
		ETag:       "",
		FetchAfter: time.Unix(defaultCurrentTime.Unix()+60, 0),
	}, cacheInfo[server.URL+"/atom"])

	b, err := os.ReadFile(path.Join(cacheDir, "feed_"+url.QueryEscape(server.URL+"/rss")))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, rss, string(b))

	b, err = os.ReadFile(path.Join(cacheDir, "feed_"+url.QueryEscape(server.URL+"/atom")))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, atom, string(b))
}

func Test_Feed_Specific_List(t *testing.T) {
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

	atom := createDefaultAtom()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(atom))
	}))
	defer server.Close()

	err = os.WriteFile(path.Join(listsDir, "default"),
		[]byte(fmt.Sprintf("%d %s\n",
			defaultCurrentTime.Unix(), server.URL+"/rss",
		),
		), 0600)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(path.Join(listsDir, "test"),
		[]byte(fmt.Sprintf("%d %s\n",
			defaultCurrentTime.Unix(), server.URL+"/atom",
		),
		), 0600)
	if err != nil {
		t.Fatal(err)
	}

	feed := internal.NewTerminalFeed(timeMock, printer, storage)
	feed.SetAgent("cleed/test")

	root, err := NewRoot("0.1.0", timeMock, printer, storage, feed)
	assert.NoError(t, err)

	os.Args = []string{"cleed", "--list", "test"}

	err = root.Cmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, `Atom Feed      • Item 2
1594 days ago  https://atom-feed.com/item-2/

Atom Feed      • Item 1
18 hours ago   https://atom-feed.com/item-1/

`, out.String())
}

func Test_Feed_NotModified(t *testing.T) {
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

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotModified)
	}))
	defer server.Close()

	err = os.WriteFile(path.Join(listsDir, "default"),
		[]byte(fmt.Sprintf("%d %s\n",
			defaultCurrentTime.Unix(), server.URL,
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
	rss := createDefaultRSS()
	err = storage.SaveFeedCache(bytes.NewBufferString(rss), server.URL)
	if err != nil {
		t.Fatal(err)
	}

	feed := internal.NewTerminalFeed(timeMock, printer, storage)
	feed.SetAgent("cleed/test")

	root, err := NewRoot("0.1.0", timeMock, printer, storage, feed)
	assert.NoError(t, err)

	os.Args = []string{"cleed"}

	err = root.Cmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, `RSS Feed        • Item 2
1688 days ago   https://rss-feed.com/item-2/

RSS Feed        • Item 1
15 minutes ago  https://rss-feed.com/item-1/

`, out.String())
}

func Test_Feed_CacheControl(t *testing.T) {
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

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=300")
		w.Write([]byte(createDefaultRSS()))
	}))
	defer server.Close()

	err = os.WriteFile(path.Join(listsDir, "default"),
		[]byte(fmt.Sprintf("%d %s\n",
			defaultCurrentTime.Unix(), server.URL,
		),
		), 0600)
	if err != nil {
		t.Fatal(err)
	}

	feed := internal.NewTerminalFeed(timeMock, printer, storage)
	feed.SetAgent("cleed/test")

	root, err := NewRoot("0.1.0", timeMock, printer, storage, feed)
	assert.NoError(t, err)

	os.Args = []string{"cleed"}

	err = root.Cmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, `RSS Feed        • Item 2
1688 days ago   https://rss-feed.com/item-2/

RSS Feed        • Item 1
15 minutes ago  https://rss-feed.com/item-1/

`, out.String())

	cacheInfo, err := storage.LoadCacheInfo()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(cacheInfo))
	assert.Equal(t, &_storage.CacheInfoItem{
		URL:        server.URL,
		LastCheck:  time.Unix(defaultCurrentTime.Unix(), 0),
		ETag:       "",
		FetchAfter: time.Unix(defaultCurrentTime.Unix()+300, 0),
	}, cacheInfo[server.URL])
}

func Test_Feed_RetryAfter(t *testing.T) {
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

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "300")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	err = os.WriteFile(path.Join(listsDir, "default"),
		[]byte(fmt.Sprintf("%d %s\n",
			defaultCurrentTime.Unix(), server.URL,
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
	rss := createDefaultRSS()
	err = storage.SaveFeedCache(bytes.NewBufferString(rss), server.URL)
	if err != nil {
		t.Fatal(err)
	}

	feed := internal.NewTerminalFeed(timeMock, printer, storage)
	feed.SetAgent("cleed/test")

	root, err := NewRoot("0.1.0", timeMock, printer, storage, feed)
	assert.NoError(t, err)

	os.Args = []string{"cleed"}

	err = root.Cmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, `RSS Feed        • Item 2
1688 days ago   https://rss-feed.com/item-2/

RSS Feed        • Item 1
15 minutes ago  https://rss-feed.com/item-1/

`, out.String())

	cacheInfo, err := storage.LoadCacheInfo()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(cacheInfo))
	assert.Equal(t, &_storage.CacheInfoItem{
		URL:        server.URL,
		LastCheck:  time.Unix(0, 0),
		ETag:       "",
		FetchAfter: time.Unix(defaultCurrentTime.Unix()+300, 0),
	}, cacheInfo[server.URL])
}

func Test_Feed_FetchAfter_Load_From_Cache(t *testing.T) {
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
		[]byte(fmt.Sprintf("%d %s\n",
			defaultCurrentTime.Unix(), "https://example.com",
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
	rss := createDefaultRSS()
	err = storage.SaveFeedCache(bytes.NewBufferString(rss), "https://example.com")
	if err != nil {
		t.Fatal(err)
	}

	storage.SaveCacheInfo(map[string]*_storage.CacheInfoItem{
		"https://example.com": {
			URL:        "https://example.com",
			LastCheck:  time.Unix(defaultCurrentTime.Unix(), 0),
			ETag:       "etag",
			FetchAfter: time.Unix(defaultCurrentTime.Unix()+300, 0),
		},
	})

	feed := internal.NewTerminalFeed(timeMock, printer, storage)
	feed.SetAgent("cleed/test")

	root, err := NewRoot("0.1.0", timeMock, printer, storage, feed)
	assert.NoError(t, err)

	os.Args = []string{"cleed"}

	err = root.Cmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, `RSS Feed        Item 2
1688 days ago   https://rss-feed.com/item-2/

RSS Feed        Item 1
15 minutes ago  https://rss-feed.com/item-1/

`, out.String())
}

func Test_Feed_Limit(t *testing.T) {
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

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotModified)
	}))
	defer server.Close()

	err = os.WriteFile(path.Join(listsDir, "default"),
		[]byte(fmt.Sprintf("%d %s\n",
			defaultCurrentTime.Unix(), server.URL,
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
	rss := createDefaultRSS()
	err = storage.SaveFeedCache(bytes.NewBufferString(rss), server.URL)
	if err != nil {
		t.Fatal(err)
	}

	feed := internal.NewTerminalFeed(timeMock, printer, storage)
	feed.SetAgent("cleed/test")

	root, err := NewRoot("0.1.0", timeMock, printer, storage, feed)
	assert.NoError(t, err)

	os.Args = []string{"cleed", "--limit", "1"}

	err = root.Cmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, `RSS Feed        • Item 1
15 minutes ago  https://rss-feed.com/item-1/

`, out.String())
}

func Test_Feed_Since_Period(t *testing.T) {
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

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotModified)
	}))
	defer server.Close()

	err = os.WriteFile(path.Join(listsDir, "default"),
		[]byte(fmt.Sprintf("%d %s\n%d %s\n",
			defaultCurrentTime.Unix(), server.URL+"/rss",
			defaultCurrentTime.Unix(), server.URL+"/atom",
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
	err = storage.SaveFeedCache(bytes.NewBufferString(createDefaultRSS()), server.URL+"/rss")
	if err != nil {
		t.Fatal(err)
	}
	err = storage.SaveFeedCache(bytes.NewBufferString(createDefaultAtom()), server.URL+"/atom")
	if err != nil {
		t.Fatal(err)
	}

	feed := internal.NewTerminalFeed(timeMock, printer, storage)
	feed.SetAgent("cleed/test")

	root, err := NewRoot("0.1.0", timeMock, printer, storage, feed)
	assert.NoError(t, err)

	os.Args = []string{"cleed", "--since", "24h"}

	err = root.Cmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, `Atom Feed       • Item 1
18 hours ago    https://atom-feed.com/item-1/

RSS Feed        • Item 1
15 minutes ago  https://rss-feed.com/item-1/

`, out.String())
}

func Test_Feed_Since_Date(t *testing.T) {
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

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotModified)
	}))
	defer server.Close()

	err = os.WriteFile(path.Join(listsDir, "default"),
		[]byte(fmt.Sprintf("%d %s\n%d %s\n",
			defaultCurrentTime.Unix(), server.URL+"/rss",
			defaultCurrentTime.Unix(), server.URL+"/atom",
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
	err = storage.SaveFeedCache(bytes.NewBufferString(createDefaultRSS()), server.URL+"/rss")
	if err != nil {
		t.Fatal(err)
	}
	err = storage.SaveFeedCache(bytes.NewBufferString(createDefaultAtom()), server.URL+"/atom")
	if err != nil {
		t.Fatal(err)
	}

	feed := internal.NewTerminalFeed(timeMock, printer, storage)
	feed.SetAgent("cleed/test")

	root, err := NewRoot("0.1.0", timeMock, printer, storage, feed)
	assert.NoError(t, err)

	os.Args = []string{"cleed", "--since", "2023-12-31T06:00:00Z"}

	err = root.Cmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, `Atom Feed       • Item 1
18 hours ago    https://atom-feed.com/item-1/

RSS Feed        • Item 1
15 minutes ago  https://rss-feed.com/item-1/

`, out.String())
}

func Test_Feed_Since_Last(t *testing.T) {
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

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotModified)
	}))
	defer server.Close()

	err = os.WriteFile(path.Join(listsDir, "default"),
		[]byte(fmt.Sprintf("%d %s\n%d %s\n",
			defaultCurrentTime.Unix(), server.URL+"/rss",
			defaultCurrentTime.Unix(), server.URL+"/atom",
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
	err = storage.SaveFeedCache(bytes.NewBufferString(createDefaultRSS()), server.URL+"/rss")
	if err != nil {
		t.Fatal(err)
	}
	err = storage.SaveFeedCache(bytes.NewBufferString(createDefaultAtom()), server.URL+"/atom")
	if err != nil {
		t.Fatal(err)
	}

	config := &_storage.Config{
		Version: "0.1.0",
		LastRun: time.Unix(defaultCurrentTime.Unix()-int64(10*60*60), 0), // 10 hours ago
	}
	configPath := path.Join(configDir, "cleed_test", "config.json")
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(configPath, b, 0600)
	if err != nil {
		t.Fatal(err)
	}

	feed := internal.NewTerminalFeed(timeMock, printer, storage)
	feed.SetAgent("cleed/test")

	root, err := NewRoot("0.1.0", timeMock, printer, storage, feed)
	assert.NoError(t, err)

	os.Args = []string{"cleed", "--since", "last"}

	err = root.Cmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, `RSS Feed        • Item 1
15 minutes ago  https://rss-feed.com/item-1/

`, out.String())

	config, err = storage.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, defaultCurrentTime, config.LastRun)
}
