package cli

import (
	"encoding/json"
	"fmt"

	"github.com/markupp-labs/markupp-cli/internal/output"
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// renderGraph prints the nodes of a graph, a blank line, then its links by node label.
func renderGraph(p *output.Printer, raw json.RawMessage) error {
	graph, err := decodeRecord[schema.Graph](raw)
	if err != nil {
		return err
	}
	if err := printGraphNodes(p, graph.Nodes); err != nil {
		return err
	}
	if err := p.Line(""); err != nil {
		return err
	}
	return printGraphLinks(p, graph, raw)
}

func printGraphNodes(p *output.Printer, nodes []schema.GraphNode) error {
	table := p.Table(graphNodeColumns...)
	if err := table.Header(); err != nil {
		return err
	}
	for _, node := range nodes {
		row := []string{
			optionalText(node.ID), dashedInt(node.Degree), optionalDecimal(node.PageRank, 4),
			dashedInt64(node.Community), nodeLabel(node),
		}
		if err := table.Row(row...); err != nil {
			return err
		}
	}
	return nil
}

func printGraphLinks(p *output.Printer, graph schema.Graph, raw json.RawMessage) error {
	table := p.Table(graphLinkColumns...)
	if err := table.Header(); err != nil {
		return err
	}
	for _, link := range graph.Links {
		source, target, err := linkEnds(graph.Nodes, link, raw)
		if err != nil {
			return err
		}
		if err := table.Row(source, linkTypeLabel(link.Type, link.Predicate), target); err != nil {
			return err
		}
	}
	return nil
}

// linkEnds returns the labels of the nodes a link joins, refusing an index outside
// the nodes.
func linkEnds(nodes []schema.GraphNode, link schema.GraphLink, raw json.RawMessage) (string, string, error) {
	for _, index := range []int{link.Source, link.Target} {
		if index < 0 || index >= len(nodes) {
			err := fmt.Errorf("link %d -> %d refers to a node outside the %d nodes", link.Source, link.Target, len(nodes))
			return "", "", malformedRecord(raw, err)
		}
	}
	return nodeLabel(nodes[link.Source]), nodeLabel(nodes[link.Target]), nil
}

// nodeLabel is the path of a note, or the cited title of a node without a note.
func nodeLabel(node schema.GraphNode) string {
	if node.Path != nil && *node.Path != "" {
		return *node.Path
	}
	if node.Unresolved {
		return node.Title + " (unresolved)"
	}
	return node.Title
}

func optionalText(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func optionalDecimal(value *float64, places int) string {
	if value == nil {
		return ""
	}
	return decimal(*value, places)
}
