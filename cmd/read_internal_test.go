package cmd

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/polunzh/mailbox-cli/internal/model"
)

// --- WriteReadJSON tests ---

func TestWriteReadJSON_CompleteDetail(t *testing.T) {
	detail := &model.MessageDetail{
		Locator:           model.MessageLocator{AccountID: "gmail:test@example.com", Provider: "gmail", ID: "msg123"},
		From:              "sender@example.com",
		To:                []string{"recipient1@example.com", "recipient2@example.com"},
		CC:                []string{"cc@example.com"},
		Subject:           "Test Subject",
		ReceivedAt:        "2024-01-15T10:30:00Z",
		Unread:            true,
		TextBody:          "This is the message body",
		HtmlBodyAvailable: true,
	}

	var buf bytes.Buffer
	err := WriteReadJSON(&buf, detail)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

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
		t.Fatalf("failed to decode JSON: %v", err)
	}

	if result.Message.Subject != "Test Subject" {
		t.Errorf("expected subject='Test Subject', got %q", result.Message.Subject)
	}
	if result.Message.From != "sender@example.com" {
		t.Errorf("expected from='sender@example.com', got %q", result.Message.From)
	}
	if len(result.Message.To) != 2 {
		t.Errorf("expected 2 recipients, got %d", len(result.Message.To))
	}
	if len(result.Message.CC) != 1 {
		t.Errorf("expected 1 CC, got %d", len(result.Message.CC))
	}
	if !result.Message.Unread {
		t.Error("expected unread=true")
	}
	if !result.Message.HTMLBodyAvailable {
		t.Error("expected htmlBodyAvailable=true")
	}
	if result.Message.TextBody != "This is the message body" {
		t.Errorf("unexpected textBody: %q", result.Message.TextBody)
	}
}

func TestWriteReadJSON_MinimalDetail(t *testing.T) {
	detail := &model.MessageDetail{
		Locator:  model.MessageLocator{AccountID: "qq:test@qq.com", Provider: "qq", ID: "msg456"},
		From:     "from@qq.com",
		Subject:  "Simple",
		TextBody: "Body",
		Unread:   false,
	}

	var buf bytes.Buffer
	err := WriteReadJSON(&buf, detail)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	msg := result["message"].(map[string]interface{})
	if msg["subject"] != "Simple" {
		t.Errorf("unexpected subject: %v", msg["subject"])
	}
	if msg["unread"] != false {
		t.Error("expected unread=false")
	}
}

// --- validateReadFlags tests ---

func TestValidateReadFlags_ValidCombinations(t *testing.T) {
	tests := []struct {
		name     string
		jsonMode bool
		browser  bool
	}{
		{"neither", false, false},
		{"json only", true, false},
		{"browser only", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateReadFlags(tt.jsonMode, tt.browser)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateReadFlags_InvalidCombination(t *testing.T) {
	err := validateReadFlags(true, true)
	if err == nil {
		t.Error("expected error when both json and browser are true")
	}
}
