package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlainTextFromHTML(t *testing.T) {
	assert.Equal(t, "", PlainTextFromHTML(""))
	assert.Equal(t, "hello", PlainTextFromHTML("  hello  "))
	assert.Equal(t, "a b", PlainTextFromHTML("<p>a</p><p>b</p>"))
	assert.Equal(t, "Bold text here.", PlainTextFromHTML("<p>Bold <b>text</b> here.</p>"))
	assert.Equal(t, "Tom & Jerry", PlainTextFromHTML("Tom &amp; Jerry"))
	assert.Equal(t, "plain only", PlainTextFromHTML("plain only"))
}
