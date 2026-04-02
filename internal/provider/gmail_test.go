package provider_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zhenqiang/mailbox-cli/internal/model"
	"github.com/zhenqiang/mailbox-cli/internal/provider"
)

// gmailListResponse mirrors the Gmail API messages.list response shape.
type gmailListResponse struct {
	Messages []struct {
		ID string `json:"id"`
	} `json:"messages"`
}

// gmailMessage mirrors the Gmail API messages.get response shape (minimal).
type gmailMessage struct {
	ID      string `json:"id"`
	Payload struct {
		Headers []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"headers"`
		Parts []struct {
			MimeType string `json:"mimeType"`
			Body     struct {
				Data string `json:"data"`
			} `json:"body"`
		} `json:"parts"`
		Body struct {
			Data string `json:"data"`
		} `json:"body"`
		MimeType string `json:"mimeType"`
	} `json:"payload"`
	LabelIDs     []string `json:"labelIds"`
	InternalDate string   `json:"internalDate"`
}

// buildGmailTestServer returns a test server responding with canned Gmail API responses.
func buildGmailTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	// messages.list
	mux.HandleFunc("/gmail/v1/users/me/messages", func(w http.ResponseWriter, r *http.Request) {
		resp := gmailListResponse{}
		resp.Messages = append(resp.Messages, struct {
			ID string `json:"id"`
		}{ID: "msg1"})
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// messages.get
	mux.HandleFunc("/gmail/v1/users/me/messages/msg1", func(w http.ResponseWriter, r *http.Request) {
		msg := gmailMessage{
			ID:           "msg1",
			LabelIDs:     []string{"UNREAD"},
			InternalDate: "1700000000000",
		}
		msg.Payload.MimeType = "text/plain"
		msg.Payload.Headers = []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		}{
			{Name: "From", Value: "sender@example.com"},
			{Name: "Subject", Value: "Hello Gmail"},
			{Name: "To", Value: "me@example.com"},
			{Name: "Date", Value: "Mon, 14 Nov 2023 22:13:20 +0000"},
		}
		// base64url-encoded "Hello Gmail body"
		msg.Payload.Body.Data = "SGVsbG8gR21haWwgYm9keQ=="
		if err := json.NewEncoder(w).Encode(msg); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// messages.send
	mux.HandleFunc("/gmail/v1/users/me/messages/send", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, `{"id":"sent1"}`)
	})

	return httptest.NewServer(mux)
}

func newGmailTestAccount() model.Account {
	return model.Account{
		ID:       "gmail:me@example.com",
		Provider: "gmail",
		Email:    "me@example.com",
	}
}

func TestGmailListMessages(t *testing.T) {
	srv := buildGmailTestServer(t)
	defer srv.Close()

	p := provider.NewGmailProviderWithBaseURL(newGmailTestAccount(), "fake-token", srv.URL)
	msgs, err := p.ListMessages(model.ListOptions{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) == 0 {
		t.Fatal("expected at least one message")
	}
	if msgs[0].Locator.ID != "msg1" {
		t.Fatalf("unexpected locator ID: %q", msgs[0].Locator.ID)
	}
	if msgs[0].Locator.AccountID != "gmail:me@example.com" {
		t.Fatalf("unexpected account ID: %q", msgs[0].Locator.AccountID)
	}
}

func TestGmailGetMessage(t *testing.T) {
	srv := buildGmailTestServer(t)
	defer srv.Close()

	p := provider.NewGmailProviderWithBaseURL(newGmailTestAccount(), "fake-token", srv.URL)
	loc := model.MessageLocator{AccountID: "gmail:me@example.com", Provider: "gmail", ID: "msg1"}
	detail, err := p.GetMessage(loc)
	if err != nil {
		t.Fatal(err)
	}
	if detail.Subject != "Hello Gmail" {
		t.Fatalf("unexpected subject: %q", detail.Subject)
	}
	if detail.From != "sender@example.com" {
		t.Fatalf("unexpected from: %q", detail.From)
	}
	if detail.TextBody == "" {
		t.Fatal("expected non-empty text body")
	}
	if !detail.Unread {
		t.Fatal("expected message to be unread")
	}
}

func TestGmailSendMessage(t *testing.T) {
	srv := buildGmailTestServer(t)
	defer srv.Close()

	p := provider.NewGmailProviderWithBaseURL(newGmailTestAccount(), "fake-token", srv.URL)
	draft := model.Draft{
		To:      []string{"recipient@example.com"},
		Subject: "Test",
		Body:    "Hello",
	}
	loc, err := p.SendMessage(draft)
	if err != nil {
		t.Fatal(err)
	}
	if loc.ID != "sent1" {
		t.Fatalf("unexpected sent ID: %q", loc.ID)
	}
	if loc.AccountID != "gmail:me@example.com" {
		t.Fatalf("unexpected account ID: %q", loc.AccountID)
	}
}

func TestGmailLocatorMapping(t *testing.T) {
	srv := buildGmailTestServer(t)
	defer srv.Close()

	p := provider.NewGmailProviderWithBaseURL(newGmailTestAccount(), "fake-token", srv.URL)
	msgs, err := p.ListMessages(model.ListOptions{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) == 0 {
		t.Fatal("no messages")
	}
	loc := msgs[0].Locator
	if loc.Provider != "gmail" {
		t.Fatalf("expected provider=gmail, got %q", loc.Provider)
	}
	if loc.AccountID != newGmailTestAccount().ID {
		t.Fatalf("locator account mismatch: %q", loc.AccountID)
	}
}
