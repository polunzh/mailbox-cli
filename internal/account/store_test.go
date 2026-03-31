package account_test

import (
	"path/filepath"
	"testing"

	"github.com/zhenqiang/mailbox-cli/internal/account"
	"github.com/zhenqiang/mailbox-cli/internal/model"
)

func newStore(t *testing.T) *account.Store {
	t.Helper()
	s, err := account.NewStore(filepath.Join(t.TempDir(), "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestAddAndGet(t *testing.T) {
	s := newStore(t)
	a := model.Account{ID: "gmail:a@b.com", Provider: "gmail", Email: "a@b.com"}
	if err := s.Add(a); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetByID("gmail:a@b.com")
	if err != nil {
		t.Fatal(err)
	}
	if got.Email != "a@b.com" {
		t.Fatalf("got %q", got.Email)
	}
}

func TestPersistence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	s, _ := account.NewStore(path)
	s.Add(model.Account{ID: "gmail:a@b.com", Provider: "gmail", Email: "a@b.com"})
	s2, _ := account.NewStore(path)
	if _, err := s2.GetByID("gmail:a@b.com"); err != nil {
		t.Fatal("not persisted")
	}
}

func TestList(t *testing.T) {
	s := newStore(t)
	s.Add(model.Account{ID: "gmail:a@b.com", Provider: "gmail", Email: "a@b.com"})
	s.Add(model.Account{ID: "qq:a@qq.com", Provider: "qq", Email: "a@qq.com"})
	list, err := s.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("got %d accounts", len(list))
	}
}

func TestDefaultAccount(t *testing.T) {
	s := newStore(t)
	s.Add(model.Account{ID: "gmail:a@b.com", Provider: "gmail", Email: "a@b.com"})
	if err := s.SetDefault("gmail:a@b.com"); err != nil {
		t.Fatal(err)
	}
	def, err := s.GetDefault()
	if err != nil {
		t.Fatal(err)
	}
	if def.ID != "gmail:a@b.com" {
		t.Fatalf("got %q", def.ID)
	}
}

func TestResolveByEmail(t *testing.T) {
	s := newStore(t)
	s.Add(model.Account{ID: "gmail:a@b.com", Provider: "gmail", Email: "a@b.com"})
	got, err := s.ResolveByEmail("a@b.com")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "gmail:a@b.com" {
		t.Fatalf("got %q", got.ID)
	}
}

func TestAmbiguousEmail(t *testing.T) {
	s := newStore(t)
	s.Add(model.Account{ID: "gmail:a@b.com", Provider: "gmail", Email: "a@b.com"})
	s.Add(model.Account{ID: "qq:a@b.com", Provider: "qq", Email: "a@b.com"})
	_, err := s.ResolveByEmail("a@b.com")
	if err != account.ErrAmbiguous {
		t.Fatalf("expected ErrAmbiguous, got %v", err)
	}
}

func TestRemove(t *testing.T) {
	s := newStore(t)
	s.Add(model.Account{ID: "gmail:a@b.com", Provider: "gmail", Email: "a@b.com"})
	if err := s.Remove("gmail:a@b.com"); err != nil {
		t.Fatal(err)
	}
	_, err := s.GetByID("gmail:a@b.com")
	if err != account.ErrNotFound {
		t.Fatal("expected ErrNotFound")
	}
}
