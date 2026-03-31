package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/zhenqiang/mailbox-cli/internal/model"
)

// viewState represents which TUI view is active.
type viewState int

const (
	viewList viewState = iota
	viewDetail
	viewCompose
)

// App is the root Bubble Tea model.
type App struct {
	accounts []model.Account
	state    viewState
}

// NewApp creates the root TUI model with pre-loaded accounts.
func NewApp(accounts []model.Account) *App {
	return &App{accounts: accounts}
}

// OnboardingMessage returns a guidance message when no accounts are configured.
// Returns "" when accounts are present.
func (a *App) OnboardingMessage() string {
	if len(a.accounts) == 0 {
		return "No accounts configured. To get started:\n  mailbox auth login --provider gmail\n  mailbox auth login --provider qq --email <qq-address>"
	}
	return ""
}

func (a *App) Init() tea.Cmd {
	return nil
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return a, tea.Quit
		}
	}
	return a, nil
}

func (a *App) View() tea.View {
	var content string
	if onboard := a.OnboardingMessage(); onboard != "" {
		content = onboard + "\n"
	} else {
		content = "[mailbox] " + listViewPlaceholder
	}
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

const listViewPlaceholder = "Loading..."
