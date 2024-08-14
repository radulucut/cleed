package utils

import (
	"fmt"
	"math"
	"strconv"
	_time "time"
)

type Time interface {
	Now() _time.Time
}

type time struct{}

func NewTime() Time {
	return &time{}
}

func (t *time) Now() _time.Time {
	return _time.Now()
}

func Relative(seconds int64) string {
	var s string
	if seconds < 60 {
		s = Pluralize(seconds, "second")
	} else if seconds < 3600 {
		s = Pluralize(int64(math.Round(float64(seconds)/60)), "minute")
	} else if seconds < 86400 {
		s = Pluralize(int64(math.Round(float64(seconds)/3600)), "hour")
	} else {
		s = Pluralize(int64(math.Round(float64(seconds)/86400)), "day")
	}
	if seconds >= 0 {
		return s + " ago"
	}
	return "in " + s
}

var durationMap = map[string]_time.Duration{
	"m": _time.Minute,
	"h": _time.Hour,
	"d": _time.Hour * 24,
}

// Only supports minutes, hours, and days
func ParseDuration(s string) (_time.Duration, error) {
	l := 0
	d := _time.Duration(0)
	for i := 0; i < len(s); i++ {
		if v, ok := durationMap[string(s[i])]; ok {
			n, err := strconv.Atoi(s[l:i])
			if err != nil {
				return 0, fmt.Errorf("invalid duration: %s", s)
			}
			d += _time.Duration(n) * v
			l = i + 1
		}
	}
	if l < len(s) {
		return 0, fmt.Errorf("invalid duration: %s", s)
	}
	return d, nil
}

var dateTimeFormats = []string{
	"2006-01-02 15:04:05",
	"2006-01-02 15:04",
	"2006-01-02",
}

func ParseDateTime(s string) (_time.Time, error) {
	for i := range dateTimeFormats {
		t, err := _time.ParseInLocation(dateTimeFormats[i], s, _time.Local)
		if err == nil {
			return t, nil
		}
	}
	t, err := _time.Parse(_time.RFC3339, s)
	if err == nil {
		return t, nil
	}
	t, err = _time.Parse(_time.RFC1123, s)
	if err == nil {
		return t, nil
	}
	t, err = _time.Parse(_time.RFC1123Z, s)
	if err == nil {
		return t, nil
	}
	return _time.Time{}, fmt.Errorf("invalid datetime: %s", s)
}
