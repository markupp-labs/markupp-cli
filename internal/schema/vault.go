package schema

// Vault is a vault of the tenant, as GET /vaults lists it.
type Vault struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

// VaultNameRequest is the body that creates or renames a vault.
type VaultNameRequest struct {
	Name string `json:"name"`
}

// VaultMemberRequest is the body that adds a tenant account to a vault.
type VaultMemberRequest struct {
	UserID string `json:"user_id"`
}
