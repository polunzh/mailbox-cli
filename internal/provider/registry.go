package provider

import (
	"fmt"

	"github.com/zhenqiang/mailbox-cli/internal/model"
)

// Factory creates a MailProvider given an account and its credential string.
type Factory func(account model.Account, credential string) MailProvider

// Registry maps provider names to their factories.
type Registry struct {
	factories map[string]Factory
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry {
	return &Registry{factories: make(map[string]Factory)}
}

// Register adds a factory for the given provider name. Returns an error if already registered.
func (r *Registry) Register(name string, f Factory) error {
	if _, exists := r.factories[name]; exists {
		return fmt.Errorf("provider %q already registered", name)
	}
	r.factories[name] = f
	return nil
}

// Build constructs a MailProvider for the given provider name.
func (r *Registry) Build(name string, account model.Account, credential string) (MailProvider, error) {
	f, ok := r.factories[name]
	if !ok {
		return nil, fmt.Errorf("unknown provider %q", name)
	}
	return f(account, credential), nil
}
