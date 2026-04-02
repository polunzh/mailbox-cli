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

// Layout mode for responsive design
const (
	minSplitWidth = 100  // Minimum width to show split pane
	listMinWidth  = 35   // Minimum list width
	listMaxWidth  = 50   // Maximum list width
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
	// Pagination
	pageSize       int
	hasMore        bool
	loadingMore    bool
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
		pageSize: 50,
		hasMore:  true,
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
			msgs, err := a.provider.ListMessages(model.ListOptions{Limit: a.pageSize})
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
		if a.loading || a.loadingMore {
			a.spinnerFrame = (a.spinnerFrame + 1) % len(spinnerFrames)
			return a, tickCmd()
		}
		return a, nil

	case msgsLoaded:
		a.loading = false
		a.loadingMore = false
		if msg.err != nil {
			a.err = msg.err.Error()
		} else {
			if len(msg.messages) < a.pageSize {
				a.hasMore = false
			}
			a.messages = append(a.messages, msg.messages...)
			a.applyFilter()
		}
		return a, nil

	case detailLoaded:
		a.loading = false
		if msg.err != nil {
			a.setStatus(fmt.Sprintf("Error: %v", msg.err))
		} else {
			a.selectedMsg = msg.detail
			a.detailScroll = 0
			// Only switch to detail view on narrow screens
			if a.width < minSplitWidth {
				a.state = viewDetail
			}
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
	listHeight := a.height - 6

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
			// Load more when near bottom
			if a.cursor >= len(a.filtered)-5 && a.hasMore && !a.loadingMore {
				return a.loadMore()
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
		a.messages = nil
		a.hasMore = true
		a.setStatus("Refreshing...")
		return a, tea.Batch(tickCmd(), func() tea.Msg {
			msgs, err := a.provider.ListMessages(model.ListOptions{Limit: a.pageSize})
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
		// Only go back to list if in full-screen detail mode (narrow screen)
		if a.width < minSplitWidth {
			a.state = viewList
		}
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

func (a *App) loadMore() (tea.Model, tea.Cmd) {
	a.loadingMore = true
	a.setStatus("Loading more...")
	return a, tea.Batch(tickCmd(), func() tea.Msg {
		opts := model.ListOptions{
			Limit:  a.pageSize,
			Offset: len(a.messages),
		}
		msgs, err := a.provider.ListMessages(opts)
		return msgsLoaded{messages: msgs, err: err}
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
	default:
		// Use split pane on wide screens, single pane on narrow
		if a.width >= minSplitWidth {
			content = a.renderSplitView()
		} else {
			content = a.renderSingleView()
		}
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// renderSplitView shows list and detail side by side (wide screens)
func (a *App) renderSplitView() string {
	listWidth := min(listMaxWidth, max(listMinWidth, a.width/3))
	detailWidth := a.width - listWidth - 1

	// Left panel - Message list
	listContent := a.renderMessageList(listWidth - 2)
	listTitle := " Inbox "
	if a.showUnreadOnly {
		listTitle = " Unread "
	}

	// Show selected indicator in header
	if a.selectedMsg != nil {
		listTitle = " ● " + strings.TrimSpace(listTitle) + " "
	}

	listPanel := StylePanelActive.
		Width(listWidth - 2).
		Height(a.height - 4).
		Render(StylePanelHeader.Render(listTitle) + "\n" + listContent)

	// Right panel - Detail view
	detailContent := a.renderDetailPanel(detailWidth - 4)
	detailPanel := StylePanel.
		Width(detailWidth - 2).
		Height(a.height - 4).
		Render(StylePanelHeader.Render(" Message ") + "\n" + detailContent)

	// Combine panels
	body := lipgloss.JoinHorizontal(lipgloss.Top, listPanel, detailPanel)
	status := a.renderStatusBar()
	toolbar := a.renderToolbar()

	return lipgloss.JoinVertical(lipgloss.Left, body, status, toolbar)
}

// renderSingleView shows only one panel at a time (narrow screens)
func (a *App) renderSingleView() string {
	var content string
	if a.state == viewDetail && a.selectedMsg != nil {
		content = a.renderDetailPanel(a.width - 4)
		detailPanel := StylePanelActive.
			Width(a.width - 2).
			Height(a.height - 4).
			Render(StylePanelHeader.Render(" Message ") + "\n" + content)
		status := a.renderStatusBar()
		toolbar := a.renderToolbar()
		return lipgloss.JoinVertical(lipgloss.Left, detailPanel, status, toolbar)
	}

	// List view
	listContent := a.renderMessageList(a.width - 2)
	listTitle := " Inbox "
	if a.showUnreadOnly {
		listTitle = " Unread "
	}
	listPanel := StylePanelActive.
		Width(a.width - 2).
		Height(a.height - 4).
		Render(StylePanelHeader.Render(listTitle) + "\n" + listContent)
	status := a.renderStatusBar()
	toolbar := a.renderToolbar()
	return lipgloss.JoinVertical(lipgloss.Left, listPanel, status, toolbar)
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
	listHeight := a.height - 10
	end := min(len(a.filtered), a.scrollOffset+listHeight)

	for i := a.scrollOffset; i < end; i++ {
		m := a.filtered[i]
		line := a.renderMessageRow(m, i == a.cursor, width)
		lines = append(lines, line)
	}

	// Add loading indicator at bottom if loading more
	if a.loadingMore {
		frame := spinnerFrames[a.spinnerFrame]
		lines = append(lines, "\n  "+StyleDim.Render(frame)+" Loading more...")
	} else if a.hasMore && len(a.filtered) > 0 {
		lines = append(lines, "\n  "+StyleDim.Render("Scroll down to load more"))
	}

	return strings.Join(lines, "\n")
}

func (a *App) renderMessageRow(m model.Message, selected bool, width int) string {
	// Fixed column widths
	statusW := 2   // ▶ or space + unread dot
	dateW := 12    // Fixed width for date
	fromW := 20    // Fixed width for sender
	subjW := width - statusW - dateW - fromW - 5 // Remaining for subject (minus spaces)
	if subjW < 10 {
		subjW = 10
	}

	// Build status indicator
	var status string
	if selected {
		status = "▶"
	} else {
		status = " "
	}

	unread := " "
	if m.Unread {
		unread = StyleUnreadDot.Render("●")
	}

	// Format columns with fixed widths
	date := padRight(formatDateShort(m.ReceivedAt), dateW)
	from := padRight(Truncate(m.From, fromW-1), fromW)
	subj := Truncate(m.Subject, subjW)

	// Build row - ensure single line, no wrapping
	row := fmt.Sprintf("%s %s %s %s %s", status, unread, date, from, subj)

	// Apply styles with MaxHeight to prevent wrapping
	if selected {
		return StyleListItemSelected.MaxHeight(1).Render(row)
	}
	if m.Unread {
		return StyleListItemUnread.MaxHeight(1).Render(row)
	}
	return StyleListItem.MaxHeight(1).Render(row)
}

// padRight pads string with spaces to ensure fixed width
func padRight(s string, width int) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

func (a *App) renderDetailPanel(width int) string {
	if a.selectedMsg == nil {
		return StyleDim.Render("\n  Select a message to view\n\n  (Use Enter to open a message)")
	}
	m := a.selectedMsg

	header := fmt.Sprintf(
		"%s %s\n%s %s\n",
		StyleBold.Render("From:"), m.From,
		StyleBold.Render("Subject:"), Truncate(m.Subject, width-10),
	)
	header += " " + StyleSeparator.Render(strings.Repeat("─", width-2)) + "\n"

	body := m.TextBody
	if body == "" {
		body = StyleDim.Render("(No text content)")
	}

	// Apply scrolling
	lines := strings.Split(body, "\n")
	visibleHeight := a.height - 14
	if a.detailScroll > 0 && a.detailScroll >= len(lines) {
		a.detailScroll = len(lines) - 1
	}
	end := min(len(lines), a.detailScroll+visibleHeight)
	if a.detailScroll < len(lines) {
		lines = lines[a.detailScroll:end]
	}

	content := header + "\n" + strings.Join(lines, "\n")

	// Show scroll indicator
	totalLines := len(strings.Split(m.TextBody, "\n"))
	if a.detailScroll > 0 || end < totalLines {
		scrollInfo := fmt.Sprintf("\n\n%s %d/%d lines",
			StyleDim.Render("─"),
			a.detailScroll+min(len(lines), visibleHeight),
			totalLines)
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
			accountInfo = parts[1]
		} else {
			accountInfo = a.accounts[0].ID
		}
	}

	// Message count with pagination info
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
	if a.hasMore {
		count += "+"
	}

	left := StyleStatusKey.Render(" "+accountInfo+" ") + " " + count
	right := " " + a.statusMsg + " "

	pad := a.width - lipgloss.Width(left) - lipgloss.Width(right)
	if pad < 0 {
		pad = 0
	}

	return StyleStatusBar.Width(a.width).Render(left + strings.Repeat(" ", pad) + right)
}

func (a *App) renderToolbar() string {
	var items []string

	switch {
	case a.state == viewDetail || (a.width < minSplitWidth && a.selectedMsg != nil):
		// Detail view toolbar
		items = []string{
			StyleToolbarKey.Render(" j/k "),
			" scroll ",
			StyleToolbarKey.Render(" r "),
			" reply ",
			StyleToolbarKey.Render(" n "),
			" new ",
			StyleToolbarKey.Render(" esc "),
			" back ",
			StyleToolbarKey.Render(" ? "),
			" help ",
			StyleToolbarKey.Render(" q "),
			" quit ",
		}
	default:
		// List view toolbar
		items = []string{
			StyleToolbarKey.Render(" j/k "),
			" nav ",
			StyleToolbarKey.Render(" enter "),
			" open ",
			StyleToolbarKey.Render(" r "),
			" refresh ",
			StyleToolbarKey.Render(" u "),
			" unread ",
			StyleToolbarKey.Render(" ? "),
			" help ",
			StyleToolbarKey.Render(" q "),
			" quit ",
		}
	}

	content := strings.Join(items, "")
	// Pad to fill width
	contentWidth := lipgloss.Width(content)
	if contentWidth < a.width {
		content += strings.Repeat(" ", a.width-contentWidth)
	}

	return StyleToolbar.Width(a.width).Render(content)
}

func (a *App) renderHelp() string {
	help := `
` + StyleTitle.Render("Keyboard Shortcuts") + `

` + StyleBold.Render("Navigation") + `
  ` + StyleHelpKey.Render("j/k") + `    ` + StyleHelpDesc.Render("Move down/up") + `
  ` + StyleHelpKey.Render("g/G") + `    ` + StyleHelpDesc.Render("First/last message") + `
  ` + StyleHelpKey.Render("Enter") + `  ` + StyleHelpDesc.Render("Open message") + `
  ` + StyleHelpKey.Render("h/←/Esc") + ` ` + StyleHelpDesc.Render("Go back (narrow mode)") + `

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
		return padRight(t[:min(len(t), 10)], 10)
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	msgDay := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, parsed.Location())

	diff := today.Sub(msgDay)
	switch diff {
	case 0:
		return parsed.Format("15:04")      // 5 chars
	case 24 * time.Hour:
		return "Yesterday"                 // 8 chars
	default:
		return parsed.Format("Jan 02")     // 6 chars like "Apr 01"
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
