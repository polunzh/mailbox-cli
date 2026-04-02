package tui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/zhenqiang/mailbox-cli/internal/model"
	"github.com/zhenqiang/mailbox-cli/internal/provider"
)

// msgsLoaded is sent when message list fetch completes
type msgsLoaded struct {
	messages []model.Message
	err      error
}

// detailLoaded is sent when message detail fetch completes
type detailLoaded struct {
	detail *model.MessageDetail
	err    error
}

// ViewState represents which view is active
type viewState int

const (
	viewList viewState = iota
	viewDetail
	viewCompose
	viewHelp
)

// App is the main TUI application
type App struct {
	accounts       []model.Account
	provider       provider.MailProvider
	state          viewState
	messages       []model.Message
	filtered       []model.Message
	selectedMsg    *model.MessageDetail
	cursor         int
	scrollOffset   int
	detailScroll   int
	loading        bool
	spinnerFrame   int
	err            string
	width          int
	height         int
	showUnreadOnly bool
	statusMsg      string
	statusTime     time.Time
}

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type tickMsg struct{}

func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

// State returns the current view state (for testing)
func (a *App) State() viewState {
	return a.state
}

// NewApp creates the TUI application
func NewApp(accounts []model.Account, p provider.MailProvider) *App {
	return &App{
		accounts: accounts,
		provider: p,
		state:    viewList,
		loading:  p != nil,
		width:    120,
		height:   30,
		cursor:   0,
	}
}

// Init initializes the app
func (a *App) Init() tea.Cmd {
	if a.provider == nil {
		return nil
	}
	return tea.Batch(
		tickCmd(),
		func() tea.Msg {
			msgs, err := a.provider.ListMessages(model.ListOptions{Limit: 50})
			return msgsLoaded{messages: msgs, err: err}
		},
	)
}

// Update handles messages
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		return a, nil

	case tickMsg:
		if a.loading {
			a.spinnerFrame = (a.spinnerFrame + 1) % len(spinnerFrames)
			return a, tickCmd()
		}
		return a, nil

	case msgsLoaded:
		a.loading = false
		if msg.err != nil {
			a.err = msg.err.Error()
		} else {
			a.messages = msg.messages
			a.applyFilter()
		}
		return a, nil

	case detailLoaded:
		a.loading = false
		if msg.err != nil {
			a.setStatus(fmt.Sprintf("Error: %v", msg.err))
			a.state = viewList
		} else {
			a.selectedMsg = msg.detail
			a.detailScroll = 0
			a.state = viewDetail
		}
		return a, nil

	case tea.KeyPressMsg:
		return a.handleKey(msg)

	default:
		return a, nil
	}
}

func (a *App) applyFilter() {
	if a.showUnreadOnly {
		a.filtered = nil
		for _, m := range a.messages {
			if m.Unread {
				a.filtered = append(a.filtered, m)
			}
		}
	} else {
		a.filtered = a.messages
	}
	if a.cursor >= len(a.filtered) {
		a.cursor = max(0, len(a.filtered)-1)
	}
}

func (a *App) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	// Global keys
	switch msg.String() {
	case "ctrl+c":
		return a, tea.Quit
	case "?":
		if a.state != viewHelp {
			a.state = viewHelp
			return a, nil
		}
	}

	switch a.state {
	case viewHelp:
		return a.handleHelpKeys(msg)
	case viewDetail:
		return a.handleDetailKeys(msg)
	case viewCompose:
		return a.handleComposeKeys(msg)
	default:
		return a.handleListKeys(msg)
	}
}

