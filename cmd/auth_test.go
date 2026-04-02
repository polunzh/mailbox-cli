package cmd_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/polunzh/mailbox-cli/cmd"
	"github.com/polunzh/mailbox-cli/internal/account"
	"github.com/polunzh/mailbox-cli/internal/model"
)

// buildTestStore creates a fresh account store in a temp dir.
func buildTestStore(t *testing.T, accounts ...model.Account) *account.Store {
	t.Helper()
	dir := t.TempDir()
	store, err := account.NewStore(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range accounts {
		if err := store.Add(a); err != nil {
			t.Fatal(err)
		}
	}
	return store
}

func TestAuthStatusJSONShape(t *testing.T) {
	a := model.Account{ID: "gmail:me@example.com", Provider: "gmail", Email: "me@example.com"}
	store := buildTestStore(t, a)
	_ = store.SetDefault(a.ID)

	var buf bytes.Buffer
	if err := writeAuthStatusJSON(&buf, store); err != nil {
		t.Fatal(err)
	}

	var out struct {
		Accounts []struct {
			ID       string `json:"id"`
			Provider string `json:"provider"`
			Email    string `json:"email"`
		} `json:"accounts"`
		DefaultAccountID string `json:"defaultAccountId"`
	}
	if err := json.NewDecoder(&buf).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out.Accounts) != 1 {
		t.Fatalf("expected 1 account, got %d", len(out.Accounts))
	}
	if out.Accounts[0].ID != a.ID {
		t.Fatalf("unexpected account ID: %q", out.Accounts[0].ID)
	}
	if out.DefaultAccountID != a.ID {
		t.Fatalf("unexpected default: %q", out.DefaultAccountID)
	}
}

func TestAuthUseUpdatesDefault(t *testing.T) {
	a1 := model.Account{ID: "gmail:a@example.com", Provider: "gmail", Email: "a@example.com"}
	a2 := model.Account{ID: "qq:b@qq.com", Provider: "qq", Email: "b@qq.com"}
	store := buildTestStore(t, a1, a2)

	if err := store.SetDefault(a1.ID); err != nil {
		t.Fatal(err)
	}
	if err := store.SetDefault(a2.ID); err != nil {
		t.Fatal(err)
	}
	def, err := store.GetDefault()
	if err != nil {
		t.Fatal(err)
	}
	if def.ID != a2.ID {
		t.Fatalf("expected default=%q, got %q", a2.ID, def.ID)
	}
}

func TestAuthLoginMissingProviderFails(t *testing.T) {
	// Simulate running auth login without --provider flag.
	os.Args = []string{"mailbox", "auth", "login"}
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when --provider is missing")
	}
}

// writeAuthStatusJSON is the testable pure-logic counterpart to the auth status command.
func writeAuthStatusJSON(w *bytes.Buffer, store *account.Store) error {
	accts, err := store.List()
	if err != nil {
		return err
	}
	defaultID := ""
	if def, err := store.GetDefault(); err == nil {
		defaultID = def.ID
	}

	type accountJSON struct {
		ID          string `json:"id"`
		Provider    string `json:"provider"`
		Email       string `json:"email"`
		DisplayName string `json:"displayName"`
	}
	type payload struct {
		Accounts         []accountJSON `json:"accounts"`
		DefaultAccountID string        `json:"defaultAccountId"`
	}

	p := payload{DefaultAccountID: defaultID}
	for _, a := range accts {
		p.Accounts = append(p.Accounts, accountJSON{
			ID:          a.ID,
			Provider:    a.Provider,
			Email:       a.Email,
			DisplayName: a.DisplayName,
		})
	}
	if p.Accounts == nil {
		p.Accounts = []accountJSON{}
	}
	return json.NewEncoder(w).Encode(p)
}
