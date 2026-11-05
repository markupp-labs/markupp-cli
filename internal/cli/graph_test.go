package cli

import (
	"strings"
	"testing"

	"github.com/markupp-labs/markupp-cli/internal/client"
	"github.com/markupp-labs/markupp-cli/internal/output/outputtest"
)

const graphBody = `{"nodes":[
{"id":"n1","title":"a","path":"a.md","degree":2,"pagerank":0.41234,"community":7,"unresolved":false},
{"id":"n2","title":"b","path":"notes/b.md","degree":null,"pagerank":null,"community":null,"unresolved":false},
{"id":null,"title":"Meting","path":null,"degree":null,"pagerank":null,"community":null,"unresolved":true}],
"links":[{"source":0,"target":1,"type":"typed","predicate":"author"},{"source":0,"target":2,"type":"explicit","predicate":null}]}`

const communitiesBody = `[{"id":7,"size":42,"hub":{"id":"n1","title":"a","path":"a.md"}},{"id":12,"size":3,"hub":{"id":"n9","title":"z","path":"z/z.md"}}]`

const communityBody = `{"id":7,"size":42,"hub":{"id":"n1","title":"a","path":"a.md"}}`

func TestGraphShowVaultText(t *testing.T) {
	mu, graph := graphMu(graphBody)
	run := mu.run(t, "graph", "show", "--vault", "v1")
	run.expectCode(t, 0)
	outputtest.Golden(t, "graph_show", []byte(run.stdout))
	expectCalls(t, graph.calls, portCall{method: "VaultGraph", subject: "v1"})
}

func TestGraphShowUsesTheConfiguredVault(t *testing.T) {
	mu, graph := graphMu(graphBody)
	mu.env["MARKUPP_VAULT"] = "v-env"
	mu.run(t, "graph", "show", "--no-orphans").expectCode(t, 0)
	expectCalls(t, graph.calls, portCall{method: "VaultGraph", subject: "v-env", graph: client.GraphFilter{HideOrphans: true}})
}

func TestGraphShowNoteSendsDepthAndFilters(t *testing.T) {
	mu, graph := graphMu(graphBody)
	mu.env["MARKUPP_VAULT"] = "v-env"
	run := mu.run(t, "graph", "show", "--note", "n1", "--depth", "3",
		"--no-orphans", "--no-unresolved", "--no-inferred", "--community", "0")
	run.expectCode(t, 0)
	zero := int64(0)
	filter := client.GraphFilter{HideOrphans: true, HideUnresolved: true, HideInferred: true, Community: &zero}
	expectCalls(t, graph.calls, portCall{method: "NoteGraph", subject: "n1", depth: 3, graph: filter})
	if conn := mu.server.connections[0]; conn.Public {
		t.Error("graph show --note connected as public")
	}
}

func TestGraphShowPublicVault(t *testing.T) {
	mu, graph := graphMu(graphBody)
	mu.run(t, "graph", "show", "--public", "my-garden", "--no-inferred").expectCode(t, 0)
	expectCalls(t, graph.calls, portCall{method: "PublicVaultGraph", subject: "my-garden", graph: client.GraphFilter{HideInferred: true}})
	if conn := mu.server.connections[0]; !conn.Public {
		t.Error("graph show --public did not connect as public")
	}
}

func TestGraphShowJSONIsTheBodyUnchanged(t *testing.T) {
	mu, _ := graphMu(graphBody)
	run := mu.run(t, "graph", "show", "--vault", "v1", "--json")
	run.expectCode(t, 0)
	if run.stdout != graphBody+"\n" {
		t.Errorf("--json printed %q, want the body unchanged", run.stdout)
	}
}

func TestGraphShowRefusesBadScopes(t *testing.T) {
	cases := map[string][]string{
		"no vault":           {"graph", "show"},
		"depth without note": {"graph", "show", "--vault", "v1", "--depth", "2"},
		"depth too deep":     {"graph", "show", "--note", "n1", "--depth", "4"},
		"depth zero":         {"graph", "show", "--note", "n1", "--depth", "0"},
		"note and vault":     {"graph", "show", "--note", "n1", "--vault", "v1"},
		"note and public":    {"graph", "show", "--note", "n1", "--public", "blog"},
		"vault and public":   {"graph", "show", "--vault", "v1", "--public", "blog"},
		"bad community":      {"graph", "show", "--vault", "v1", "--community", "seven"},
		"positional arg":     {"graph", "show", "v1"},
	}
	for name, args := range cases {
		mu, graph := graphMu(graphBody)
		run := mu.run(t, args...)
		if run.code != 2 || len(graph.calls) != 0 {
			t.Errorf("%s: exit code %d after %d calls, want 2 before any call\nstderr: %s", name, run.code, len(graph.calls), run.stderr)
		}
	}
}

