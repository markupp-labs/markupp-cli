package tui

// vaultRecord is a vault of the tenant (VaultResponse).
type vaultRecord struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

// openVaults is :vaults.
func openVaults(s *shell, _ string) error {
	if err := s.need(s.ports.Vaults != nil, "vaults"); err != nil {
		return err
	}
	s.reset(newVaultsView(s))
	return nil
}

// newVaultsView lists the vaults of the tenant; Enter opens the notes of one.
func newVaultsView(s *shell) *listView {
	return newListView(s, listing{
		resource: "vaults",
		columns:  []string{"NAME", "CREATED", "ID"},
		load:     arrayOf(s.ports.Vaults.ListVaults, rowsOf(vaultRow)),
		enter:    s.enterVault,
		keys:     s.vaultKeys,
	})
}

func vaultRow(vault vaultRecord) record {
	return record{id: vault.ID, label: vault.Name, cells: []string{vault.Name, vault.CreatedAt, vault.ID}}
}

// enterVault makes the vault current and opens its notes.
func (s *shell) enterVault(r record) {
	if err := s.need(s.ports.Notes != nil, "notes"); err != nil {
		s.flashError(err)
		return
	}
	s.vault = r.id
	s.push(newNotesView(s, r.id))
}

// vaultKeys are m for the members of the vault and y to copy its id.
func (s *shell) vaultKeys(selected func() (record, bool)) []Hotkey {
	members := func() {
		if row, ok := selected(); ok {
			s.push(newMembersView(s, "vault members", row.label, fetchOf(s.ports.Vaults.ListVaultMembers, row.id)))
		}
	}
	return append([]Hotkey{runeKey('m', "members", members)}, s.copyKeys(selected)...)
}
