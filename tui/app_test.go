package tui_test

import (
	"testing"

	"github.com/zhenqiang/mailbox-cli/internal/model"
	"github.com/zhenqiang/mailbox-cli/tui"
)

func TestAppNoAccounts(t *testing.T) {
	m := tui.NewApp(nil, nil)
	view := m.View()
	if view.Content == "" {
		t.Fatal("expected non-empty view when no accounts configured")
	}
	// Should show onboarding info in the view
	if m.State() != 0 { // viewList
		t.Fatalf("expected initial state to be list, got %d", m.State())
	}
}

func TestAppWithAccounts(t *testing.T) {
	accounts := []model.Account{
		{ID: "gmail:a@example.com", Provider: "gmail", Email: "a@example.com"},
	}
	m := tui.NewApp(accounts, nil)
	view := m.View()
	if view.Content == "" {
		t.Fatal("expected non-empty view when accounts exist")
	}
}

func TestAppInitialState(t *testing.T) {
	m := tui.NewApp(nil, nil)
	view := m.View()
	if view.Content == "" {
		t.Fatal("initial view must return non-empty content")
	}
}
