// Package config locates, reads and writes the mu configuration file and resolves
// each setting from a flag, the environment or the file, in that order.
package config

import (
	"errors"
	"path/filepath"
	"strings"
)

// LookupEnv reads an environment variable, like os.LookupEnv.
type LookupEnv func(key string) (string, bool)

const (
	configDirName  = "markupp"
	configFileName = "cli.json"
)

// Path returns $XDG_CONFIG_HOME/markupp/cli.json, or ~/.config/markupp/cli.json when
// XDG_CONFIG_HOME is unset or relative. The home comes from HOME, then USERPROFILE.
func Path(lookup LookupEnv) (string, error) {
	base, err := configHome(lookup)
	if err != nil {
		return "", err
	}
	return filepath.Join(base, configDirName, configFileName), nil
}

func configHome(lookup LookupEnv) (string, error) {
	if xdg := envValue(lookup, "XDG_CONFIG_HOME"); filepath.IsAbs(xdg) {
		return xdg, nil
	}
	for _, key := range []string{"HOME", "USERPROFILE"} {
		if home := envValue(lookup, key); home != "" {
			return filepath.Join(home, ".config"), nil
		}
	}
	return "", errors.New("cannot locate the config directory: XDG_CONFIG_HOME, HOME and USERPROFILE are all unset, expected an absolute directory in one of them")
}

func envValue(lookup LookupEnv, key string) string {
	value, _ := lookup(key)
	return strings.TrimSpace(value)
}
