package tui

// vaultRecord is a vault of the tenant (VaultResponse).
type vaultRecord struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

// openVaults is :vaults.
func openVaults(s *shell, _ string) error { return *new(error) }

// newVaultsView lists the vaults of the tenant; Enter opens the notes of one.
func newVaultsView(s *shell) *listView { return *new(*listView) }

func vaultRow(vault vaultRecord) record { return *new(record) }

// enterVault makes the vault current and opens its notes.
func (s *shell) enterVault(r record) {}

// vaultKeys are m for the members of the vault and y to copy its id.
func (s *shell) vaultKeys(selected func() (record, bool)) []Hotkey { return *new([]Hotkey) }
