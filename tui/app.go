package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/zhenqiang/mailbox-cli/internal/model"
	"github.com/zhenqiang/mailbox-cli/internal/provider"
)

// msgsLoaded is sent when the initial message fetch completes.
type msgsLoaded struct {
	messages []model.Message
	err      error
}

// App is the root Bubble Tea model.
type App struct {
	accounts []model.Account
	provider provider.MailProvider
	listView *ListView
	state    viewState
	err      string
	loading  bool
}

// viewState represents which TUI view is active.
type viewState int

const (
	viewList viewState = iota
	viewDetail
	viewCompose
)

// NewApp creates the root TUI model with pre-loaded accounts.
// p may be nil when no accounts are configured.
func NewApp(accounts []model.Account, p provider.MailProvider) *App {
	return &App{
		accounts: accounts,
		provider: p,
		loading:  p != nil,
	}
}

// OnboardingMessage returns a guidance message when no accounts are configured.
func (a *App) OnboardingMessage() string {
	if len(a.accounts) == 0 {
		return "No accounts configured. To get started:\n  mailbox auth login --provider gmail\n  mailbox auth login --provider qq --email <qq-address>"
	}
	return ""
}

func (a *App) Init() tea.Cmd {
	if a.provider == nil {
		return nil
	}
	// Kick off async message load.
	return func() tea.Msg {
		msgs, err := a.provider.ListMessages(model.ListOptions{Limit: 50})
		return msgsLoaded{messages: msgs, err: err}
	}
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case msgsLoaded:
		a.loading = false
		if msg.err != nil {
			a.err = fmt.Sprintf("Error loading messages: %v", msg.err)
		} else {
			a.listView = NewListView(msg.messages)
			a.state = viewList
		}
		return a, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return a, tea.Quit
		case "j", "down":
			if a.listView != nil {
				a.listView.MoveDown()
			}
		case "k", "up":
			if a.listView != nil {
				a.listView.MoveUp()
			}
		case "u":
			if a.listView != nil {
				a.listView.SetUnreadFilter(!a.listView.IsUnreadFilter())
			}
		}
	}
	return a, nil
}

func (a *App) View() tea.View {
	var content string
	switch {
	case a.OnboardingMessage() != "":
		content = a.OnboardingMessage() + "\n"
	case a.err != "":
		content = StyleError.Render(a.err) + "\n\nPress q to quit.\n"
	case a.loading:
		content = StyleDim.Render("Loading messages...") + "\n"
	case a.listView != nil:
		content = a.listView.Render(80)
	default:
		content = StyleDim.Render("No messages.") + "\n"
	}
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}
