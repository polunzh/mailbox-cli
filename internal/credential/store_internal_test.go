package credential

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileStore_Get_NotFound(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(filepath.Join(dir, "creds.json"))

	_, err := store.Get("nonexistent-key")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestFileStore_Get_EmptyStore(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(filepath.Join(dir, "creds.json"))

	_, err := store.Get("any-key")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound for empty store, got %v", err)
	}
}

func TestFileStore_Set_UpdatesExisting(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(filepath.Join(dir, "creds.json"))

	// Set initial value
	if err := store.Set("key1", "value1"); err != nil {
		t.Fatal(err)
	}

	// Update the value
	if err := store.Set("key1", "value2"); err != nil {
		t.Fatal(err)
	}

	// Verify the update
	val, err := store.Get("key1")
	if err != nil {
		t.Fatal(err)
	}
	if val != "value2" {
		t.Errorf("expected 'value2', got %q", val)
	}
}

func TestFileStore_Delete_NotFound(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(filepath.Join(dir, "creds.json"))

	err := store.Delete("nonexistent-key")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestFileStore_MultipleKeys(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(filepath.Join(dir, "creds.json"))

	// Set multiple keys
	keys := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	for k, v := range keys {
		if err := store.Set(k, v); err != nil {
			t.Fatal(err)
		}
	}

	// Verify all keys
	for k, expected := range keys {
		val, err := store.Get(k)
		if err != nil {
			t.Fatalf("failed to get %q: %v", k, err)
		}
		if val != expected {
			t.Errorf("key %q: expected %q, got %q", k, expected, val)
		}
	}
}

func TestFileStore_Delete_OneOfMany(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(filepath.Join(dir, "creds.json"))

	// Set multiple keys
	_ = store.Set("key1", "value1")
	_ = store.Set("key2", "value2")
	_ = store.Set("key3", "value3")

	// Delete one key
	if err := store.Delete("key2"); err != nil {
		t.Fatal(err)
	}

	// Verify deleted key is gone
	_, err := store.Get("key2")
	if err != ErrNotFound {
		t.Fatal("expected deleted key to not be found")
	}

	// Verify other keys still exist
	val1, _ := store.Get("key1")
	val3, _ := store.Get("key3")
	if val1 != "value1" || val3 != "value3" {
		t.Error("other keys should not be affected by delete")
	}
}

func TestFileStore_FileCreated(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "creds.json")
	store := NewFileStore(path)

	// Set should create the file
	if err := store.Set("key", "value"); err != nil {
		t.Fatal(err)
	}

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("credential file was not created")
	}
}

func TestFileStore_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "creds.json")

	// Create file with invalid JSON
	if err := os.WriteFile(path, []byte("not valid json"), 0600); err != nil {
		t.Fatal(err)
	}

	store := NewFileStore(path)
	_, err := store.Get("any-key")
	if err == nil {
		t.Fatal("expected error when loading invalid JSON")
	}
}

func TestFileStore_LoadsExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "creds.json")

	// Pre-populate the file
	content := `{"existing-key":"existing-value"}`
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	store := NewFileStore(path)
	val, err := store.Get("existing-key")
	if err != nil {
		t.Fatalf("failed to load existing credential: %v", err)
	}
	if val != "existing-value" {
		t.Errorf("expected 'existing-value', got %q", val)
	}
}

func TestFileStore_Set_CorruptedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "creds.json")

	// Create file with invalid JSON
	if err := os.WriteFile(path, []byte("corrupted"), 0600); err != nil {
		t.Fatal(err)
	}

	store := NewFileStore(path)
	// Try to set a key - should fail when loading
	err := store.Set("key", "value")
	if err == nil {
		t.Fatal("expected error when saving to corrupted file")
	}
}

func TestFileStore_Delete_CorruptedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "creds.json")

	// Create file with invalid JSON
	if err := os.WriteFile(path, []byte("corrupted"), 0600); err != nil {
		t.Fatal(err)
	}

	store := NewFileStore(path)
	// Try to delete - should fail when loading
	err := store.Delete("key")
	if err == nil {
		t.Fatal("expected error when deleting from corrupted file")
	}
}
