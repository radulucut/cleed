package storage

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

const (
	cacheInfoFile = "cache_info"
)

type CacheInfoItem struct {
	LastFetch  time.Time
	FetchAfter time.Time
	ETag       string
	URL        string
}

func (s *LocalStorage) LoadCacheInfo() (map[string]*CacheInfoItem, error) {
	cacheinfo := make(map[string]*CacheInfoItem)
	path, err := s.JoinCacheDir(cacheInfoFile)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cacheinfo, nil
		}
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		item, err := parseCacheInfoItem(scanner.Text())
		if err != nil {
			return nil, err
		}
		cacheinfo[item.URL] = item
	}
	return cacheinfo, scanner.Err()
}

func (s *LocalStorage) SaveCacheInfo(cacheinfo map[string]*CacheInfoItem) error {
	path, err := s.JoinCacheDir(cacheInfoFile)
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, item := range cacheinfo {
		_, err = f.Write(getCacheInfoItemLine(item))
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *LocalStorage) SaveFeedCache(r io.Reader, name string) error {
	path, err := s.JoinCacheDir("feed_" + url.QueryEscape(name))
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}

func (s *LocalStorage) OpenFeedCache(name string) (io.ReadCloser, error) {
	path, err := s.JoinCacheDir("feed_" + url.QueryEscape(name))
	if err != nil {
		return nil, err
	}
	return os.Open(path)
}

func (s *LocalStorage) SaveParsedFeedCache(feed *gofeed.Feed, name string) error {
	path, err := s.JoinCacheDir("parsed_" + url.QueryEscape(name))
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	encoder := gob.NewEncoder(f)
	return encoder.Encode(feed)
}

func (s *LocalStorage) LoadParsedFeedCache(name string) (*gofeed.Feed, error) {
	path, err := s.JoinCacheDir("parsed_" + url.QueryEscape(name))
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	decoder := gob.NewDecoder(f)
	var feed gofeed.Feed
	err = decoder.Decode(&feed)
	if err != nil {
		return nil, err
	}
	return &feed, nil
}

func (s *LocalStorage) HasParsedFeedCache(name string) bool {
	path, err := s.JoinCacheDir("parsed_" + url.QueryEscape(name))
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

func (s *LocalStorage) RemoveFeedCaches(names []string) error {
	cacheinfo, err := s.LoadCacheInfo()
	if err != nil {
		return err
	}
	for i := range names {
		delete(cacheinfo, names[i])
	}
	err = s.SaveCacheInfo(cacheinfo)
	if err != nil {
		return err
	}
	for i := range names {
		rawPath, err := s.JoinCacheDir("feed_" + url.QueryEscape(names[i]))
		if err == nil {
			os.Remove(rawPath)
		}
		parsedPath, err := s.JoinCacheDir("parsed_" + url.QueryEscape(names[i]))
		if err == nil {
			os.Remove(parsedPath)
		}
	}
	return nil
}

func getCacheInfoItemLine(item *CacheInfoItem) []byte {
	return []byte(fmt.Sprintf("%s %d %s %d\n",
		item.URL,
		item.LastFetch.Unix(),
		url.QueryEscape(item.ETag),
		item.FetchAfter.Unix()),
	)
}

func parseCacheInfoItem(line string) (*CacheInfoItem, error) {
	parts := strings.Split(line, " ")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid cache info line: %s", line)
	}
	lastCheck, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return nil, err
	}
	etag, err := url.QueryUnescape(parts[2])
	if err != nil {
		return nil, err
	}
	var fetchAfter int64
	if len(parts) == 4 {
		fetchAfter, _ = strconv.ParseInt(parts[3], 10, 64)
	}
	return &CacheInfoItem{
		LastFetch:  time.Unix(lastCheck, 0),
		ETag:       etag,
		URL:        parts[0],
		FetchAfter: time.Unix(fetchAfter, 0),
	}, nil
}
