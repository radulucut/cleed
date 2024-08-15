package storage

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type ListItem struct {
	AddedAt time.Time
	Address string
}

func (s *LocalStorage) AddToList(urls []string, list string) error {
	path, err := s.joinListsDir(list)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	m := make(map[string]*ListItem)
	err = s.LoadFeedsFromList(m, list)
	if err != nil {
		return err
	}
	now := s.time.Now()
	for _, url := range urls {
		_, ok := m[url]
		if ok {
			continue
		}
		_, err := f.Write(getListItemLine(now, url))
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *LocalStorage) RemoveFromList(urls []string, list string) ([]bool, error) {
	l, err := s.GetFeedsFromList(list)
	if err != nil {
		return nil, err
	}
	if len(l) == 0 {
		return nil, fmt.Errorf("no items in list: %s", list)
	}
	results := make([]bool, len(urls))
	remaining := make([]*ListItem, 0)
	for i := range l {
		remove := false
		for j := range urls {
			if l[i].Address == urls[j] {
				remove = true
				results[j] = true
				break
			}
		}
		if !remove {
			remaining = append(remaining, l[i])
		}
	}
	path, err := s.joinListsDir(list)
	if err != nil {
		return nil, err
	}
	b := new(bytes.Buffer)
	for i := range remaining {
		b.Write(getListItemLine(remaining[i].AddedAt, remaining[i].Address))
	}
	err = os.WriteFile(path, b.Bytes(), 0600)
	if err != nil {
		return nil, err
	}
	s.tidyCachesAfterRemove(urls, list)
	return results, nil
}

func (s *LocalStorage) GetFeedsFromList(list string) ([]*ListItem, error) {
	path, err := s.joinListsDir(list)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	l := make([]*ListItem, 0)
	for scanner.Scan() {
		feedItem, err := parseListItemLine(scanner.Text())
		if err != nil {
			return nil, err
		}
		l = append(l, feedItem)
	}
	return l, scanner.Err()
}

func (s *LocalStorage) LoadFeedsFromList(m map[string]*ListItem, list string) error {
	path, err := s.joinListsDir(list)
	if err != nil {
		return err
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		feedItem, err := parseListItemLine(scanner.Text())
		if err != nil {
			return err
		}
		m[feedItem.Address] = feedItem
	}
	return scanner.Err()
}

func (s *LocalStorage) LoadLists() ([]string, error) {
	dir, err := s.joinListsDir("")
	if err != nil {
		return nil, err
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	lists := make([]string, 0)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		lists = append(lists, file.Name())
	}
	return lists, nil
}

func (s *LocalStorage) tidyCachesAfterRemove(urls []string, list string) {
	lists, err := s.LoadLists()
	if err != nil {
		return
	}
	m := make(map[string]*ListItem)
	if len(lists) != 0 {
		for i := range lists {
			if lists[i] == list {
				continue
			}
			err = s.LoadFeedsFromList(m, lists[i])
			if err != nil {
				return
			}
		}
	}
	feedsToRemove := make([]string, 0)
	for i := range urls {
		_, ok := m[urls[i]]
		if !ok {
			feedsToRemove = append(feedsToRemove, urls[i])
		}
	}
	s.RemoveFeedCaches(feedsToRemove)
}

func getListItemLine(
	addedAt time.Time,
	address string,
) []byte {
	return []byte(fmt.Sprintf("%d %s\n", addedAt.Unix(), address))
}

func parseListItemLine(line string) (*ListItem, error) {
	parts := strings.Split(line, " ")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid feed list item: %s", line)
	}
	addedAt, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, err
	}
	return &ListItem{
		AddedAt: time.Unix(addedAt, 0),
		Address: parts[1],
	}, nil
}
