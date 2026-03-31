package provider

import "github.com/zhenqiang/mailbox-cli/internal/model"

// AuthResult holds credentials returned after authentication.
type AuthResult struct {
	Credential string // serialized credential (token JSON, password, etc.)
}

// MailProvider is the interface all email backends must implement.
type MailProvider interface {
	Authenticate() (*AuthResult, error)
	ListMessages(opts model.ListOptions) ([]model.Message, error)
	GetMessage(loc model.MessageLocator) (*model.MessageDetail, error)
	SendMessage(d model.Draft) (*model.MessageLocator, error)
}
