package utils

import (
	"html"
	"regexp"
	"strings"

	htmlpkg "golang.org/x/net/html"
)

var htmlTagRe = regexp.MustCompile(`<[^>]*>`)

// PlainTextFromHTML converts typical RSS/Atom HTML fragments to plain text:
// tags removed, HTML entities decoded, whitespace collapsed to single spaces.
func PlainTextFromHTML(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	ctx := &htmlpkg.Node{Type: htmlpkg.ElementNode, Data: "div"}
	nodes, err := htmlpkg.ParseFragment(strings.NewReader(s), ctx)
	if err == nil && len(nodes) > 0 {
		var b strings.Builder
		var walk func(*htmlpkg.Node)
		walk = func(n *htmlpkg.Node) {
			if n.Type == htmlpkg.TextNode {
				b.WriteString(n.Data)
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				walk(c)
			}
		}
		for _, n := range nodes {
			walk(n)
		}
		out := strings.TrimSpace(b.String())
		if out != "" {
			return normalizePlainWhitespace(html.UnescapeString(out))
		}
	}
	stripped := htmlTagRe.ReplaceAllString(s, " ")
	return normalizePlainWhitespace(html.UnescapeString(stripped))
}

func normalizePlainWhitespace(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	// Strip stray spaces before punctuation often left when removing inline tags (e.g. "word </b>.").
	for strings.Contains(s, " .") {
		s = strings.ReplaceAll(s, " .", ".")
	}
	for strings.Contains(s, " ,") {
		s = strings.ReplaceAll(s, " ,", ",")
	}
	return s
}
