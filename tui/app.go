package tui

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/zhenqiang/mailbox-cli/internal/model"
	"github.com/zhenqiang/mailbox-cli/internal/provider"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type tickMsg struct{}

func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

// msgsLoaded is sent when the initial message fetch completes.
type msgsLoaded struct {
	messages []model.Message
	err      error
}

// detailLoaded is sent when a message detail fetch completes.
type detailLoaded struct {
	detail *model.MessageDetail
	err    error
}

// App is the root Bubble Tea model.
type App struct {
	accounts     []model.Account
	provider     provider.MailProvider
	listView     *ListView
	detailView   *DetailView
	state        viewState
	err          string
	loading      bool
	spinnerFrame int
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
	return tea.Batch(
		func() tea.Msg {
			msgs, err := a.provider.ListMessages(model.ListOptions{Limit: 50})
			return msgsLoaded{messages: msgs, err: err}
		},
		tickCmd(),
	)
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		if a.loading {
			a.spinnerFrame = (a.spinnerFrame + 1) % len(spinnerFrames)
			return a, tickCmd()
		}
		return a, nil

	case msgsLoaded:
		a.loading = false
		if msg.err != nil {
			a.err = fmt.Sprintf("Error loading messages: %v", msg.err)
		} else {
			a.listView = NewListView(msg.messages)
			a.state = viewList
		}
		return a, nil

	case detailLoaded:
		a.loading = false
		if msg.err != nil {
			a.err = fmt.Sprintf("Error loading message: %v", msg.err)
			a.state = viewList
		} else {
			a.detailView = NewDetailView(msg.detail)
			a.state = viewDetail
		}
		return a, nil

	case tea.KeyPressMsg:
		switch a.state {
		case viewList:
			return a.updateList(msg)
		case viewDetail:
			return a.updateDetail(msg)
		}
	}
	return a, nil
}

func (a *App) updateList(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
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
	case "enter":
		if a.listView != nil {
			if sel := a.listView.Selected(); sel != nil {
				loc := sel.Locator
				a.loading = true
				return a, tea.Batch(
					func() tea.Msg {
						detail, err := a.provider.GetMessage(loc)
						return detailLoaded{detail: detail, err: err}
					},
					tickCmd(),
				)
			}
		}
	}
	return a, nil
}

func (a *App) updateDetail(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return a, tea.Quit
	case "esc", "backspace":
		a.state = viewList
		a.detailView = nil
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
		frame := spinnerFrames[a.spinnerFrame]
		content = StyleTitle.Render(frame) + " " + StyleDim.Render("Fetching messages...") + "\n"
	case a.state == viewDetail && a.detailView != nil:
		content = a.detailView.RenderContent() + "\n" + helpDetail()
	case a.listView != nil:
		content = a.listView.Render(80) + "\n" + helpList(a.listView.IsUnreadFilter())
	default:
		content = StyleDim.Render("No messages.") + "\n"
	}
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func helpList(unreadOnly bool) string {
	filter := "u: unread filter"
	if unreadOnly {
		filter = StyleTitle.Render("u: all mail")
	}
	return StyleDim.Render("j/k: navigate  enter: open  "+filter+"  q: quit") + "\n"
}

func helpDetail() string {
	return StyleDim.Render("esc/backspace: back  q: quit") + "\n"
}
