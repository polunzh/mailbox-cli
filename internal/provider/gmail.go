package provider

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/polunzh/mailbox-cli/internal/model"
)

// GmailProvider implements MailProvider against the Gmail REST API.
type GmailProvider struct {
	account model.Account
	token   string
	baseURL string
	client  *http.Client
}

// NewGmailProvider creates a GmailProvider using the real Gmail API base URL.
// httpClient should be an oauth2-aware client that handles token refresh automatically.
// Pass nil to use a plain client with the token string as Bearer (for tests only).
func NewGmailProvider(account model.Account, token string) *GmailProvider {
	return NewGmailProviderWithBaseURL(account, token, "https://gmail.googleapis.com")
}

// NewGmailProviderWithHTTPClient creates a GmailProvider with a pre-built HTTP client.
// Use this in production to pass an oauth2-aware client.
func NewGmailProviderWithHTTPClient(account model.Account, httpClient *http.Client, baseURL string) *GmailProvider {
	return &GmailProvider{
		account: account,
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  httpClient,
	}
}

// NewGmailProviderWithBaseURL creates a GmailProvider with a configurable base URL (for tests).
func NewGmailProviderWithBaseURL(account model.Account, token, baseURL string) *GmailProvider {
	return &GmailProvider{
		account: account,
		token:   token,
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *GmailProvider) Authenticate() (*AuthResult, error) {
	// OAuth flow is handled externally; this provider expects a valid token.
	return &AuthResult{Credential: p.token}, nil
}

func (p *GmailProvider) ListMessages(opts model.ListOptions) ([]model.Message, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 20
	}
	url := fmt.Sprintf("%s/gmail/v1/users/me/messages?maxResults=%d", p.baseURL, limit)
	if opts.Unread {
		url += "&q=is:unread"
	}

	body, err := p.get(url)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Messages []struct {
			ID string `json:"id"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("gmail: decode list: %w", err)
	}

	msgs := make([]model.Message, 0, len(resp.Messages))
	for _, m := range resp.Messages {
		detail, err := p.fetchMessage(m.ID)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, model.Message{
			Locator: model.MessageLocator{
				AccountID: p.account.ID,
				Provider:  "gmail",
				ID:        m.ID,
			},
			From:       detail.From,
			Subject:    detail.Subject,
			ReceivedAt: detail.ReceivedAt,
			Unread:     detail.Unread,
		})
	}
	return msgs, nil
}

func (p *GmailProvider) GetMessage(loc model.MessageLocator) (*model.MessageDetail, error) {
	return p.fetchMessage(loc.ID)
}

func (p *GmailProvider) SendMessage(d model.Draft) (*model.MessageLocator, error) {
	raw := buildRFC822(p.account.Email, d)
	encoded := base64.URLEncoding.EncodeToString([]byte(raw))
	payload := fmt.Sprintf(`{"raw":"%s"}`, encoded)

	url := fmt.Sprintf("%s/gmail/v1/users/me/messages/send", p.baseURL)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if p.token != "" {
		req.Header.Set("Authorization", "Bearer "+p.token)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gmail: send: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("gmail: send %d: %s", resp.StatusCode, body)
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("gmail: decode send response: %w", err)
	}
	return &model.MessageLocator{
		AccountID: p.account.ID,
		Provider:  "gmail",
		ID:        result.ID,
	}, nil
}

// fetchMessage retrieves a single message by Gmail ID.
func (p *GmailProvider) fetchMessage(id string) (*model.MessageDetail, error) {
	url := fmt.Sprintf("%s/gmail/v1/users/me/messages/%s", p.baseURL, id)
	body, err := p.get(url)
	if err != nil {
		return nil, err
	}

	var msg struct {
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
	if err := json.Unmarshal(body, &msg); err != nil {
		return nil, fmt.Errorf("gmail: decode message: %w", err)
	}

	headers := make(map[string]string)
	for _, h := range msg.Payload.Headers {
		headers[strings.ToLower(h.Name)] = h.Value
	}

	// Parse From: strip display name, keep address.
	from := parseAddress(headers["from"])

	to := parseAddressList(headers["to"])

	var textBody string
	if msg.Payload.MimeType == "text/plain" {
		textBody = decodeBase64(msg.Payload.Body.Data)
	} else {
		for _, part := range msg.Payload.Parts {
			if part.MimeType == "text/plain" {
				textBody = decodeBase64(part.Body.Data)
				break
			}
		}
	}

	unread := false
	for _, l := range msg.LabelIDs {
		if l == "UNREAD" {
			unread = true
			break
		}
	}

	receivedAt := headers["date"]

	return &model.MessageDetail{
		Locator:    model.MessageLocator{AccountID: p.account.ID, Provider: "gmail", ID: id},
		From:       from,
		To:         to,
		Subject:    headers["subject"],
		ReceivedAt: receivedAt,
		Unread:     unread,
		TextBody:   textBody,
	}, nil
}

func (p *GmailProvider) get(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	// Only set Authorization manually when using the plain token fallback.
	// oauth2-aware clients inject the header automatically.
	if p.token != "" {
		req.Header.Set("Authorization", "Bearer "+p.token)
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gmail: get %s: %w", url, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("gmail: %d from %s: %s", resp.StatusCode, url, body)
	}
	return body, nil
}

func decodeBase64(s string) string {
	// Gmail uses URL-safe base64 without padding.
	s = strings.ReplaceAll(s, "-", "+")
	s = strings.ReplaceAll(s, "_", "/")
	// Add padding.
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return s // return raw on failure
	}
	return string(b)
}

func parseAddress(raw string) string {
	// "Display Name <email>" → "email"
	if i := strings.Index(raw, "<"); i >= 0 {
		if j := strings.Index(raw[i:], ">"); j >= 0 {
			return raw[i+1 : i+j]
		}
	}
	return strings.TrimSpace(raw)
}

func parseAddressList(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		out = append(out, parseAddress(strings.TrimSpace(p)))
	}
	return out
}

func buildRFC822(from string, d model.Draft) string {
	var sb strings.Builder
	sb.WriteString("From: " + from + "\r\n")
	sb.WriteString("To: " + strings.Join(d.To, ", ") + "\r\n")
	sb.WriteString("Subject: " + d.Subject + "\r\n")
	sb.WriteString("MIME-Version: 1.0\r\n")
	sb.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	sb.WriteString("\r\n")
	sb.WriteString(d.Body)
	return sb.String()
}
