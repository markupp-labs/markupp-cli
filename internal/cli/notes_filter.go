package cli

import (
	"github.com/spf13/cobra"

	"github.com/markupp-labs/markupp-cli/internal/client"
)

// The values the API accepts for the order of a note listing.
var (
	noteSorts      = []string{"pagerank", "degree", "updated"}
	sortDirections = []string{"asc", "desc"}
)

// noteFilterFlags are the filters and the order of mu notes list.
type noteFilterFlags struct {
	orphan    bool
	community int64
	sort      string
	direction string
}

func addNoteFilterFlags(cmd *cobra.Command, f *noteFilterFlags) {
	flags := cmd.Flags()
	flags.BoolVar(&f.orphan, "orphan", false, "only the notes without links (--orphan=false: only the linked ones)")
	flags.Int64Var(&f.community, "community", 0, "only the notes of the community with this id")
	flags.StringVar(&f.sort, "sort", "", "order by pagerank, degree or updated (default: by path)")
	flags.StringVar(&f.direction, "direction", "", "order direction, asc or desc (server default desc)")
}

// noteFilter validates the flags and sends only the ones given on the command line,
// so --orphan=false still filters and an absent --community does not.
func (f *noteFilterFlags) noteFilter(cmd *cobra.Command) (client.NoteFilter, error) {
	if err := checkChoice("--sort", f.sort, noteSorts); err != nil {
		return client.NoteFilter{}, err
	}
	if err := checkChoice("--direction", f.direction, sortDirections); err != nil {
		return client.NoteFilter{}, err
	}
	filter := client.NoteFilter{Sort: f.sort, Direction: f.direction}
	if cmd.Flags().Changed("orphan") {
		filter.Orphan = &f.orphan
	}
	if cmd.Flags().Changed("community") {
		filter.Community = &f.community
	}
	return filter, nil
}
