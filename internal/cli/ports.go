package cli

import (
	"context"

	"github.com/markupp-labs/markupp-cli/internal/config"
)

// ConfigStore is the configuration file as the commands see it. *config.Store
// satisfies it.
type ConfigStore interface {
	Path() (string, error)
	Load() (config.Settings, error)
	Update(ctx context.Context, mutate func(*config.Settings) error) error
}
