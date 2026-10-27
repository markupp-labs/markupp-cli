package cli

import (
	"context"
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/markupp-labs/markupp-cli/internal/output"
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

const notesMentionsLong = `List the notes whose title this note cites in its text without linking to
them, one per line, highest prior first. PRIOR is the probability that the cited
title points to that note. With --back, list the notes that cite a title of this
one without linking to it.

Mentions are link suggestions computed on read, over what the indexer has
processed.`

const notesSuggestLong = `List the notes this note could link to, best first, leaving out the notes
already linked in either direction.

SCORE fuses two rankings by Reciprocal Rank Fusion: STRUCT is the resource
allocation over the two hop neighbourhood (0 outside the structural ranking) and
COSINE the tf-idf similarity of the texts (0 outside the lexical ranking).`

// noteMentionsFlags are the flags of mu notes mentions.
type noteMentionsFlags struct {
	list listFlags
	back bool
}

var mentionColumns = []output.Column{
	{Name: "ID", Width: idWidth}, {Name: "PRIOR", Width: 6}, {Name: "PATH", Width: pathWidth}, {Name: "CITED"},
}

var suggestionColumns = []output.Column{
	{Name: "ID", Width: idWidth}, {Name: "SCORE", Width: 6}, {Name: "STRUCT", Width: 6},
	{Name: "COSINE", Width: 6}, {Name: "PATH"},
}

func newNotesMentionsCmd(app *App) *cobra.Command {
	var flags noteMentionsFlags
	cmd := &cobra.Command{
		Use:   "mentions ID",
		Short: "List the titles a note cites without linking, or with --back who cites it",
		Long:  notesMentionsLong,
		Example: `  mu notes mentions 3f2a9c1e-5b7d-4e8a-9c0f-1a2b3c4d5e6f --limit 5
  mu notes mentions 3f2a9c1e-5b7d-4e8a-9c0f-1a2b3c4d5e6f --back --ndjson`,
		Args: exactArgs("ID"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.listNoteMentions(cmd.Context(), args[0], flags)
		},
	}
	addListFlags(cmd, &flags.list)
	cmd.Flags().BoolVar(&flags.back, "back", false, "list the notes that cite a title of this one without linking")
	return cmd
}

func (a *App) listNoteMentions(ctx context.Context, id string, flags noteMentionsFlags) error {
	page, r, err := a.startListing(flags.list)
	if err != nil {
		return err
	}
	svc, err := a.connect(ctx, Target{})
	if err != nil {
		return err
	}
	fetch := svc.NoteLinks.NoteMentions
	if flags.back {
		fetch = svc.NoteLinks.NoteBackmentions
	}
	body, err := fetch(ctx, id, page)
	if err != nil {
		return err
	}
	return r.list(body, mentionColumns, mentionRow)
}

func mentionRow(raw json.RawMessage) ([]string, error) {
	mention, err := decodeRecord[schema.Mention](raw)
	if err != nil {
		return nil, err
	}
	return []string{mention.Note.ID, decimal(mention.Prior, 2), mention.Note.Path, mention.CitedTitle}, nil
}

func newNotesSuggestCmd(app *App) *cobra.Command {
	var flags listFlags
	cmd := &cobra.Command{
		Use:   "suggest ID",
		Short: "List the notes a note could link to, best first",
		Long:  notesSuggestLong,
		Example: `  mu notes suggest 3f2a9c1e-5b7d-4e8a-9c0f-1a2b3c4d5e6f --limit 5
  mu notes suggest 3f2a9c1e-5b7d-4e8a-9c0f-1a2b3c4d5e6f --json | jq -r '.[].note.path'`,
		Args: exactArgs("ID"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.listNoteSuggestions(cmd.Context(), args[0], flags)
		},
	}
	addListFlags(cmd, &flags)
	return cmd
}

func (a *App) listNoteSuggestions(ctx context.Context, id string, flags listFlags) error {
	page, r, err := a.startListing(flags)
	if err != nil {
		return err
	}
	svc, err := a.connect(ctx, Target{})
	if err != nil {
		return err
	}
	body, err := svc.NoteLinks.NoteSuggestions(ctx, id, page)
	if err != nil {
		return err
	}
	return r.list(body, suggestionColumns, suggestionRow)
}

func suggestionRow(raw json.RawMessage) ([]string, error) {
	suggestion, err := decodeRecord[schema.LinkSuggestion](raw)
	if err != nil {
		return nil, err
	}
	return []string{
		suggestion.Note.ID, decimal(suggestion.Score, 4), decimal(suggestion.ResourceAllocation, 2),
		decimal(suggestion.Cosine, 2), suggestion.Note.Path,
	}, nil
}
