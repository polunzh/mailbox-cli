package provider_test

import (
	"testing"

	"github.com/zhenqiang/mailbox-cli/internal/model"
	"github.com/zhenqiang/mailbox-cli/internal/provider"
)

type fakeProvider struct{}

func (f *fakeProvider) ListMessages(opts model.ListOptions) ([]model.Message, error) {
	return nil, nil
}
func (f *fakeProvider) GetMessage(loc model.MessageLocator) (*model.MessageDetail, error) {
	return nil, nil
}
func (f *fakeProvider) SendMessage(d model.Draft) (*model.MessageLocator, error) { return nil, nil }
func (f *fakeProvider) Authenticate() (*provider.AuthResult, error)              { return nil, nil }

func TestRegistryRegisterAndBuild(t *testing.T) {
	r := provider.NewRegistry()
	r.Register("fake", func(a model.Account, cred string) provider.MailProvider {
		return &fakeProvider{}
	})
	p, err := r.Build("fake", model.Account{}, "cred")
	if err != nil {
		t.Fatal(err)
	}
	if p == nil {
		t.Fatal("expected provider")
	}
}

func TestRegistryUnknownProvider(t *testing.T) {
	r := provider.NewRegistry()
	_, err := r.Build("unknown", model.Account{}, "")
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestRegistryDuplicateRegistration(t *testing.T) {
	r := provider.NewRegistry()
	factory := func(a model.Account, cred string) provider.MailProvider { return &fakeProvider{} }
	if err := r.Register("fake", factory); err != nil {
		t.Fatal(err)
	}
	if err := r.Register("fake", factory); err == nil {
		t.Fatal("expected error for duplicate registration")
	}
}
