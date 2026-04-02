package provider

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/polunzh/mailbox-cli/internal/model"
)

// ErrMessageNotFound is returned when a message UID does not exist.
var ErrMessageNotFound = errors.New("message not found")

// IMAPMessageSummary is the summary view of an IMAP message.
type IMAPMessageSummary struct {
	UID     uint32
	Subject string
	From    string
	Date    time.Time
	Unread  bool
}

// IMAPMessageDetail is the full content of an IMAP message.
type IMAPMessageDetail struct {
	UID      uint32
	Subject  string
	From     string
	To       []string
	TextBody string
	HTMLBody string
	Date     time.Time
	Unread   bool
}

// IMAPClient abstracts IMAP operations for testing.
type IMAPClient interface {
	ListMessages(limit int) ([]IMAPMessageSummary, error)
	GetMessage(uid uint32) (*IMAPMessageDetail, error)
}

// SMTPClient abstracts SMTP send for testing.
type SMTPClient interface {
	Send(from string, draft model.Draft) error
}

// QQProvider implements MailProvider using IMAP (read) and SMTP (send).
type QQProvider struct {
	account model.Account
	imap    IMAPClient
	smtp    SMTPClient
}

// NewQQProvider creates a production QQ provider connecting to real IMAP/SMTP servers.
// imap and smtp clients are constructed lazily from the account credential.
func NewQQProvider(account model.Account, password string) *QQProvider {
	imap := newRealIMAPClient(account.Email, password)
	smtp := newRealSMTPClient(account.Email, password)
	return &QQProvider{account: account, imap: imap, smtp: smtp}
}

// NewQQProviderWithClients creates a QQ provider with injected clients (for tests).
func NewQQProviderWithClients(account model.Account, imap IMAPClient, smtp SMTPClient) *QQProvider {
	return &QQProvider{account: account, imap: imap, smtp: smtp}
}

func (p *QQProvider) Authenticate() (*AuthResult, error) {
	// QQ auth uses app password; validity is confirmed by a successful IMAP connect.
	return &AuthResult{Credential: ""}, nil
}

func (p *QQProvider) ListMessages(opts model.ListOptions) ([]model.Message, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 20
	}
	summaries, err := p.imap.ListMessages(limit)
	if err != nil {
		return nil, fmt.Errorf("qq: list messages: %w", err)
	}
	msgs := make([]model.Message, 0, len(summaries))
	for _, s := range summaries {
		if opts.Unread && !s.Unread {
			continue
		}
		msgs = append(msgs, model.Message{
			Locator: model.MessageLocator{
				AccountID: p.account.ID,
				Provider:  "qq",
				ID:        strconv.FormatUint(uint64(s.UID), 10),
			},
			From:       s.From,
			Subject:    s.Subject,
			ReceivedAt: s.Date.Format(time.RFC3339),
			Unread:     s.Unread,
		})
	}
	return msgs, nil
}

func (p *QQProvider) GetMessage(loc model.MessageLocator) (*model.MessageDetail, error) {
	uid, err := strconv.ParseUint(loc.ID, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("qq: invalid message ID %q: %w", loc.ID, err)
	}
	detail, err := p.imap.GetMessage(uint32(uid))
	if err != nil {
		return nil, fmt.Errorf("qq: get message: %w", err)
	}
	return &model.MessageDetail{
		Locator:    model.MessageLocator{AccountID: p.account.ID, Provider: "qq", ID: loc.ID},
		From:       detail.From,
		To:         detail.To,
		Subject:    detail.Subject,
		ReceivedAt: detail.Date.Format(time.RFC3339),
		Unread:     detail.Unread,
		TextBody:   detail.TextBody,
		HTMLBody:   detail.HTMLBody,
	}, nil
}

func (p *QQProvider) SendMessage(d model.Draft) (*model.MessageLocator, error) {
	if err := p.smtp.Send(p.account.Email, d); err != nil {
		return nil, fmt.Errorf("qq: send: %w", err)
	}
	return &model.MessageLocator{
		AccountID: p.account.ID,
		Provider:  "qq",
		ID:        fmt.Sprintf("sent:%d", time.Now().UnixMilli()),
	}, nil
}

// --- real client stubs (placeholders; wired in production) ---

type realIMAPClient struct {
	email    string
	password string
}

func newRealIMAPClient(email, password string) IMAPClient {
	return &realIMAPClient{email: email, password: password}
}

func (c *realIMAPClient) ListMessages(limit int) ([]IMAPMessageSummary, error) {
	return nil, errors.New("real IMAP client not implemented")
}

func (c *realIMAPClient) GetMessage(uid uint32) (*IMAPMessageDetail, error) {
	return nil, errors.New("real IMAP client not implemented")
}

type realSMTPClient struct {
	email    string
	password string
}

func newRealSMTPClient(email, password string) SMTPClient {
	return &realSMTPClient{email: email, password: password}
}

func (c *realSMTPClient) Send(from string, draft model.Draft) error {
	host := "smtp.qq.com:465"
	raw := buildRFC822(from, draft)
	_ = host
	_ = raw
	// Full implementation wired in Task 9 production path.
	_ = strings.Contains // suppress import
	return errors.New("real SMTP client not implemented")
}
