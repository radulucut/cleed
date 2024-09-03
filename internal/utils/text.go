package utils

import (
	"fmt"
	"math"
	"unicode"
)

func Pluralize(count int64, singular string) string {
	if count == 1 {
		return fmt.Sprintf("%d %s", count, singular)
	}
	return fmt.Sprintf("%d %ss", count, singular)
}

func Tokenize(s string, tokens [][]rune) [][]rune {
	var token []rune
	for _, r := range s {
		if !unicode.IsSpace(r) {
			token = append(token, r)
		} else if len(token) > 0 {
			tokens = append(tokens, token)
			token = nil
		}
	}
	if len(token) > 0 {
		tokens = append(tokens, token)
	}
	return tokens
}

func Score(q, b [][]rune) int {
	var score int
	skip := true
	for i := range q {
		best := math.MaxInt
		for j := range b {
			best = min(best, LevenshteinDistance(q[i], b[j]))
		}
		if best <= 2 {
			skip = false
		}
		score += best
	}
	if skip {
		return -1
	}
	return score
}

func LevenshteinDistance(a, b []rune) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}
	if len(a) > len(b) {
		a, b = b, a
	}
	la, lb := len(a), len(b)
	row := make([]int, la+1)
	for i := 1; i <= la; i++ {
		row[i] = i
	}
	for i := 1; i <= lb; i++ {
		prev := i
		for j := 1; j <= la; j++ {
			curr := row[j-1]
			if b[i-1] != a[j-1] {
				curr = min(row[j-1]+1, prev+1, row[j]+1)
			}
			row[j-1] = prev
			prev = curr
		}
		row[la] = prev
	}
	return row[la]
}
