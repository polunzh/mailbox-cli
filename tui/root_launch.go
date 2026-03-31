package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/zhenqiang/mailbox-cli/internal/account"
)

// Launch starts the TUI. If no accounts are configured, it prints guidance and exits.
func Launch(store *account.Store) error {
	accounts, err := store.List()
	if err != nil {
		return fmt.Errorf("tui: load accounts: %w", err)
	}

	app := NewApp(accounts)

	// No accounts → print onboarding and exit without launching full TUI.
	if msg := app.OnboardingMessage(); msg != "" {
		fmt.Println(msg)
		return nil
	}

	p := tea.NewProgram(app)
	_, err = p.Run()
	return err
}
