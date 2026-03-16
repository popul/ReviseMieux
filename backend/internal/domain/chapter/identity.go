package chapter

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// CanonicalKey computes the canonical identity key for an item.
// Two items are considered identical if they share the same canonical key.
// Normalization: lowercase, trim whitespace, remove accents (NFD decomposition).
func CanonicalKey(term string, itemType ItemType, packID *string) string {
	normalized := normalizeTerm(term)
	pack := ""
	if packID != nil {
		pack = *packID
	}
	return normalized + "|" + string(itemType) + "|" + pack
}

// normalizeTerm applies lowercase, trim, and accent removal.
func normalizeTerm(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	// NFD decomposition splits accented chars into base + combining mark
	// Then we strip combining marks (unicode.Mn = Mark, nonspacing)
	var b strings.Builder
	for _, r := range norm.NFD.String(s) {
		if !unicode.Is(unicode.Mn, r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}
