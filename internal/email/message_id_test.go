package email

import (
	"testing"
)

func TestExtractMessageID(t *testing.T) {
	tests := []struct {
		name     string
		headers  string
		expected string
	}{
		{
			name: "standard Message-ID",
			headers: `Message-ID: <173706974044.1927324.7824600141282028094.stgit@frogsfrogsfrogs>
Date: Thu, 16 Jan 2025 15:23:34 -0800`,
			expected: "173706974044.1927324.7824600141282028094.stgit@frogsfrogsfrogs",
		},
		{
			name: "Message-Id variant",
			headers: `Message-Id: <1234567890.1234567890.1234567890@example.com>
Date: Thu, 16 Jan 2025 15:23:34 -0800`,
			expected: "1234567890.1234567890.1234567890@example.com",
		},
		{
			name: "message-id lowercase",
			headers: `message-id: <lowercase@example.com>
Date: Thu, 16 Jan 2025 15:23:34 -0800`,
			expected: "lowercase@example.com",
		},
		{
			name: "Message-ID with spaces",
			headers: `Message-ID: < 1234567890.1234567890.1234567890@example.com >
Date: Thu, 16 Jan 2025 15:23:34 -0800`,
			expected: "1234567890.1234567890.1234567890@example.com",
		},
		{
			name: "Message-ID with special characters",
			headers: `Message-ID: <test+special@example.com>
Date: Thu, 16 Jan 2025 15:23:34 -0800`,
			expected: "test+special@example.com",
		},
		{
			name: "no Message-ID",
			headers: `Date: Thu, 16 Jan 2025 15:23:34 -0800
Subject: Test Subject`,
			expected: "",
		},
		{
			name:     "empty headers",
			headers:  "",
			expected: "",
		},
		{
			name: "malformed Message-ID",
			headers: `Message-ID: <incomplete
Date: Thu, 16 Jan 2025 15:23:34 -0800`,
			expected: "",
		},
		{
			name: "multiple Message-IDs",
			headers: `Message-ID: <first@example.com>
Message-ID: <second@example.com>
Date: Thu, 16 Jan 2025 15:23:34 -0800`,
			expected: "first@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractMessageID(tt.headers)
			if got != tt.expected {
				t.Errorf("ExtractMessageID() = %v, want %v", got, tt.expected)
			}
		})
	}
}
