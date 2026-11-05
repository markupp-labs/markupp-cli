package schema

// Publication is how a vault is published for reading without an account.
type Publication struct {
	VaultID     string `json:"vault_id"`
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	PublishedBy string `json:"published_by"`
	PublishedAt string `json:"published_at"`
}

// PublicationRequest publishes a vault under Slug. An empty Title is left out, and
// the server uses the vault name.
type PublicationRequest struct {
	Slug  string `json:"slug"`
	Title string `json:"title,omitempty"`
}
