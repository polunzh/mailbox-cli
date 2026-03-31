package credential

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

var ErrNotFound = errors.New("credential not found")

type Store interface {
	Set(key, value string) error
	Get(key string) (string, error)
	Delete(key string) error
}

type fileStore struct {
	path string
}

func NewFileStore(path string) Store {
	return &fileStore{path: path}
}

func (s *fileStore) load() (map[string]string, error) {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, err
	}
	m := map[string]string{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *fileStore) save(m map[string]string) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0700); err != nil {
		return err
	}
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0600)
}

func (s *fileStore) Set(key, value string) error {
	m, err := s.load()
	if err != nil {
		return err
	}
	m[key] = value
	return s.save(m)
}

func (s *fileStore) Get(key string) (string, error) {
	m, err := s.load()
	if err != nil {
		return "", err
	}
	val, ok := m[key]
	if !ok {
		return "", ErrNotFound
	}
	return val, nil
}

func (s *fileStore) Delete(key string) error {
	m, err := s.load()
	if err != nil {
		return err
	}
	if _, ok := m[key]; !ok {
		return ErrNotFound
	}
	delete(m, key)
	return s.save(m)
}
