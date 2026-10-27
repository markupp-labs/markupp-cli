package cli

import (
	"context"

	"github.com/markupp-labs/markupp-cli/internal/config"
)

// ConfigStore is the configuration file as the commands see it. *config.Store
// satisfies it. Save and Lock are for the session, which refreshes the login under
// the lock.
type ConfigStore interface {
	Path() (string, error)
	Load() (config.Settings, error)
	Save(settings config.Settings) error
	Lock(ctx context.Context) (unlock func() error, err error)
	Update(ctx context.Context, mutate func(*config.Settings) error) error
}
