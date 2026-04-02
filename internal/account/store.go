package account

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/polunzh/mailbox-cli/internal/model"
)

var (
	ErrNotFound  = errors.New("account not found")
	ErrAmbiguous = errors.New("ambiguous account: multiple accounts match that email")
)

type configFile struct {
	Accounts         []model.Account `json:"accounts"`
	DefaultAccountID string          `json:"defaultAccountId"`
}

type Store struct {
	path string
	cfg  configFile
}

func NewStore(path string) (*Store, error) {
	s := &Store{path: path}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		s.cfg = configFile{}
		return nil
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &s.cfg)
}

func (s *Store) save() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s.cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0600)
}

func (s *Store) Add(a model.Account) error {
	for _, existing := range s.cfg.Accounts {
		if existing.ID == a.ID {
			return errors.New("account already exists: " + a.ID)
		}
	}
	s.cfg.Accounts = append(s.cfg.Accounts, a)
	return s.save()
}

func (s *Store) Remove(id string) error {
	for i, a := range s.cfg.Accounts {
		if a.ID == id {
			s.cfg.Accounts = append(s.cfg.Accounts[:i], s.cfg.Accounts[i+1:]...)
			if s.cfg.DefaultAccountID == id {
				s.cfg.DefaultAccountID = ""
			}
			return s.save()
		}
	}
	return ErrNotFound
}

func (s *Store) GetByID(id string) (model.Account, error) {
	for _, a := range s.cfg.Accounts {
		if a.ID == id {
			return a, nil
		}
	}
	return model.Account{}, ErrNotFound
}

func (s *Store) List() ([]model.Account, error) {
	out := make([]model.Account, len(s.cfg.Accounts))
	copy(out, s.cfg.Accounts)
	return out, nil
}

func (s *Store) SetDefault(id string) error {
	if _, err := s.GetByID(id); err != nil {
		return err
	}
	s.cfg.DefaultAccountID = id
	return s.save()
}

func (s *Store) GetDefault() (model.Account, error) {
	if s.cfg.DefaultAccountID == "" {
		return model.Account{}, ErrNotFound
	}
	return s.GetByID(s.cfg.DefaultAccountID)
}

func (s *Store) ResolveByEmail(email string) (model.Account, error) {
	var matches []model.Account
	for _, a := range s.cfg.Accounts {
		if a.Email == email {
			matches = append(matches, a)
		}
	}
	switch len(matches) {
	case 0:
		return model.Account{}, ErrNotFound
	case 1:
		return matches[0], nil
	default:
		return model.Account{}, ErrAmbiguous
	}
}
