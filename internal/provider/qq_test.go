package provider_test

import (
	"testing"
	"time"

	"github.com/polunzh/mailbox-cli/internal/model"
	"github.com/polunzh/mailbox-cli/internal/provider"
)

// fakeIMAPClient implements provider.IMAPClient for testing.
type fakeIMAPClient struct {
	messages []*fakeIMAPMsg
}

type fakeIMAPMsg struct {
	uid     uint32
	subject string
	from    string
	body    string
	unread  bool
}

func (f *fakeIMAPClient) ListMessages(limit int) ([]provider.IMAPMessageSummary, error) {
	var result []provider.IMAPMessageSummary
	for _, m := range f.messages {
		result = append(result, provider.IMAPMessageSummary{
			UID:     m.uid,
			Subject: m.subject,
			From:    m.from,
			Date:    time.Now(),
			Unread:  m.unread,
		})
	}
	return result, nil
}

func (f *fakeIMAPClient) GetMessage(uid uint32) (*provider.IMAPMessageDetail, error) {
	for _, m := range f.messages {
		if m.uid == uid {
			return &provider.IMAPMessageDetail{
				UID:      m.uid,
				Subject:  m.subject,
				From:     m.from,
				TextBody: m.body,
				Date:     time.Now(),
				Unread:   m.unread,
			}, nil
		}
	}
	return nil, provider.ErrMessageNotFound
}

// fakeSMTPClient implements provider.SMTPClient for testing.
type fakeSMTPClient struct {
	sent []model.Draft
}

func (f *fakeSMTPClient) Send(from string, draft model.Draft) error {
	f.sent = append(f.sent, draft)
	return nil
}

func newQQTestAccount() model.Account {
	return model.Account{
		ID:       "qq:me@qq.com",
		Provider: "qq",
		Email:    "me@qq.com",
	}
}

func TestQQListMessages(t *testing.T) {
	imap := &fakeIMAPClient{messages: []*fakeIMAPMsg{
		{uid: 1, subject: "Hello QQ", from: "sender@qq.com", unread: true},
	}}
	smtp := &fakeSMTPClient{}
	p := provider.NewQQProviderWithClients(newQQTestAccount(), imap, smtp)

	msgs, err := p.ListMessages(model.ListOptions{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if msgs[0].Subject != "Hello QQ" {
		t.Fatalf("unexpected subject: %q", msgs[0].Subject)
	}
	if msgs[0].Locator.Provider != "qq" {
		t.Fatalf("expected provider=qq, got %q", msgs[0].Locator.Provider)
	}
	if msgs[0].Locator.AccountID != "qq:me@qq.com" {
		t.Fatalf("unexpected account ID: %q", msgs[0].Locator.AccountID)
	}
}

func TestQQGetMessage(t *testing.T) {
	imap := &fakeIMAPClient{messages: []*fakeIMAPMsg{
		{uid: 42, subject: "Test Subject", from: "a@b.com", body: "body text", unread: false},
	}}
	smtp := &fakeSMTPClient{}
	p := provider.NewQQProviderWithClients(newQQTestAccount(), imap, smtp)

	loc := model.MessageLocator{AccountID: "qq:me@qq.com", Provider: "qq", ID: "42"}
	detail, err := p.GetMessage(loc)
	if err != nil {
		t.Fatal(err)
	}
	if detail.Subject != "Test Subject" {
		t.Fatalf("unexpected subject: %q", detail.Subject)
	}
	if detail.TextBody != "body text" {
		t.Fatalf("unexpected body: %q", detail.TextBody)
	}
}

func TestQQSendMessage(t *testing.T) {
	imap := &fakeIMAPClient{}
	smtp := &fakeSMTPClient{}
	p := provider.NewQQProviderWithClients(newQQTestAccount(), imap, smtp)

	draft := model.Draft{
		To:      []string{"recipient@qq.com"},
		Subject: "Hi",
		Body:    "Hello",
	}
	loc, err := p.SendMessage(draft)
	if err != nil {
		t.Fatal(err)
	}
	if loc.AccountID != "qq:me@qq.com" {
		t.Fatalf("unexpected account ID: %q", loc.AccountID)
	}
	if len(smtp.sent) != 1 {
		t.Fatalf("expected 1 sent message, got %d", len(smtp.sent))
	}
}

func TestQQLocatorMapping(t *testing.T) {
	imap := &fakeIMAPClient{messages: []*fakeIMAPMsg{
		{uid: 7, subject: "Locator Test", from: "x@y.com"},
	}}
	smtp := &fakeSMTPClient{}
	p := provider.NewQQProviderWithClients(newQQTestAccount(), imap, smtp)

	msgs, err := p.ListMessages(model.ListOptions{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	loc := msgs[0].Locator
	if loc.Provider != "qq" {
		t.Fatalf("expected provider=qq, got %q", loc.Provider)
	}
	if loc.ID != "7" {
		t.Fatalf("expected ID=7, got %q", loc.ID)
	}
}
