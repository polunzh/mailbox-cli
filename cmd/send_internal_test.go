package cmd

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/polunzh/mailbox-cli/internal/model"
)

// --- WriteSendJSON tests ---

func TestWriteSendJSON_WithoutReplyTo(t *testing.T) {
	loc := &model.MessageLocator{AccountID: "gmail:test@example.com", Provider: "gmail", ID: "sent123"}

	var buf bytes.Buffer
	err := WriteSendJSON(&buf, loc, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Sent      bool                 `json:"sent"`
		Locator   model.MessageLocator `json:"locator"`
		InReplyTo *model.MessageLocator `json:"inReplyTo,omitempty"`
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	if !result.Sent {
		t.Error("expected sent=true")
	}
	if result.Locator.ID != "sent123" {
		t.Errorf("expected ID=sent123, got %q", result.Locator.ID)
	}
	if result.InReplyTo != nil {
		t.Error("expected inReplyTo to be nil")
	}
}

func TestWriteSendJSON_WithReplyTo(t *testing.T) {
	loc := &model.MessageLocator{AccountID: "gmail:test@example.com", Provider: "gmail", ID: "reply456"}
	inReplyTo := &model.MessageLocator{AccountID: "gmail:test@example.com", Provider: "gmail", ID: "orig789"}

	var buf bytes.Buffer
	err := WriteSendJSON(&buf, loc, inReplyTo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Sent      bool                  `json:"sent"`
		Locator   model.MessageLocator  `json:"locator"`
		InReplyTo *model.MessageLocator `json:"inReplyTo,omitempty"`
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	if !result.Sent {
		t.Error("expected sent=true")
	}
	if result.InReplyTo == nil {
		t.Fatal("expected inReplyTo to not be nil")
	}
	if result.InReplyTo.ID != "orig789" {
		t.Errorf("expected inReplyTo.ID=orig789, got %q", result.InReplyTo.ID)
	}
}

func TestWriteSendJSON_QQProvider(t *testing.T) {
	loc := &model.MessageLocator{AccountID: "qq:test@qq.com", Provider: "qq", ID: "sent999"}

	var buf bytes.Buffer
	err := WriteSendJSON(&buf, loc, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	locator := result["locator"].(map[string]interface{})
	if locator["provider"] != "qq" {
		t.Errorf("expected provider=qq, got %v", locator["provider"])
	}
}
