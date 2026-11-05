package schema

import "testing"

func TestGraphDecodesNodesAndLinks(t *testing.T) {
	got := decode[Graph](t, `{"nodes":[`+
		`{"id":"n1","title":"a","path":"a.md","degree":2,"pagerank":0.5,"community":7,"unresolved":false},`+
		`{"id":null,"title":"ghost","path":null,"degree":null,"pagerank":null,"community":null,"unresolved":true}],`+
		`"links":[{"source":0,"target":1,"type":"typed","predicate":"author"}]}`)
	if len(got.Nodes) != 2 || len(got.Links) != 1 {
		t.Fatalf("graph = %+v, want 2 nodes and 1 link", got)
	}
	note, ghost := got.Nodes[0], got.Nodes[1]
	if note.ID == nil || *note.ID != "n1" || note.Path == nil || note.Degree == nil || *note.Community != 7 {
		t.Errorf("note node = %+v", note)
	}
	if ghost.ID != nil || ghost.Path != nil || !ghost.Unresolved || ghost.Title != "ghost" {
		t.Errorf("unresolved node = %+v", ghost)
	}
	link := got.Links[0]
	if link.Source != 0 || link.Target != 1 || link.Type != "typed" || link.Predicate == nil || *link.Predicate != "author" {
		t.Errorf("link = %+v", link)
	}
}

func TestCommunityDecodesHub(t *testing.T) {
	got := decode[Community](t, `{"id":7,"size":42,"hub":{"id":"n1","title":"a","path":"a.md"}}`)
	want := Community{ID: 7, Size: 42, Hub: NoteSummary{ID: "n1", Title: "a", Path: "a.md"}}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}
