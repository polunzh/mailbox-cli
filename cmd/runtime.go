package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/polunzh/mailbox-cli/internal/account"
	"github.com/polunzh/mailbox-cli/internal/credential"
	"github.com/polunzh/mailbox-cli/internal/model"
	"github.com/polunzh/mailbox-cli/internal/provider"
)

// configDir returns the mailbox config directory, creating it if needed.
func configDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("config dir: %w", err)
	}
	dir := filepath.Join(base, "mailbox")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("config dir: %w", err)
	}
	return dir, nil
}

// loadAccountStore returns the AccountStore backed by the config directory.
func loadAccountStore() (*account.Store, error) {
	dir, err := configDir()
	if err != nil {
		return nil, err
	}
	return account.NewStore(filepath.Join(dir, "config.json"))
}

// loadCredentialStore returns the file-backed CredentialStore.
func loadCredentialStore() (credential.Store, error) {
	dir, err := configDir()
	if err != nil {
		return nil, err
	}
	return credential.NewFileStore(filepath.Join(dir, "credentials.json")), nil
}

// resolveAccount resolves the effective account using --account flag or default.
func resolveAccount(store *account.Store, accountFlag string) (model.Account, error) {
	if accountFlag == "" {
		acct, err := store.GetDefault()
		if err != nil {
			return model.Account{}, fmt.Errorf("%w: run 'mailbox auth login' first", ErrNoDefaultAccount)
		}
		return acct, nil
	}
	// Try canonical ID first.
	acct, err := store.GetByID(accountFlag)
	if err == nil {
		return acct, nil
	}
	// Fall back to email lookup.
	return store.ResolveByEmail(accountFlag)
}

// buildProvider returns a MailProvider for the given account.
func buildProvider(reg *provider.Registry, acct model.Account, credStore credential.Store) (provider.MailProvider, error) {
	cred, err := credStore.Get(acct.CredKey)
	if err != nil {
		return nil, fmt.Errorf("credential for %s: %w", acct.ID, err)
	}
	return reg.Build(acct.Provider, acct, cred)
}

// newRegistry returns a Registry pre-registered with all known providers.
func newRegistry() *provider.Registry {
	reg := provider.NewRegistry()
	_ = reg.Register("gmail", func(a model.Account, cred string) provider.MailProvider {
		// Build an oauth2-aware HTTP client that auto-refreshes the token.
		httpClient, err := refreshGmailToken(cred)
		if err != nil {
			// Fall back to plain token (e.g. in tests with fake token).
			return provider.NewGmailProvider(a, cred)
		}
		return provider.NewGmailProviderWithHTTPClient(a, httpClient, "https://gmail.googleapis.com")
	})
	_ = reg.Register("qq", func(a model.Account, cred string) provider.MailProvider {
		return provider.NewQQProvider(a, cred)
	})
	return reg
}

// --- Selector ---

// Selector is the parsed result of a positional <id> or --locator flag.
type Selector struct {
	model.MessageLocator
	positional bool // true if built from positional ID
}

// ParseSelector validates and parses the message selector.
// Exactly one of positionalID or locatorJSON must be non-empty.
func ParseSelector(positionalID, locatorJSON string) (*Selector, error) {
	if positionalID != "" && locatorJSON != "" {
		return nil, errors.New("provide either <id> or --locator, not both")
	}
	if locatorJSON != "" {
		var loc model.MessageLocator
		if err := json.Unmarshal([]byte(locatorJSON), &loc); err != nil {
			return nil, fmt.Errorf("invalid --locator JSON: %w", err)
		}
		return &Selector{MessageLocator: loc}, nil
	}
	if positionalID != "" {
		return &Selector{MessageLocator: model.MessageLocator{ID: positionalID}, positional: true}, nil
	}
	return nil, errors.New("message selector required: provide <id> or --locator")
}

// ToLocator fills in AccountID and Provider from the account when using positional form.
func (s *Selector) ToLocator(acct model.Account) (model.MessageLocator, error) {
	if !s.positional {
		return s.MessageLocator, nil
	}
	return model.MessageLocator{
		AccountID: acct.ID,
		Provider:  acct.Provider,
		ID:        s.ID,
	}, nil
}

// --- Body resolution ---

// ResolveBody determines the message body from flag > stdinContent > $EDITOR.
// In JSON mode, editor is disallowed; returns error if no body available.
func ResolveBody(flagBody, stdinContent string, jsonMode bool) (string, error) {
	if flagBody != "" {
		return flagBody, nil
	}
	if stdinContent != "" {
		return stdinContent, nil
	}
	if jsonMode {
		return "", errors.New("body required in --json mode: use --body or pipe via stdin")
	}
	// Interactive editor path (non-JSON mode).
	return openEditor()
}

// openEditor opens $EDITOR (or vi) and returns the written content.
func openEditor() (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	tmp, err := os.CreateTemp("", "mailbox-*.txt")
	if err != nil {
		return "", fmt.Errorf("editor: create temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return "", fmt.Errorf("editor: close temp file: %w", err)
	}
	defer func() {
		_ = os.Remove(tmp.Name())
	}()

	if err := runEditor(editor, tmp.Name()); err != nil {
		return "", fmt.Errorf("editor: %w", err)
	}
	content, err := os.ReadFile(tmp.Name())
	if err != nil {
		return "", fmt.Errorf("editor: read result: %w", err)
	}
	return string(content), nil
}

func runEditor(editor, path string) error {
	cmd := exec.Command(editor, path) //nolint:gosec
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
