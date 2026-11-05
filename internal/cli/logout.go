package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newLogoutCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Revoke the stored login and remove it from the config file",
		Long: `Revoke the stored login on its server and remove it from the config file.

The login is removed even when the server cannot be reached; mu then says so, and
the authorization can be revoked in the web panel. MARKUPP_TOKEN is not affected.`,
		Example: "  mu logout",
		Args:    exactArgs(),
		RunE:    app.logout,
	}
}

func (a *App) logout(cmd *cobra.Command, _ []string) error {
	removed, err := a.authenticator().Logout(cmd.Context())
	if err != nil {
		return err
	}
	message := "Not logged in; nothing to remove."
	if removed {
		message = "Logged out: the login was revoked and removed from the config file."
	}
	_, _ = fmt.Fprintln(a.deps.Stderr, message)
	return nil
}
