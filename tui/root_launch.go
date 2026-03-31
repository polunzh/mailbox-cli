package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/zhenqiang/mailbox-cli/internal/account"
	"github.com/zhenqiang/mailbox-cli/internal/credential"
	"github.com/zhenqiang/mailbox-cli/internal/provider"
)

// Launch starts the TUI.
func Launch(accountStore *account.Store, credStore credential.Store, reg *provider.Registry) error {
	accounts, err := accountStore.List()
	if err != nil {
		return fmt.Errorf("tui: load accounts: %w", err)
	}

	if len(accounts) == 0 {
		fmt.Println("No accounts configured. To get started:")
		fmt.Println("  mailbox auth login --provider gmail")
		fmt.Println("  mailbox auth login --provider qq --email <qq-address>")
		return nil
	}

	// Use the default account to drive the initial list view.
	acct, err := accountStore.GetDefault()
	if err != nil {
		// Fall back to first account.
		acct = accounts[0]
	}

	cred, err := credStore.Get(acct.CredKey)
	if err != nil {
		return fmt.Errorf("tui: load credential for %s: %w", acct.ID, err)
	}
	p, err := reg.Build(acct.Provider, acct, cred)
	if err != nil {
		return fmt.Errorf("tui: build provider for %s: %w", acct.ID, err)
	}

	app := NewApp(accounts, p)
	prog := tea.NewProgram(app)
	_, err = prog.Run()
	return err
}
