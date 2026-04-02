package credential_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/polunzh/mailbox-cli/internal/credential"
)

func newFileStore(t *testing.T) credential.Store {
	t.Helper()
	return credential.NewFileStore(filepath.Join(t.TempDir(), "creds.json"))
}

func TestSetAndGet(t *testing.T) {
	s := newFileStore(t)
	if err := s.Set("key1", "secret"); err != nil {
		t.Fatal(err)
	}
	val, err := s.Get("key1")
	if err != nil {
		t.Fatal(err)
	}
	if val != "secret" {
		t.Fatalf("got %q", val)
	}
}

func TestDelete(t *testing.T) {
	s := newFileStore(t)
	if err := s.Set("key1", "secret"); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete("key1"); err != nil {
		t.Fatal(err)
	}
	_, err := s.Get("key1")
	if err != credential.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestReload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "creds.json")
	s1 := credential.NewFileStore(path)
	if err := s1.Set("k", "v"); err != nil {
		t.Fatal(err)
	}
	s2 := credential.NewFileStore(path)
	val, err := s2.Get("k")
	if err != nil {
		t.Fatal(err)
	}
	if val != "v" {
		t.Fatalf("got %q", val)
	}
}

func TestFilePermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "creds.json")
	s := credential.NewFileStore(path)
	if err := s.Set("k", "v"); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("got %o, want 0600", info.Mode().Perm())
	}
}
