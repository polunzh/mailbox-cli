package model_test

import (
	"encoding/json"
	"testing"

	"github.com/polunzh/mailbox-cli/internal/model"
)

func TestAccountID(t *testing.T) {
	a := model.Account{Provider: "gmail", Email: "you@gmail.com"}
	a.ID = model.MakeAccountID(a.Provider, a.Email)
	if a.ID != "gmail:you@gmail.com" {
		t.Fatalf("got %q, want %q", a.ID, "gmail:you@gmail.com")
	}
}

func TestMessageLocatorJSON(t *testing.T) {
	loc := model.MessageLocator{
		AccountID: "gmail:you@gmail.com",
		Provider:  "gmail",
		ID:        "abc123",
	}
	b, _ := json.Marshal(loc)
	want := `{"accountId":"gmail:you@gmail.com","provider":"gmail","id":"abc123"}`
	if string(b) != want {
		t.Fatalf("got %s, want %s", b, want)
	}
}
