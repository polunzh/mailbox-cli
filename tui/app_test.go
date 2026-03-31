package tui_test

import (
	"testing"

	"github.com/zhenqiang/mailbox-cli/internal/model"
	"github.com/zhenqiang/mailbox-cli/tui"
)

func TestAppNoAccounts(t *testing.T) {
	m := tui.NewApp(nil)
	msg := m.OnboardingMessage()
	if msg == "" {
		t.Fatal("expected non-empty onboarding message when no accounts configured")
	}
}

func TestAppNoDefaultAccount(t *testing.T) {
	accounts := []model.Account{
		{ID: "gmail:a@example.com", Provider: "gmail", Email: "a@example.com"},
	}
	m := tui.NewApp(accounts)
	// No default → NoDefault guidance should be non-empty.
	if m.OnboardingMessage() != "" {
		t.Fatal("expected no onboarding message when accounts exist")
	}
}

func TestAppInitialState(t *testing.T) {
	m := tui.NewApp(nil)
	view := m.View()
	if view.Content == "" {
		t.Fatal("initial view must return non-empty content")
	}
}
