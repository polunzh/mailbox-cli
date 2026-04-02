package account

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/polunzh/mailbox-cli/internal/model"
)

func TestNewStore_LoadsExistingConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	// Create a pre-existing config file
	config := `{"accounts":[{"id":"gmail:test@example.com","provider":"gmail","email":"test@example.com"}],"defaultAccountId":"gmail:test@example.com"}`
	if err := os.WriteFile(path, []byte(config), 0600); err != nil {
		t.Fatal(err)
	}

	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("failed to load existing config: %v", err)
	}

	acct, err := store.GetByID("gmail:test@example.com")
	if err != nil {
		t.Fatalf("failed to get account from loaded config: %v", err)
	}
	if acct.Email != "test@example.com" {
		t.Errorf("unexpected email: %q", acct.Email)
	}
}

func TestNewStore_CreatesNewWhenNotExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent", "config.json")

	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("failed to create new store: %v", err)
	}

	list, err := store.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Errorf("expected empty list for new store, got %d items", len(list))
	}
}

func TestNewStore_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	// Create invalid JSON
	if err := os.WriteFile(path, []byte("not valid json"), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := NewStore(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestAdd_DuplicateAccount(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewStore(filepath.Join(dir, "config.json"))

	acct := model.Account{ID: "gmail:test@example.com", Provider: "gmail", Email: "test@example.com"}
	if err := store.Add(acct); err != nil {
		t.Fatal(err)
	}

	// Try to add again
	err := store.Add(acct)
	if err == nil {
		t.Fatal("expected error for duplicate account")
	}
}

func TestRemove_NotFound(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewStore(filepath.Join(dir, "config.json"))

	err := store.Remove("nonexistent")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestRemove_ClearsDefault(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewStore(filepath.Join(dir, "config.json"))

	acct := model.Account{ID: "gmail:test@example.com", Provider: "gmail", Email: "test@example.com"}
	store.Add(acct)
	store.SetDefault(acct.ID)

	// Remove the default account
	if err := store.Remove(acct.ID); err != nil {
		t.Fatal(err)
	}

	// Verify default is cleared
	_, err := store.GetDefault()
	if err != ErrNotFound {
		t.Fatal("expected default to be cleared after removing default account")
	}
}

func TestGetByID_NotFound(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewStore(filepath.Join(dir, "config.json"))

	_, err := store.GetByID("nonexistent")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestSetDefault_NotFound(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewStore(filepath.Join(dir, "config.json"))

	err := store.SetDefault("nonexistent")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestGetDefault_NotSet(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewStore(filepath.Join(dir, "config.json"))

	_, err := store.GetDefault()
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound when no default set, got %v", err)
	}
}

func TestResolveByEmail_NotFound(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewStore(filepath.Join(dir, "config.json"))

	_, err := store.ResolveByEmail("nonexistent@example.com")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestList_ReturnsCopy(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewStore(filepath.Join(dir, "config.json"))

	acct := model.Account{ID: "gmail:test@example.com", Provider: "gmail", Email: "test@example.com"}
	store.Add(acct)

	list1, _ := store.List()
	list2, _ := store.List()

	// Modify list1
	if len(list1) > 0 {
		list1[0].Email = "modified@example.com"
	}

	// list2 should not be affected
	if list2[0].Email != "test@example.com" {
		t.Error("List() should return a copy of the accounts")
	}
}

func TestStore_PersistsToFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	store1, _ := NewStore(path)
	acct := model.Account{ID: "gmail:test@example.com", Provider: "gmail", Email: "test@example.com"}
	store1.Add(acct)

	// Create new store pointing to same file
	store2, _ := NewStore(path)
	list, _ := store2.List()
	if len(list) != 1 {
		t.Fatalf("expected 1 account after loading from file, got %d", len(list))
	}
	if list[0].ID != acct.ID {
		t.Errorf("expected account ID %q, got %q", acct.ID, list[0].ID)
	}
}