func (a *App) handleListKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	listHeight := a.height - 5

	switch msg.String() {
	case "q":
		return a, tea.Quit
	case "enter", "l", "right":
		if len(a.filtered) > 0 && a.cursor < len(a.filtered) {
			return a.openMessage(a.filtered[a.cursor].Locator)
		}
	case "j", "down":
		if a.cursor < len(a.filtered)-1 {
			a.cursor++
			if a.cursor >= a.scrollOffset+listHeight {
				a.scrollOffset++
			}
		}
	case "k", "up":
		if a.cursor > 0 {
			a.cursor--
			if a.cursor < a.scrollOffset {
				a.scrollOffset--
			}
		}
	case "g":
		a.cursor = 0
		a.scrollOffset = 0
	case "G":
		a.cursor = len(a.filtered) - 1
		if a.cursor < 0 {
			a.cursor = 0
		}
		a.scrollOffset = max(0, len(a.filtered)-listHeight)
	case "u":
		a.showUnreadOnly = !a.showUnreadOnly
		a.applyFilter()
		filter := "all"
		if a.showUnreadOnly {
			filter = "unread"
		}
		a.setStatus(fmt.Sprintf("Filter: %s", filter))
	case "r":
		a.loading = true
		a.setStatus("Refreshing...")
		return a, tea.Batch(tickCmd(), func() tea.Msg {
			msgs, err := a.provider.ListMessages(model.ListOptions{Limit: 50})
			return msgsLoaded{messages: msgs, err: err}
		})
	case "n":
		a.state = viewCompose
		a.setStatus("Compose: use 'mailbox send' command")
	}
	return a, nil
}

func (a *App) handleDetailKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return a, tea.Quit
	case "esc", "h", "left", "backspace":
		a.state = viewList
		a.selectedMsg = nil
	case "j", "down", "ctrl+d":
		a.detailScroll += 3
	case "k", "up", "ctrl+u":
		a.detailScroll = max(0, a.detailScroll-3)
	case "n":
		a.state = viewCompose
	case "r":
		a.setStatus("Reply: use 'mailbox reply' command")
	}
	return a, nil
}

func (a *App) handleHelpKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "?":
		a.state = viewList
	}
	return a, nil
}

func (a *App) handleComposeKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		a.state = viewList
		a.setStatus("")
	}
	return a, nil
}

func (a *App) openMessage(loc model.MessageLocator) (tea.Model, tea.Cmd) {
	a.loading = true
	a.setStatus("Loading...")
	return a, tea.Batch(tickCmd(), func() tea.Msg {
		detail, err := a.provider.GetMessage(loc)
		return detailLoaded{detail: detail, err: err}
	})
}

func (a *App) setStatus(msg string) {
	a.statusMsg = msg
	a.statusTime = time.Now()
}

