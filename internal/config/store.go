package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

const (
	dirMode  = 0o700
	fileMode = 0o600
)

// Store reads and writes the configuration file. Writes go to a private temporary file
// in the same directory, synced and renamed over the old one, so a reader never sees
// half a file.
type Store struct {
	path    string
	pathErr error
	now     func() time.Time
}

// NewStore returns a Store for the file at path.
func NewStore(path string) *Store {
	return &Store{path: path, now: time.Now}
}

// DefaultStore returns the Store at Path(lookup). When the path cannot be found the
// Store still builds, so mu starts, and every use returns that error.
func DefaultStore(lookup LookupEnv) *Store {
	path, err := Path(lookup)
	return &Store{path: path, pathErr: err, now: time.Now}
}

// Path returns the location of the configuration file.
func (s *Store) Path() (string, error) { return s.path, s.pathErr }

// Load reads the file. A missing file is empty settings, not an error.
func (s *Store) Load() (Settings, error) {
	if s.pathErr != nil {
		return Settings{}, s.pathErr
	}
	raw, err := os.ReadFile(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		return Settings{}, nil
	}
	if err != nil {
		return Settings{}, fmt.Errorf("reading config file %s: %w", s.path, err)
	}
	return decodeSettings(s.path, raw)
}

// decodeSettings names the file but never quotes it, since it holds tokens.
func decodeSettings(path string, raw []byte) (Settings, error) {
	var settings Settings
	if err := json.Unmarshal(raw, &settings); err != nil {
		return Settings{}, fmt.Errorf("config file %s is not valid, expected a JSON object with url, tenant, vault and auth; fix it or delete it: %w", path, err)
	}
	return settings, nil
}

// Save writes settings atomically with mode 0600 in a 0700 directory.
func (s *Store) Save(settings Settings) error {
	if s.pathErr != nil {
		return s.pathErr
	}
	raw, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding config for %s: %w", s.path, err)
	}
	if err := os.MkdirAll(filepath.Dir(s.path), dirMode); err != nil {
		return fmt.Errorf("creating config directory %s: %w", filepath.Dir(s.path), err)
	}
	return writeAtomic(s.path, append(raw, '\n'))
}

func writeAtomic(path string, content []byte) (err error) {
	temp, err := os.CreateTemp(filepath.Dir(path), ".cli-*.json.tmp")
	if err != nil {
		return fmt.Errorf("creating temporary file next to %s: %w", path, err)
	}
	defer func() {
		if err != nil {
			_ = os.Remove(temp.Name())
		}
	}()
	if err = writeAndClose(temp, content); err != nil {
		return err
	}
	if err = os.Rename(temp.Name(), path); err != nil {
		return fmt.Errorf("replacing config file %s: %w", path, err)
	}
	return nil
}

func writeAndClose(temp *os.File, content []byte) error {
	err := errors.Join(temp.Chmod(fileMode), writeSynced(temp, content))
	if closeErr := temp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return fmt.Errorf("writing temporary config file %s: %w", temp.Name(), err)
	}
	return nil
}

func writeSynced(temp *os.File, content []byte) error {
	if _, err := temp.Write(content); err != nil {
		return err
	}
	return temp.Sync()
}

// Update runs mutate on the current settings under the lock and saves the result.
// When mutate fails nothing is written.
func (s *Store) Update(ctx context.Context, mutate func(*Settings) error) error {
	unlock, err := s.Lock(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = unlock() }()
	settings, err := s.Load()
	if err != nil {
		return err
	}
	if err := mutate(&settings); err != nil {
		return err
	}
	return s.Save(settings)
}
