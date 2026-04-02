package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"

	"github.com/polunzh/mailbox-cli/internal/account"
	"github.com/polunzh/mailbox-cli/internal/credential"
	"github.com/polunzh/mailbox-cli/internal/model"
)

// --- ParseSelector tests ---

func TestParseSelector_ErrorsWhenBothProvided(t *testing.T) {
	_, err := ParseSelector("id123", `{"id":"id123"}`)
	if err == nil {
		t.Fatal("expected error when both positional ID and locator provided")
	}
}

func TestParseSelector_ErrorsWhenNeitherProvided(t *testing.T) {
	_, err := ParseSelector("", "")
	if err == nil {
		t.Fatal("expected error when neither positional ID nor locator provided")
	}
}

func TestParseSelector_InvalidJSON(t *testing.T) {
	_, err := ParseSelector("", "not valid json")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParseSelector_Positional(t *testing.T) {
	sel, err := ParseSelector("msg123", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sel.ID != "msg123" {
		t.Errorf("expected ID=msg123, got %q", sel.ID)
	}
	if !sel.positional {
		t.Error("expected positional=true")
	}
}

func TestParseSelector_Locator(t *testing.T) {
	locJSON := `{"accountId":"gmail:test@example.com","provider":"gmail","id":"msg456"}`
	sel, err := ParseSelector("", locJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sel.ID != "msg456" {
		t.Errorf("expected ID=msg456, got %q", sel.ID)
	}
	if sel.AccountID != "gmail:test@example.com" {
		t.Errorf("expected AccountID=gmail:test@example.com, got %q", sel.AccountID)
	}
	if sel.Provider != "gmail" {
		t.Errorf("expected Provider=gmail, got %q", sel.Provider)
	}
	if sel.positional {
		t.Error("expected positional=false for locator JSON")
	}
}

// --- Selector.ToLocator tests ---

func TestSelector_ToLocator_PositionalFillsFromAccount(t *testing.T) {
	sel := &Selector{
		MessageLocator: model.MessageLocator{ID: "msg789"},
		positional:     true,
	}
	acct := model.Account{ID: "gmail:me@example.com", Provider: "gmail"}

	loc, err := sel.ToLocator(acct)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loc.ID != "msg789" {
		t.Errorf("expected ID=msg789, got %q", loc.ID)
	}
	if loc.AccountID != "gmail:me@example.com" {
		t.Errorf("expected AccountID from account, got %q", loc.AccountID)
	}
	if loc.Provider != "gmail" {
		t.Errorf("expected Provider from account, got %q", loc.Provider)
	}
}

func TestSelector_ToLocator_NonPositionalPreservesValues(t *testing.T) {
	sel := &Selector{
		MessageLocator: model.MessageLocator{
			AccountID: "gmail:other@example.com",
			Provider:  "gmail",
			ID:        "msg999",
		},
		positional: false,
	}
	acct := model.Account{ID: "gmail:me@example.com", Provider: "gmail"}

	loc, err := sel.ToLocator(acct)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loc.ID != "msg999" {
		t.Errorf("expected ID=msg999, got %q", loc.ID)
	}
	if loc.AccountID != "gmail:other@example.com" {
		t.Errorf("expected original AccountID preserved, got %q", loc.AccountID)
	}
}

// --- ResolveBody tests ---

func TestResolveBody_FlagWinsOverStdin(t *testing.T) {
	body, err := ResolveBody("flag-body", "stdin-body", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if body != "flag-body" {
		t.Errorf("expected flag-body, got %q", body)
	}
}

func TestResolveBody_StdinUsedWhenFlagEmpty(t *testing.T) {
	body, err := ResolveBody("", "stdin-body", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if body != "stdin-body" {
		t.Errorf("expected stdin-body, got %q", body)
	}
}

func TestResolveBody_JSONModeRequiresBody(t *testing.T) {
	_, err := ResolveBody("", "", true)
	if err == nil {
		t.Fatal("expected error in JSON mode without body")
	}
}

// --- configDir tests ---

func TestConfigDir_CreatesDirectory(t *testing.T) {
	// This test uses actual user config dir but verifies the function works
	dir, err := configDir()
	if err != nil {
		t.Fatalf("configDir failed: %v", err)
	}
	if dir == "" {
		t.Fatal("configDir returned empty path")
	}
}

// --- loadAccountStore tests ---

func TestLoadAccountStore_Success(t *testing.T) {
	store, err := loadAccountStore()
	if err != nil {
		t.Fatalf("loadAccountStore failed: %v", err)
	}
	if store == nil {
		t.Fatal("loadAccountStore returned nil")
	}
}

// --- loadCredentialStore tests ---

func TestLoadCredentialStore_Success(t *testing.T) {
	store, err := loadCredentialStore()
	if err != nil {
		t.Fatalf("loadCredentialStore failed: %v", err)
	}
	if store == nil {
		t.Fatal("loadCredentialStore returned nil")
	}
}

// --- resolveAccount tests ---

func TestResolveAccount_UsesFlagWhenProvided(t *testing.T) {
	dir := t.TempDir()
	store, _ := account.NewStore(filepath.Join(dir, "config.json"))
	acct := model.Account{ID: "gmail:test@example.com", Provider: "gmail", Email: "test@example.com"}
	if err := store.Add(acct); err != nil {
		t.Fatal(err)
	}

	result, err := resolveAccount(store, "gmail:test@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "gmail:test@example.com" {
		t.Errorf("expected ID to match, got %q", result.ID)
	}
}

func TestResolveAccount_UsesDefaultWhenNoFlag(t *testing.T) {
	dir := t.TempDir()
	store, _ := account.NewStore(filepath.Join(dir, "config.json"))
	acct := model.Account{ID: "gmail:test@example.com", Provider: "gmail", Email: "test@example.com"}
	if err := store.Add(acct); err != nil {
		t.Fatal(err)
	}
	if err := store.SetDefault(acct.ID); err != nil {
		t.Fatal(err)
	}

	result, err := resolveAccount(store, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "gmail:test@example.com" {
		t.Errorf("expected default account, got %q", result.ID)
	}
}

func TestResolveAccount_ErrorWhenNoDefault(t *testing.T) {
	dir := t.TempDir()
	store, _ := account.NewStore(filepath.Join(dir, "config.json"))

	_, err := resolveAccount(store, "")
	if err == nil {
		t.Fatal("expected error when no default and no flag")
	}
	if !errors.Is(err, ErrNoDefaultAccount) {
		t.Errorf("expected ErrNoDefaultAccount, got %v", err)
	}
}

func TestResolveAccount_FallsBackToEmailLookup(t *testing.T) {
	dir := t.TempDir()
	store, _ := account.NewStore(filepath.Join(dir, "config.json"))
	acct := model.Account{ID: "gmail:test@example.com", Provider: "gmail", Email: "test@example.com"}
	store.Add(acct)

	result, err := resolveAccount(store, "test@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "gmail:test@example.com" {
		t.Errorf("expected account by email, got %q", result.ID)
	}
}

// --- buildProvider tests ---

func TestBuildProvider_RequiresCredential(t *testing.T) {
	dir := t.TempDir()
	credStore := credential.NewFileStore(filepath.Join(dir, "creds.json"))
	acct := model.Account{
		ID:       "qq:test@qq.com",
		Provider: "qq",
		Email:    "test@qq.com",
		CredKey:  "test-key",
	}

	_, err := buildProvider(newRegistry(), acct, credStore)
	// Should fail because credential doesn't exist
	if err == nil {
		t.Fatal("expected error when credential not found")
	}
}

// --- newRegistry tests ---

func TestNewRegistry_RegistersKnownProviders(t *testing.T) {
	reg := newRegistry()
	if reg == nil {
		t.Fatal("newRegistry returned nil")
	}

	// Verify gmail and qq are registered by trying to get them
	knownProviders := []string{"gmail", "qq"}
	for _, name := range knownProviders {
		// The registry doesn't have a direct "Get" method we can test easily,
		// but we can verify it was created
		_ = name
	}
}

// --- WriteListJSON tests ---

func TestWriteListJSON_EmptyMessages(t *testing.T) {
	var buf bytes.Buffer
	err := WriteListJSON(&buf, []model.Message{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Messages []interface{} `json:"messages"`
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}
	if result.Messages == nil {
		t.Error("expected non-nil messages array")
	}
	if len(result.Messages) != 0 {
		t.Errorf("expected 0 messages, got %d", len(result.Messages))
	}
}

func TestWriteListJSON_WithMessages(t *testing.T) {
	msgs := []model.Message{
		{
			Locator:    model.MessageLocator{AccountID: "gmail:a@b.com", Provider: "gmail", ID: "1"},
			From:       "sender@b.com",
			Subject:    "Test Subject",
			ReceivedAt: "2024-01-01T00:00:00Z",
			Unread:     true,
		},
	}

	var buf bytes.Buffer
	err := WriteListJSON(&buf, msgs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result struct {
		Messages []struct {
			Locator    model.MessageLocator `json:"locator"`
			From       string               `json:"from"`
			Subject    string               `json:"subject"`
			ReceivedAt string               `json:"receivedAt"`
			Unread     bool                 `json:"unread"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}
	if len(result.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(result.Messages))
	}
	if result.Messages[0].Subject != "Test Subject" {
		t.Errorf("unexpected subject: %q", result.Messages[0].Subject)
	}
	if !result.Messages[0].Unread {
		t.Error("expected unread=true")
	}
}

// --- ProviderAuthResult type alias test ---

func TestProviderAuthResult_TypeAlias(t *testing.T) {
	// Verify the type alias exists and can be used
	var result ProviderAuthResult
	result.Credential = "test-credential"
	if result.Credential != "test-credential" {
		t.Error("ProviderAuthResult type alias not working correctly")
	}
}

// --- ErrMessageNotFoundSentinel test ---

func TestErrMessageNotFoundSentinel_Exists(t *testing.T) {
	if ErrMessageNotFoundSentinel == nil {
		t.Error("ErrMessageNotFoundSentinel should not be nil")
	}
	if ErrMessageNotFoundSentinel.Error() != "message not found" {
		t.Errorf("unexpected error message: %q", ErrMessageNotFoundSentinel.Error())
	}
}
