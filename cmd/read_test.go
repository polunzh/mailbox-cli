package cmd

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/polunzh/mailbox-cli/internal/model"
)

func TestValidateReadFlags(t *testing.T) {
	tests := []struct {
		name      string
		jsonMode  bool
		browser   bool
		wantError bool
	}{
		{
			name:      "valid: json mode only",
			jsonMode:  true,
			browser:   false,
			wantError: false,
		},
		{
			name:      "valid: browser mode only",
			jsonMode:  false,
			browser:   true,
			wantError: false,
		},
		{
			name:      "valid: neither flag",
			jsonMode:  false,
			browser:   false,
			wantError: false,
		},
		{
			name:      "invalid: both flags",
			jsonMode:  true,
			browser:   true,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateReadFlags(tt.jsonMode, tt.browser)
			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestWriteReadJSON(t *testing.T) {
	detail := &model.MessageDetail{
		Locator: model.MessageLocator{
			AccountID: "gmail:test@example.com",
			Provider:  "gmail",
			ID:        "msg123",
		},
		From:              "sender@example.com",
		To:                []string{"recipient@example.com"},
		CC:                []string{"cc@example.com"},
		Subject:           "Test Subject",
		ReceivedAt:        "2024-01-01T00:00:00Z",
		Unread:            true,
		TextBody:          "Test body content",
		HtmlBodyAvailable: true,
	}

	var buf bytes.Buffer
	err := WriteReadJSON(&buf, detail)
	if err != nil {
		t.Fatalf("WriteReadJSON failed: %v", err)
	}

	// Verify JSON structure
	var result struct {
		Message struct {
			Locator           model.MessageLocator `json:"locator"`
			From              string               `json:"from"`
			To                []string             `json:"to"`
			CC                []string             `json:"cc"`
			Subject           string               `json:"subject"`
			ReceivedAt        string               `json:"receivedAt"`
			Unread            bool                 `json:"unread"`
			TextBody          string               `json:"textBody"`
			HTMLBodyAvailable bool                 `json:"htmlBodyAvailable"`
		} `json:"message"`
	}

	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if result.Message.Subject != detail.Subject {
		t.Errorf("subject mismatch: got %q, want %q", result.Message.Subject, detail.Subject)
	}
	if result.Message.From != detail.From {
		t.Errorf("from mismatch: got %q, want %q", result.Message.From, detail.From)
	}
	if result.Message.TextBody != detail.TextBody {
		t.Errorf("textBody mismatch: got %q, want %q", result.Message.TextBody, detail.TextBody)
	}
	if !result.Message.Unread {
		t.Error("expected unread to be true")
	}
	if !result.Message.HTMLBodyAvailable {
		t.Error("expected htmlBodyAvailable to be true")
	}
}
