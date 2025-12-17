package utils

import (
	"strings"
	"unicode"
	"github.com/joaoleau/ezreports/internal/config"
	"golang.org/x/text/unicode/norm"
)

func SlugPanelName(input string, panel bool) string {
	t := norm.NFD.String(input)
	var b strings.Builder
	for _, r := range t {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		b.WriteRune(r)
	}

	slug := strings.ToLower(b.String())
	var out strings.Builder
	lastDash := false

	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			out.WriteRune(r)
			lastDash = false
		} else if !lastDash {
			out.WriteRune('-')
			lastDash = true
		}
	}

	result := strings.Trim(out.String(), "-")
	
	if panel {
		return result + "-panel"
	}
	
	return result + "-report"
}

func IsValidReceiver(receiver string, receivers []config.Receiver) bool {
	for _, r := range receivers {
		if receiver == r.Name {
			return true
		}
	}

	return false
}
