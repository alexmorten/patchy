package email

import (
	"regexp"
	"strings"
)

var messageIDRegex = regexp.MustCompile(`(?i)Message-Id:\s*<([^>]+)>`)

// ExtractMessageID extracts the Message-ID from email headers.
// Returns the Message-ID if found, empty string if not found.
// Handles both "Message-ID" and "Message-Id" header variants.
func ExtractMessageID(headers string) string {
	match := messageIDRegex.FindStringSubmatch(headers)
	if len(match) > 1 {
		// Trim any whitespace from the captured message ID
		return strings.TrimSpace(match[1])
	}
	return ""
}
