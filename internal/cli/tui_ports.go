package cli

import (
	"context"
	"io"

	"github.com/markupp-labs/markupp-cli/internal/client"
	"github.com/markupp-labs/markupp-cli/internal/config"
	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/schema"
	"github.com/markupp-labs/markupp-cli/internal/tui"
)

// tuiPorts hands the services to the interface. A port the services lack stays nil,
// and the interface says the view is not available.
func tuiPorts(svc Services) tui.Ports {
	ports := tui.Ports{Notes: svc.Notes, Relations: svc.NoteLinks, Account: svc.Account}
	if svc.Graph != nil {
		ports.Graph = tuiGraph{graph: svc.Graph}
	}
	if svc.Vaults != nil && svc.VaultMembers != nil {
		ports.Vaults = tuiVaults{vaults: svc.Vaults, members: svc.VaultMembers}
	}
	if svc.Members != nil && svc.Invites != nil {
		ports.Team = tuiTeam{members: svc.Members, invites: svc.Invites}
	}
	if svc.Notifications != nil {
		ports.Notifications = tuiNotifications{feed: svc.Notifications}
	}
	return ports
}

// tuiVaults lists the vaults and their members for the interface, naming the server
// feature when the server has no vault routes.
type tuiVaults struct {
	vaults  VaultAdmin
	members VaultMembership
}

func (v tuiVaults) ListVaults(ctx context.Context) (io.ReadCloser, error) {
	return featureRead(featureVaults)(v.vaults.ListVaults(ctx))
}

func (v tuiVaults) ListVaultMembers(ctx context.Context, id string) (io.ReadCloser, error) {
	return featureRead(featureVaults)(v.members.ListVaultMembers(ctx, id))
}

// tuiTeam lists the members of the tenant and the invites addressed to the account.
type tuiTeam struct {
	members MemberAdmin
	invites InviteInbox
}

func (t tuiTeam) ListMembers(ctx context.Context) (io.ReadCloser, error) {
	return t.members.ListMembers(ctx)
}

func (t tuiTeam) ListMyInvites(ctx context.Context) (io.ReadCloser, error) {
	return t.invites.ListMyInvites(ctx)
}

// tuiNotifications reads the feed, naming the server feature when it has none.
type tuiNotifications struct {
	feed NotificationReader
}

func (n tuiNotifications) Notifications(ctx context.Context) (io.ReadCloser, error) {
	return featureRead(featureNotifications)(n.feed.Notifications(ctx))
}

// tuiPublic connects to the public vault slug without credentials. Only its notes
// are reachable, read only.
func (a *App) tuiPublic(ctx context.Context, resolved config.Resolved, settings config.Settings, slug string) (tui.Ports, error) {
	svc, err := a.dial(ctx, resolved, settings, true)
	if err != nil {
		return tui.Ports{}, err
	}
	public, err := publicOf(svc)
	if err != nil {
		return tui.Ports{}, err
	}
	return tui.Ports{Notes: tuiPublicNotes{slug: slug, public: public}}, nil
}

// tuiPublicNotes is the notes port of a public vault: the reads go to the public
// routes of slug, whatever vault the interface names, and the writes are refused.
type tuiPublicNotes struct {
	slug   string
	public PublicReader
}

func (p tuiPublicNotes) ListNotes(ctx context.Context, _ string, filter client.NoteFilter, page client.Page) (io.ReadCloser, error) {
	return publicRead(p.public.PublicNotes(ctx, p.slug, filter, page))
}

func (p tuiPublicNotes) GetNote(ctx context.Context, id string) (io.ReadCloser, error) {
	return publicRead(p.public.PublicNote(ctx, p.slug, id))
}

func (p tuiPublicNotes) SearchNotes(ctx context.Context, text, _ string, page client.Page) (io.ReadCloser, error) {
	return publicRead(p.public.PublicSearch(ctx, p.slug, text, page))
}

func (p tuiPublicNotes) UpdateNote(context.Context, string, schema.NoteUpdateRequest) (io.ReadCloser, error) {
	return nil, p.readOnly()
}

func (p tuiPublicNotes) DeleteNote(context.Context, string) error {
	return p.readOnly()
}

func (p tuiPublicNotes) readOnly() error {
	return failure.New(failure.Forbidden, ":public "+p.slug, "a public vault is read only, expected a vault of the account")
}
