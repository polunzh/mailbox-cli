package model

import "fmt"

func MakeAccountID(provider, email string) string {
	return fmt.Sprintf("%s:%s", provider, email)
}

type Account struct {
	ID          string `json:"id"`
	Provider    string `json:"provider"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	CredKey     string `json:"credKey"`
}

type MessageLocator struct {
	AccountID string `json:"accountId"`
	Provider  string `json:"provider"`
	ID        string `json:"id"`
}

type Message struct {
	Locator    MessageLocator `json:"locator"`
	From       string         `json:"from"`
	Subject    string         `json:"subject"`
	ReceivedAt string         `json:"receivedAt"`
	Unread     bool           `json:"unread"`
}

type MessageDetail struct {
	Locator           MessageLocator `json:"locator"`
	From              string         `json:"from"`
	To                []string       `json:"to"`
	CC                []string       `json:"cc"`
	Subject           string         `json:"subject"`
	ReceivedAt        string         `json:"receivedAt"`
	Unread            bool           `json:"unread"`
	TextBody          string         `json:"textBody"`
	HTMLBody          string         `json:"-"`
	HtmlBodyAvailable bool           `json:"htmlBodyAvailable"`
}

type Draft struct {
	To        []string
	Subject   string
	Body      string
	InReplyTo *MessageLocator
}

type ListOptions struct {
	Limit  int
	Offset int
	Unread bool
}
