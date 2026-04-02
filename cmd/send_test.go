package cmd_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/polunzh/mailbox-cli/cmd"
	"github.com/polunzh/mailbox-cli/internal/model"
)

func TestSendJSONShape(t *testing.T) {
	loc := &model.MessageLocator{AccountID: "gmail:a@b.com", Provider: "gmail", ID: "sent1"}
	var buf bytes.Buffer
	if err := cmd.WriteSendJSON(&buf, loc, nil); err != nil {
		t.Fatal(err)
	}
	var out struct {
		Sent    bool                 `json:"sent"`
		Locator model.MessageLocator `json:"locator"`
	}
	if err := json.NewDecoder(&buf).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !out.Sent {
		t.Fatal("expected sent=true")
	}
	if out.Locator.ID != "sent1" {
		t.Fatalf("unexpected locator id: %q", out.Locator.ID)
	}
}

func TestReplyJSONShape(t *testing.T) {
	loc := &model.MessageLocator{AccountID: "gmail:a@b.com", Provider: "gmail", ID: "reply1"}
	inReplyTo := &model.MessageLocator{AccountID: "gmail:a@b.com", Provider: "gmail", ID: "orig1"}
	var buf bytes.Buffer
	if err := cmd.WriteSendJSON(&buf, loc, inReplyTo); err != nil {
		t.Fatal(err)
	}
	var out struct {
		Sent      bool                  `json:"sent"`
		Locator   model.MessageLocator  `json:"locator"`
		InReplyTo *model.MessageLocator `json:"inReplyTo"`
	}
	if err := json.NewDecoder(&buf).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.InReplyTo == nil {
		t.Fatal("expected inReplyTo in reply JSON")
	}
	if out.InReplyTo.ID != "orig1" {
		t.Fatalf("unexpected inReplyTo id: %q", out.InReplyTo.ID)
	}
}

func TestBodyResolutionPriority(t *testing.T) {
	// --body flag wins over stdin
	body, err := cmd.ResolveBody("flag", "stdin", false)
	if err != nil {
		t.Fatal(err)
	}
	if body != "flag" {
		t.Fatalf("expected flag-body, got %q", body)
	}
}

func TestSendJSONRejectsEditorBody(t *testing.T) {
	// empty flag + empty stdin + jsonMode → error
	_, err := cmd.ResolveBody("", "", true)
	if err == nil {
		t.Fatal("expected error in JSON mode with no body")
	}
}

func TestReplyRejectsBothIDAndLocator(t *testing.T) {
	_, err := cmd.ParseSelector("id1", `{"accountId":"a","provider":"gmail","id":"id1"}`)
	if err == nil {
		t.Fatal("expected error when both id and --locator provided")
	}
}
