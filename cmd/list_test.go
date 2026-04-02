package cmd_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/zhenqiang/mailbox-cli/cmd"
	"github.com/zhenqiang/mailbox-cli/internal/model"
)

func TestListJSONShape(t *testing.T) {
	msgs := []model.Message{
		{
			Locator: model.MessageLocator{AccountID: "gmail:a@b.com", Provider: "gmail", ID: "1"},
			From:    "sender@b.com",
			Subject: "Hello",
			Unread:  true,
		},
	}
	var buf bytes.Buffer
	if err := cmd.WriteListJSON(&buf, msgs); err != nil {
		t.Fatal(err)
	}
	var out struct {
		Messages []struct {
			Locator model.MessageLocator `json:"locator"`
			Subject string               `json:"subject"`
			Unread  bool                 `json:"unread"`
		} `json:"messages"`
	}
	if err := json.NewDecoder(&buf).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(out.Messages))
	}
	if out.Messages[0].Subject != "Hello" {
		t.Fatalf("unexpected subject: %q", out.Messages[0].Subject)
	}
	if !out.Messages[0].Unread {
		t.Fatal("expected unread=true")
	}
}

func TestReadJSONShape(t *testing.T) {
	detail := &model.MessageDetail{
		Locator:  model.MessageLocator{AccountID: "gmail:a@b.com", Provider: "gmail", ID: "1"},
		From:     "sender@b.com",
		Subject:  "Hello",
		TextBody: "Body text",
		Unread:   false,
	}
	var buf bytes.Buffer
	if err := cmd.WriteReadJSON(&buf, detail); err != nil {
		t.Fatal(err)
	}
	var out struct {
		Message struct {
			Locator  model.MessageLocator `json:"locator"`
			Subject  string               `json:"subject"`
			TextBody string               `json:"textBody"`
		} `json:"message"`
	}
	if err := json.NewDecoder(&buf).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Message.Subject != "Hello" {
		t.Fatalf("unexpected subject: %q", out.Message.Subject)
	}
	if out.Message.TextBody != "Body text" {
		t.Fatalf("unexpected body: %q", out.Message.TextBody)
	}
}

func TestReadRejectsBothIDAndLocator(t *testing.T) {
	_, err := cmd.ParseSelector("id123", `{"accountId":"a","provider":"gmail","id":"id123"}`)
	if err == nil {
		t.Fatal("expected error when both <id> and --locator provided")
	}
}

func TestReadBrowserRejectedInJSONMode(t *testing.T) {
	err := cmd.ValidateReadFlags(true, true) // jsonMode=true, browserFlag=true
	if err == nil {
		t.Fatal("expected error: --browser not allowed in --json mode")
	}
}
