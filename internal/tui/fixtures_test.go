package tui

// Payloads of the fake server, in the shapes of openapi.yaml.
const (
	vaultID = "9b2f4c1e-0a1b-4c2d-8e3f-a1b2c3d4e5f6"
	noteA   = "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	noteB   = "0e8d2c4a-1b3f-4a5e-9c7d-2f4e6a8b0c1d"
	noteC   = "5a6b7c8d-9e0f-4a1b-8c2d-3e4f5a6b7c8d"
	tenantA = "7d8e9f0a-1b2c-4d3e-8f4a-5b6c7d8e9f0a"
)

const notesBody = `[
 {"id":"` + noteA + `","title":"meeting","path":"work/meeting.md","updated_at":"2026-05-15T10:00:00Z","content_sha256":"ab","degree":3,"pagerank":0.25,"community":2},
 {"id":"` + noteB + `","title":"zebra","path":"animals/zebra.md","updated_at":"2026-05-14T09:00:00Z","content_sha256":"cd","degree":12,"pagerank":null,"community":null},
 {"id":"` + noteC + `","title":"plan","path":"work/plan.md","updated_at":"2026-05-13T08:00:00Z","content_sha256":"ef","degree":1,"pagerank":0.5,"community":2}
]`

// noteBody carries an escape sequence and a style tag, which must reach the screen
// as harmless text.
const noteBody = `{"id":"` + noteA + `","vault_id":"` + vaultID + `","path":"work/meeting.md",
 "content":"# Meeting\nagenda \u001b[31mred\u001b[0m [red]not a tag[-]\nend","content_sha256":"ab",
 "created_at":"2026-05-01T10:00:00Z","updated_at":"2026-05-15T10:00:00Z","degree":3,"pagerank":0.25,"community":2}`

const searchBody = `[
 {"note":{"id":"` + noteA + `","title":"meeting","path":"work/meeting.md"},"vault_id":"` + vaultID + `","rank":0.9,"snippet":"weekly \u0002meeting\u0003 notes"}
]`

const meBody = `{"id":"u1","email":"ana@example.com","email_verified":true}`

const tenantsBody = `[
 {"id":"` + tenantA + `","name":"Acme","role":"owner","notes_count":42,"last_note_at":null}
]`

const linksBody = `[
 {"source":{"id":"` + noteA + `","title":"meeting","path":"work/meeting.md"},"target":{"id":"` + noteC + `","title":"plan","path":"work/plan.md"},"cited_title":"plan","type":"explicit","predicate":null,"inferred_target":null},
 {"source":{"id":"` + noteA + `","title":"meeting","path":"work/meeting.md"},"target":null,"cited_title":"Budget","type":"explicit","predicate":null,"inferred_target":null}
]`

const graphBodyA = `{"nodes":[
 {"id":"` + noteA + `","title":"meeting","path":"work/meeting.md","degree":3,"pagerank":0.25,"community":2,"unresolved":false},
 {"id":"` + noteC + `","title":"plan","path":"work/plan.md","degree":1,"pagerank":0.5,"community":2,"unresolved":false},
 {"id":"` + noteB + `","title":"zebra","path":"animals/zebra.md","degree":12,"pagerank":null,"community":null,"unresolved":false},
 {"id":null,"title":"Budget","path":null,"degree":null,"pagerank":null,"community":null,"unresolved":true}
],"links":[
 {"source":0,"target":1,"type":"explicit","predicate":null},
 {"source":0,"target":3,"type":"explicit","predicate":null},
 {"source":2,"target":0,"type":"typed","predicate":"see_also"}
]}`

const graphBodyC = `{"nodes":[
 {"id":"` + noteC + `","title":"plan","path":"work/plan.md","degree":1,"pagerank":0.5,"community":2,"unresolved":false},
 {"id":"` + noteB + `","title":"roadmap","path":"work/roadmap.md","degree":1,"pagerank":null,"community":null,"unresolved":false}
],"links":[
 {"source":0,"target":1,"type":"embed","predicate":null}
]}`

const vaultsBody = `[{"id":"` + vaultID + `","name":"Personal","created_at":"2026-01-02T03:04:05Z"}]`

const membersBody = `[{"id":"u1","email":"ana@example.com","role":"owner","joined_at":"2026-01-01T00:00:00Z"}]`

const notificationsBody = `{"items":[{"kind":"invite","id":"i1","tenant_id":"t9","tenant_name":"Beta","role":"member","expires_at":"2026-06-01T00:00:00Z"}],"unread":1}`

// notesAPI is a fake server with a vault of three notes.
func notesAPI() *fakeAPI {
	return newFakeAPI(map[string]string{
		"list": notesBody, "get": noteBody, "search": searchBody, "me": meBody,
		"tenants": tenantsBody, "links": linksBody, "update": noteBody,
		"graph " + noteA: graphBodyA, "graph " + noteC: graphBodyC,
		"vaults": vaultsBody, "notifications": notificationsBody,
		"members": membersBody, "vault-members": membersBody,
		"invites": `[{"id":"i1","tenant_id":"t9","tenant_name":"Beta","role":"member","expires_at":"2026-06-01T00:00:00Z"}]`,
	})
}

// vaultConfig starts on the notes of the vault.
func vaultConfig(connector *fakeConnector) Config {
	cfg := testConfig(connector)
	cfg.Vault = vaultID
	return cfg
}
