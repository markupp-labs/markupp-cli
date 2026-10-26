package tui

import (
	"context"
	"io"
	"strconv"

	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// openCommunities is :communities [vault].
func openCommunities(s *shell, vault string) error {
	vault, err := s.graphVault(":communities", vault)
	if err != nil {
		return err
	}
	fetch := func(ctx context.Context) (io.ReadCloser, error) {
		return s.ports.Graph.VaultCommunities(ctx, vault, listPage)
	}
	s.reset(newListView(s, listing{
		resource: "communities", context: vault,
		columns: []string{"COMMUNITY", "SIZE", "HUB", "ID"},
		load:    arrayOf(fetch, rowsOf(communityRow)),
		enter:   func(r record) { s.viewNote(r.note) },
		keys:    s.noteKeys,
	}))
	return nil
}

// communityRow is a community; its note is the hub, which Enter opens.
func communityRow(community schema.Community) record {
	cells := []string{strconv.FormatInt(community.ID, 10), strconv.Itoa(community.Size), community.Hub.Path, community.Hub.ID}
	return record{id: community.Hub.ID, note: community.Hub.ID, label: community.Hub.Path, cells: cells}
}

// openVaultLinks is :links [vault], every link of the vault.
func openVaultLinks(s *shell, vault string) error {
	vault, err := s.graphVault(":links", vault)
	if err != nil {
		return err
	}
	fetch := func(ctx context.Context) (io.ReadCloser, error) {
		return s.ports.Graph.VaultLinks(ctx, vault, listPage)
	}
	s.reset(newListView(s, listing{
		resource: "links", context: vault,
		columns: []string{"TYPE", "SOURCE", "TARGET", "ID"},
		load:    arrayOf(fetch, rowsOf(vaultLinkRow)),
		enter:   func(r record) { s.viewNote(r.note) },
		keys:    s.noteKeys,
	}))
	return nil
}

// vaultLinkRow is a link of the vault; its note is the source, which Enter opens.
func vaultLinkRow(link schema.Link) record {
	cells := []string{link.Type, link.Source.Path, targetLabel(link), link.Source.ID}
	return record{id: link.Source.ID, note: link.Source.ID, label: link.Source.Path, cells: cells}
}

// graphVault is the vault of a vault wide graph command: the argument or the current.
func (s *shell) graphVault(command, vault string) (string, error) {
	if err := s.need(s.ports.Graph != nil, command[1:]); err != nil {
		return "", err
	}
	if vault == "" {
		vault = s.vault
	}
	if vault == "" {
		return "", failure.New(failure.Usage, command, "no vault selected, expected "+command+" <vault id>")
	}
	s.vault = vault
	return vault, nil
}
