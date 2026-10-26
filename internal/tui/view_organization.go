package tui

import (
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// openCommunities is :communities [vault].
func openCommunities(s *shell, vault string) error { return *new(error) }

// communityRow is a community; its note is the hub, which Enter opens.
func communityRow(community schema.Community) record { return *new(record) }

// openVaultLinks is :links [vault], every link of the vault.
func openVaultLinks(s *shell, vault string) error { return *new(error) }

// vaultLinkRow is a link of the vault; its note is the source, which Enter opens.
func vaultLinkRow(link schema.Link) record { return *new(record) }

// graphVault is the vault of a vault wide graph command: the argument or the current.
func (s *shell) graphVault(command, vault string) (string, error) { return *new(string), *new(error) }
