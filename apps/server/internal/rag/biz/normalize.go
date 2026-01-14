package biz

import (
	"strings"
	"unicode"
)

func normalizeQuery(text string) string {
	if strings.TrimSpace(text) == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(text))
	space := false
	for _, r := range strings.ToLower(text) {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			b.WriteRune(r)
			space = false
			continue
		}
		if unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsSymbol(r) {
			if !space {
				b.WriteRune(' ')
				space = true
			}
			continue
		}
	}
	return strings.TrimSpace(b.String())
}
