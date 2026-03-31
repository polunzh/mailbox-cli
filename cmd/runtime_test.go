package cmd_test

import (
	"testing"

	"github.com/zhenqiang/mailbox-cli/cmd"
	"github.com/zhenqiang/mailbox-cli/internal/model"
)

// --- Selector tests ---

func TestSelectorRejectsBothIDAndLocator(t *testing.T) {
	_, err := cmd.ParseSelector("msg123", `{"accountId":"a","provider":"gmail","id":"msg123"}`)
	if err == nil {
		t.Fatal("expected error when both <id> and --locator are provided")
	}
}

func TestSelectorAcceptsPositionalID(t *testing.T) {
	sel, err := cmd.ParseSelector("msg123", "")
	if err != nil {
		t.Fatal(err)
	}
	if sel.ID != "msg123" {
		t.Fatalf("unexpected ID: %q", sel.ID)
	}
}

func TestSelectorAcceptsLocatorJSON(t *testing.T) {
	raw := `{"accountId":"gmail:me@example.com","provider":"gmail","id":"abc"}`
	sel, err := cmd.ParseSelector("", raw)
	if err != nil {
		t.Fatal(err)
	}
	if sel.AccountID != "gmail:me@example.com" {
		t.Fatalf("unexpected accountId: %q", sel.AccountID)
	}
	if sel.ID != "abc" {
		t.Fatalf("unexpected id: %q", sel.ID)
	}
}

// --- Body resolution tests ---

func TestBodyResolutionFlagWins(t *testing.T) {
	body, err := cmd.ResolveBody("flag-body", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if body != "flag-body" {
		t.Fatalf("expected flag-body, got %q", body)
	}
}

func TestBodyResolutionJSONRejectsEditor(t *testing.T) {
	// Empty flag + no stdin + jsonMode=true → should error (no editor allowed)
	_, err := cmd.ResolveBody("", "", true)
	if err == nil {
		t.Fatal("expected error in JSON mode with no body source")
	}
}

func TestBodyResolutionStdinOverridesEditor(t *testing.T) {
	body, err := cmd.ResolveBody("", "stdin-content", false)
	if err != nil {
		t.Fatal(err)
	}
	if body != "stdin-content" {
		t.Fatalf("expected stdin-content, got %q", body)
	}
}

// --- MessageLocator from selector ---

func TestSelectorToLocator(t *testing.T) {
	acct := model.Account{ID: "gmail:me@example.com", Provider: "gmail"}
	raw := `{"accountId":"gmail:me@example.com","provider":"gmail","id":"xyz"}`
	sel, err := cmd.ParseSelector("", raw)
	if err != nil {
		t.Fatal(err)
	}
	loc, err := sel.ToLocator(acct)
	if err != nil {
		t.Fatal(err)
	}
	if loc.ID != "xyz" {
		t.Fatalf("unexpected locator ID: %q", loc.ID)
	}
}