func TestGraphShowNamesTheDepth(t *testing.T) {
	mu, _ := graphMu(graphBody)
	run := mu.run(t, "graph", "show", "--note", "n1", "--depth", "4")
	if !strings.Contains(run.stderr, "4") || !strings.Contains(run.stderr, "1 to 3") {
		t.Errorf("stderr %q should name the depth and the range", run.stderr)
	}
}

func TestGraphShowRefusesALinkOutsideTheNodes(t *testing.T) {
	mu, _ := graphMu(`{"nodes":[{"id":"n1","path":"a.md"}],"links":[{"source":0,"target":5,"type":"explicit"}]}`)
	run := mu.run(t, "graph", "show", "--vault", "v1")
	run.expectCode(t, 1)
	if !strings.Contains(run.stderr, "0 -> 5") {
		t.Errorf("stderr %q should name the link", run.stderr)
	}
}

func TestGraphCommunitiesText(t *testing.T) {
	mu, graph := graphMu(communitiesBody)
	run := mu.run(t, "graph", "communities", "--vault", "v1", "--limit", "2")
	run.expectCode(t, 0)
	outputtest.Golden(t, "graph_communities", []byte(run.stdout))
	expectCalls(t, graph.calls, portCall{method: "VaultCommunities", subject: "v1", page: client.Page{Limit: 2}})
}

func TestGraphCommunityText(t *testing.T) {
	mu, graph := graphMu(communityBody)
	mu.store.settings.Vault = "v-file"
	run := mu.run(t, "graph", "communities", "7")
	run.expectCode(t, 0)
	outputtest.Golden(t, "graph_community", []byte(run.stdout))
	expectCalls(t, graph.calls, portCall{method: "VaultCommunity", subject: "v-file", cid: 7})
}

func TestGraphCommunitiesRefusesBadArguments(t *testing.T) {
	cases := map[string][]string{
		"not an integer": {"graph", "communities", "--vault", "v1", "seven"},
		"limit with cid": {"graph", "communities", "--vault", "v1", "7", "--limit", "3"},
		"two cids":       {"graph", "communities", "--vault", "v1", "7", "8"},
		"no vault":       {"graph", "communities"},
	}
	for name, args := range cases {
		mu, graph := graphMu(communityBody)
		run := mu.run(t, args...)
		if run.code != 2 || len(graph.calls) != 0 {
			t.Errorf("%s: exit code %d after %d calls, want 2 before any call\nstderr: %s", name, run.code, len(graph.calls), run.stderr)
		}
	}
}

func TestGraphLinksText(t *testing.T) {
	mu, graph := graphMu(linksBody)
	run := mu.run(t, "graph", "links", "--vault", "v1", "--status", "unresolved", "--type", "explicit", "--offset", "20")
	run.expectCode(t, 0)
	outputtest.Golden(t, "graph_links", []byte(run.stdout))
	filter := client.LinkFilter{Status: "unresolved", Type: "explicit"}
	expectCalls(t, graph.calls, portCall{method: "VaultLinks", subject: "v1", links: filter, page: client.Page{Offset: 20}})
}

func TestGraphLinksPublicVault(t *testing.T) {
	mu, graph := graphMu(linksBody)
	mu.run(t, "graph", "links", "--public", "my-garden", "--ndjson").expectCode(t, 0)
	expectCalls(t, graph.calls, portCall{method: "PublicVaultLinks", subject: "my-garden"})
	if !mu.server.connections[0].Public {
		t.Error("graph links --public did not connect as public")
	}
}

func TestGraphLinksRefusesAnUnknownStatus(t *testing.T) {
	mu, graph := graphMu(linksBody)
	run := mu.run(t, "graph", "links", "--vault", "v1", "--status", "broken")
	run.expectCode(t, 2)
	if !strings.Contains(run.stderr, `"broken"`) || len(graph.calls) != 0 {
		t.Errorf("stderr %q should name the status, and no call should be made", run.stderr)
	}
}