// View renders the UI
func (a *App) View() tea.View {
	if a.err != "" {
		v := tea.NewView(StyleError.Render(a.err) + "\n\nPress q to quit.\n")
		v.AltScreen = true
		return v
	}

	var content string
	switch a.state {
	case viewHelp:
		content = a.renderHelp()
	case viewCompose:
		content = a.renderCompose()
	case viewDetail:
		content = a.renderDetailView()
	default:
		content = a.renderListView()
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func (a *App) renderListView() string {
	listWidth := min(45, a.width/3)
	detailWidth := a.width - listWidth - 1

	// Left panel - Message list
	listContent := a.renderMessageList(listWidth - 2)
	listTitle := " Inbox "
	if a.showUnreadOnly {
		listTitle = " Unread "
	}
	listPanel := StylePanelActive.
		Width(listWidth - 2).
		Height(a.height - 3).
		Render(StylePanelHeader.Render(listTitle) + "\n" + listContent)

	// Right panel - Preview or placeholder
	var detailContent string
	if a.loading && a.selectedMsg == nil {
		frame := spinnerFrames[a.spinnerFrame]
		detailContent = fmt.Sprintf("\n  %s Loading...", StyleTitle.Render(frame))
	} else {
		detailContent = a.renderPreview(detailWidth - 4)
	}
	detailPanel := StylePanel.
		Width(detailWidth - 2).
		Height(a.height - 3).
		Render(StylePanelHeader.Render(" Preview ") + "\n" + detailContent)

	// Combine panels
	body := lipgloss.JoinHorizontal(lipgloss.Top, listPanel, detailPanel)
	status := a.renderStatusBar()

	return lipgloss.JoinVertical(lipgloss.Left, body, status)
}

func (a *App) renderDetailView() string {
	content := a.renderFullMessage()
	detailPanel := StylePanelActive.
		Width(a.width - 2).
		Height(a.height - 3).
		Render(content)
	status := a.renderStatusBar()
	return lipgloss.JoinVertical(lipgloss.Left, detailPanel, status)
}

func (a *App) renderMessageList(width int) string {
	if len(a.accounts) == 0 {
		return "\n  No accounts.\n  Run: mailbox auth login"
	}
	if a.loading && len(a.filtered) == 0 {
		frame := spinnerFrames[a.spinnerFrame]
		return fmt.Sprintf("\n  %s Loading...", StyleTitle.Render(frame))
	}
	if len(a.filtered) == 0 {
		return "\n  No messages."
	}

	var lines []string
	listHeight := a.height - 6
	end := min(len(a.filtered), a.scrollOffset+listHeight)

	for i := a.scrollOffset; i < end; i++ {
		m := a.filtered[i]
		line := a.renderMessageRow(m, i == a.cursor, width)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func (a *App) renderMessageRow(m model.Message, selected bool, width int) string {
	unread := " "
	if m.Unread {
		unread = StyleUnreadDot.Render("●")
	}

	from := Truncate(m.From, 16)
	subj := Truncate(m.Subject, width-26)
	date := formatDateShort(m.ReceivedAt)

	row := fmt.Sprintf(" %s %s %-18s %s", unread, date, from, subj)

	if selected {
		return StyleListItemSelected.Width(width).Render(row)
	}
	if m.Unread {
		return StyleListItemUnread.Width(width).Render(row)
	}
	return StyleListItem.Width(width).Render(row)
}

func (a *App) renderPreview(width int) string {
	if a.selectedMsg == nil {
		return StyleDim.Render("\n  Select a message to preview")
	}
	m := a.selectedMsg

	header := fmt.Sprintf(
		"%s %s\n%s %s\n",
		StyleBold.Render("From:"), m.From,
		StyleBold.Render("Subj:"), Truncate(m.Subject, width-8),
	)

	body := m.TextBody
	if body == "" {
		body = StyleDim.Render("(No text content)")
	}

	// Truncate preview
	lines := strings.Split(body, "\n")
	maxLines := a.height - 12
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		lines = append(lines, StyleDim.Render("..."))
	}

	return header + "\n" + strings.Join(lines, "\n")
}

func (a *App) renderFullMessage() string {
	if a.selectedMsg == nil {
		return StyleDim.Render("No message selected")
	}
	m := a.selectedMsg

	header := StylePanelHeader.Render(" Message ") + "\n"
	header += fmt.Sprintf(
		"\n %s %s\n %s %s\n %s %s\n",
		StyleBold.Render("From:"), m.From,
		StyleBold.Render("Subject:"), m.Subject,
		StyleBold.Render("Date:"), m.ReceivedAt,
	)
	header += " " + StyleSeparator.Render(strings.Repeat("─", a.width-4)) + "\n"

	body := m.TextBody
	if body == "" {
		body = StyleDim.Render("(No text content)")
	}

	// Apply scrolling
	lines := strings.Split(body, "\n")
	visibleHeight := a.height - 10
	if a.detailScroll > 0 && a.detailScroll >= len(lines) {
		a.detailScroll = len(lines) - 1
	}
	end := min(len(lines), a.detailScroll+visibleHeight)
	if a.detailScroll < len(lines) {
		lines = lines[a.detailScroll:end]
	}

	content := header + "\n" + strings.Join(lines, "\n")

	// Show scroll indicator
	if a.detailScroll > 0 || end < len(strings.Split(m.TextBody, "\n")) {
		scrollInfo := fmt.Sprintf("\n%s Line %d/%d", StyleDim.Render("─"), a.detailScroll+1, len(strings.Split(m.TextBody, "\n")))
		content += scrollInfo
	}

	return content
}

func (a *App) renderStatusBar() string {
	// Account info
	accountInfo := "No account"
	if len(a.accounts) > 0 {
		parts := strings.Split(a.accounts[0].ID, ":")
		if len(parts) == 2 {
			accountInfo = parts[1] // email only
		} else {
			accountInfo = a.accounts[0].ID
		}
	}

	// Message count
	count := fmt.Sprintf("%d msgs", len(a.messages))
	if a.showUnreadOnly {
		unread := 0
		for _, m := range a.messages {
			if m.Unread {
				unread++
			}
		}
		count = fmt.Sprintf("%d/%d unread", unread, len(a.messages))
	}

	// Status/hints
	status := a.statusMsg
	if status == "" {
		switch a.state {
		case viewList:
			status = "?:help"
		case viewDetail:
			status = "esc:back j/k:scroll"
		default:
			status = "esc:close"
		}
	}

	left := StyleStatusKey.Render(" "+accountInfo+" ") + " " + count
	right := " " + status + " "

	pad := a.width - lipgloss.Width(left) - lipgloss.Width(right)
	if pad < 0 {
		pad = 0
	}

	return StyleStatusBar.Width(a.width).Render(left + strings.Repeat(" ", pad) + right)
}

func (a *App) renderHelp() string {
	help := `
` + StyleTitle.Render("Keyboard Shortcuts") + `

` + StyleBold.Render("Navigation") + `
  ` + StyleHelpKey.Render("j/k") + `    ` + StyleHelpDesc.Render("Move down/up") + `
  ` + StyleHelpKey.Render("g/G") + `    ` + StyleHelpDesc.Render("First/last message") + `
  ` + StyleHelpKey.Render("Enter") + `  ` + StyleHelpDesc.Render("Open message") + `
  ` + StyleHelpKey.Render("h/←/Esc") + ` ` + StyleHelpDesc.Render("Go back") + `

` + StyleBold.Render("Actions") + `
  ` + StyleHelpKey.Render("n") + `      ` + StyleHelpDesc.Render("New message") + `
  ` + StyleHelpKey.Render("r") + `      ` + StyleHelpDesc.Render("Refresh / Reply") + `
  ` + StyleHelpKey.Render("u") + `      ` + StyleHelpDesc.Render("Toggle unread filter") + `

` + StyleBold.Render("General") + `
  ` + StyleHelpKey.Render("?") + `      ` + StyleHelpDesc.Render("Toggle help") + `
  ` + StyleHelpKey.Render("q") + `      ` + StyleHelpDesc.Render("Quit") + `

` + StyleDim.Render("Press Esc or ? to close") + `
`

	return StylePanelActive.
		Width(min(50, a.width-4)).
		Height(min(18, a.height-4)).
		Render(help)
}

func (a *App) renderCompose() string {
	content := `
` + StyleTitle.Render("Compose Message") + `

` + StyleBold.Render("To:") + `
  (use command line for now)

` + StyleBold.Render("Subject:") + `
  (use command line for now)

` + StyleDim.Render("Run: mailbox send --to=... --subject=...") + `

` + StyleDim.Render("Press Esc to return") + `
`
	return StylePanelActive.
		Width(min(60, a.width-4)).
		Height(min(12, a.height-4)).
		Render(content)
}

func formatDateShort(t string) string {
	// Try parsing common formats
	formats := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	var parsed time.Time
	var err error
	for _, f := range formats {
		parsed, err = time.Parse(f, t)
		if err == nil {
			break
		}
	}
	if err != nil {
		return t[:min(len(t), 10)]
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	msgDay := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, parsed.Location())

	diff := today.Sub(msgDay)
	switch diff {
	case 0:
		return parsed.Format("15:04")
	case 24 * time.Hour:
		return "Yesterday"
	default:
		return parsed.Format("01/02")
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
