package config

import (
	"path/filepath"
	"strings"
	"testing"
)

type fakeEnv map[string]string

func (e fakeEnv) lookup(key string) (string, bool) {
	value, ok := e[key]
	return value, ok
}

func TestPathUsesAbsoluteXDGConfigHome(t *testing.T) {
	env := fakeEnv{"XDG_CONFIG_HOME": "/xdg", "HOME": "/home/u"}
	got, err := Path(env.lookup)
	if err != nil {
		t.Fatalf("Path: %v", err)
	}
	if want := filepath.Join("/xdg", "markupp", "cli.json"); got != want {
		t.Errorf("Path = %q, want %q", got, want)
	}
}

func TestPathIgnoresRelativeXDGConfigHome(t *testing.T) {
	env := fakeEnv{"XDG_CONFIG_HOME": "relative", "HOME": "/home/u"}
	got, err := Path(env.lookup)
	if err != nil {
		t.Fatalf("Path: %v", err)
	}
	if want := filepath.Join("/home/u", ".config", "markupp", "cli.json"); got != want {
		t.Errorf("Path = %q, want %q", got, want)
	}
}

func TestPathFallsBackToUserProfile(t *testing.T) {
	env := fakeEnv{"USERPROFILE": "/users/u"}
	got, err := Path(env.lookup)
	if err != nil {
		t.Fatalf("Path: %v", err)
	}
	if want := filepath.Join("/users/u", ".config", "markupp", "cli.json"); got != want {
		t.Errorf("Path = %q, want %q", got, want)
	}
}

func TestPathWithoutHomeNamesTheVariables(t *testing.T) {
	_, err := Path(fakeEnv{}.lookup)
	if err == nil || !strings.Contains(err.Error(), "XDG_CONFIG_HOME") || !strings.Contains(err.Error(), "HOME") {
		t.Fatalf("Path error = %v, want one naming XDG_CONFIG_HOME and HOME", err)
	}
}
