package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/markupp-labs/markupp-cli/internal/output"
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// NotificationReader reads the notification feed of the account.
type NotificationReader interface {
	Notifications(ctx context.Context) (io.ReadCloser, error)
}

const notificationsLong = `List what asks for the attention of the account, oldest first: today the
pending invites to other tenants, which mu invites accept or decline. The number of
unread items goes to stderr, so stdout keeps only the items. Items of a kind this
version of mu does not know are left out of the text output and kept in --json.`

func newNotificationsCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:     "notifications",
		Short:   "List the notifications of the account",
		Long:    requiresServer(featureNotifications, notificationsLong),
		Example: "  mu notifications\n  mu notifications --json 2>/dev/null | jq -r '.items[] | select(.kind == \"invite\") | .id'",
		Args:    exactArgs(),
		RunE:    app.showNotifications,
	}
}

var notificationColumns = []output.Column{
	{Name: "KIND", Width: 6}, {Name: "ID", Width: 36}, {Name: "ROLE", Width: 6}, {Name: "TENANT", Width: 24}, {Name: "EXPIRES"},
}

func (a *App) showNotifications(cmd *cobra.Command, _ []string) error {
	r, err := a.renderer(false)
	if err != nil {
		return err
	}
	raw, feed, err := a.fetchNotifications(cmd.Context())
	if err != nil {
		return err
	}
	a.inform(fmt.Sprintf("%d unread", feed.Unread))
	render := func(p *output.Printer, _ json.RawMessage) error { return printNotifications(p, feed.Items) }
	return r.object(io.NopCloser(bytes.NewReader(raw)), render)
}

// fetchNotifications reads the whole feed, which is small, to count the unread items
// before printing it.
func (a *App) fetchNotifications(ctx context.Context) (json.RawMessage, schema.NotificationFeed, error) {
	body, err := a.openFeature(ctx, featureNotifications, func(svc Services) (io.ReadCloser, error) {
		return svc.Notifications.Notifications(ctx)
	})
	if err != nil {
		return nil, schema.NotificationFeed{}, err
	}
	raw, err := readObject(body)
	if err != nil {
		return nil, schema.NotificationFeed{}, err
	}
	feed, err := decodeRecord[schema.NotificationFeed](raw)
	return raw, feed, err
}

func printNotifications(p *output.Printer, items []schema.Notification) error {
	table := p.Table(notificationColumns...)
	if err := table.Header(); err != nil {
		return err
	}
	for _, item := range items {
		if item.Kind != schema.NotificationInvite {
			continue
		}
		if err := table.Row(item.Kind, item.ID, item.Role, item.TenantName, item.ExpiresAt); err != nil {
			return err
		}
	}
	return nil
}
