package utils

import (
	"testing"
	_time "time"

	"github.com/stretchr/testify/assert"
)

func Test_ParseDuration(t *testing.T) {
	d, err := ParseDuration("1d1h1m")
	assert.NoError(t, err)
	assert.Equal(t, 24*_time.Hour+_time.Hour+_time.Minute, d)

	d, err = ParseDuration("2d")
	assert.NoError(t, err)
	assert.Equal(t, 48*_time.Hour, d)

	d, err = ParseDuration("12h")
	assert.NoError(t, err)
	assert.Equal(t, 12*_time.Hour, d)

	d, err = ParseDuration("30m")
	assert.NoError(t, err)
	assert.Equal(t, 30*_time.Minute, d)

	d, err = ParseDuration("1")
	assert.EqualError(t, err, "invalid duration: 1")
	assert.Equal(t, _time.Duration(0), d)

	d, err = ParseDuration("1d1h1")
	assert.EqualError(t, err, "invalid duration: 1d1h1")
	assert.Equal(t, _time.Duration(0), d)

	d, err = ParseDuration("1d1h1m1")
	assert.EqualError(t, err, "invalid duration: 1d1h1m1")
	assert.Equal(t, _time.Duration(0), d)

	d, err = ParseDuration("1d1h1m1s")
	assert.EqualError(t, err, "invalid duration: 1d1h1m1s")
	assert.Equal(t, _time.Duration(0), d)
}

func Test_ParseDateTime(t *testing.T) {
	tm, err := ParseDateTime("2024-01-01 12:03:04")
	assert.NoError(t, err)
	assert.Equal(t, _time.Date(2024, 1, 1, 12, 3, 4, 0, _time.Local), tm)

	tm, err = ParseDateTime("2024-01-01 23:45")
	assert.NoError(t, err)
	assert.Equal(t, _time.Date(2024, 1, 1, 23, 45, 0, 0, _time.Local), tm)

	tm, err = ParseDateTime("2024-12-31")
	assert.NoError(t, err)
	assert.Equal(t, _time.Date(2024, 12, 31, 0, 0, 0, 0, _time.Local), tm)

	tm, err = ParseDateTime("2024-12-31T23:59:59Z")
	assert.NoError(t, err)
	assert.Equal(t, _time.Date(2024, 12, 31, 23, 59, 59, 0, _time.UTC), tm)

	tm, err = ParseDateTime("Tue, 13 Aug 2024 08:01:29 GMT")
	assert.NoError(t, err)
	assert.Equal(t, _time.Date(2024, 8, 13, 8, 1, 29, 0, _time.FixedZone("GMT", 0)), tm)

	tm, err = ParseDateTime("Tue, 13 Aug 2024 08:01:29 +0200")
	assert.NoError(t, err)
	assert.Equal(t, _time.Date(2024, 8, 13, 8, 1, 29, 0, _time.FixedZone("", 2*60*60)), tm)
}
