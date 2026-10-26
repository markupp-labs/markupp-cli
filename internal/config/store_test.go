package config

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func tempStore(t *testing.T) *Store {
	t.Helper()
	return NewStore(filepath.Join(t.TempDir(), "markupp", "cli.json"))
}

func TestLoadMissingFileIsEmpty(t *testing.T) {
	got, err := tempStore(t).Load()
	if err != nil || got != (Settings{}) {
		t.Fatalf("Load = %+v, %v, want empty settings", got, err)
	}
}

func TestSaveThenLoadRoundTrips(t *testing.T) {
	store := tempStore(t)
	want := Settings{URL: "http://localhost:8080", Vault: "v1", Auth: &Credential{
		Issuer: "http://localhost:8080", AccessToken: "a", RefreshToken: "r",
		ExpiresAt: time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC),
	}}
	if err := store.Save(want); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.URL != want.URL || got.Vault != want.Vault || got.Auth == nil || *got.Auth != *want.Auth {
		t.Errorf("Load = %+v, want %+v", got, want)
	}
}

func TestSaveIsPrivate(t *testing.T) {
	store := tempStore(t)
	if err := store.Save(Settings{URL: "http://localhost:8080"}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	path, _ := store.Path()
	assertMode(t, path, 0o600)
	assertMode(t, filepath.Dir(path), 0o700|os.ModeDir)
	entries, _ := os.ReadDir(filepath.Dir(path))
	if len(entries) != 1 {
		t.Errorf("config dir holds %d entries, want only cli.json (no temp file left)", len(entries))
	}
}

func assertMode(t *testing.T, path string, want os.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat(%s): %v", path, err)
	}
	if info.Mode() != want {
		t.Errorf("mode of %s = %v, want %v", path, info.Mode(), want)
	}
}

func TestLoadInvalidFileNamesPathButNotContent(t *testing.T) {
	store := tempStore(t)
	path, _ := store.Path()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`{"auth":{"access_token":"secret-tok"`), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := store.Load()
	if err == nil || !strings.Contains(err.Error(), path) || strings.Contains(err.Error(), "secret-tok") {
		t.Fatalf("Load error = %v, want one naming %s without the content", err, path)
	}
}

func TestDefaultStoreWithoutHomeFailsOnUse(t *testing.T) {
	store := DefaultStore(fakeEnv{}.lookup)
	if _, err := store.Load(); err == nil {
		t.Fatal("Load without a config dir succeeded, want an error")
	}
	if err := store.Update(context.Background(), func(*Settings) error { return nil }); err == nil {
		t.Fatal("Update without a config dir succeeded, want an error")
	}
}

func TestUpdateSerializesConcurrentWriters(t *testing.T) {
	store := tempStore(t)
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			err := store.Update(context.Background(), func(s *Settings) error {
				s.Vault += "x"
				return nil
			})
			if err != nil {
				t.Errorf("Update: %v", err)
			}
		})
	}
	wg.Wait()
	got, _ := store.Load()
	if got.Vault != "xxxxxxxx" {
		t.Errorf("Vault = %q after 8 concurrent appends, want 8 x", got.Vault)
	}
}

func TestUpdateDoesNotSaveWhenMutateFails(t *testing.T) {
	store := tempStore(t)
	boom := errors.New("rejected")
	err := store.Update(context.Background(), func(s *Settings) error {
		s.URL = "http://changed"
		return boom
	})
	if !errors.Is(err, boom) {
		t.Fatalf("Update error = %v, want %v", err, boom)
	}
	if got, _ := store.Load(); got.URL != "" {
		t.Errorf("URL = %q, want unchanged", got.URL)
	}
}

func TestLockWaitsForTheHolder(t *testing.T) {
	store := tempStore(t)
	unlock, err := store.Lock(context.Background())
	if err != nil {
		t.Fatalf("Lock: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	if _, err := store.Lock(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("second Lock error = %v, want the deadline", err)
	}
	if err := unlock(); err != nil {
		t.Fatalf("unlock: %v", err)
	}
	unlockAgain, err := store.Lock(context.Background())
	if err != nil {
		t.Fatalf("Lock after release: %v", err)
	}
	_ = unlockAgain()
}

func TestLockTakesOverAStaleLock(t *testing.T) {
	store := tempStore(t)
	path, _ := store.Path()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path+".lock", nil, 0o600); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-StaleAfter - time.Second)
	if err := os.Chtimes(path+".lock", old, old); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	unlock, err := store.Lock(ctx)
	if err != nil {
		t.Fatalf("Lock over a stale lock: %v", err)
	}
	_ = unlock()
}
