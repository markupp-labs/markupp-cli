package cli

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/markupp-labs/markupp-cli/internal/client"
	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// The values the API accepts for the type and status filters of a link listing.
var (
	linkTypes    = []string{"explicit", "embed", "typed", "inferred"}
	linkStatuses = []string{"resolved", "unresolved"}
)

const (
	linkTypeHelp = "only links of this type: explicit, embed, typed or inferred"
	idWidth      = 36
	pathWidth    = 40
	typeWidth    = 16
)

// checkChoice accepts an empty value or one of choices, and names the value and the
// choices otherwise.
func checkChoice(flag, value string, choices []string) error {
	if value == "" || slices.Contains(choices, value) {
		return nil
	}
	message := fmt.Sprintf("invalid value %q, expected one of: %s", value, strings.Join(choices, ", "))
	return failure.New(failure.Usage, flag, message)
}

// startListing validates the listing flags and picks the renderer, before any request.
func (a *App) startListing(f listFlags) (client.Page, renderer, error) {
	page, err := f.page()
	if err != nil {
		return client.Page{}, renderer{}, err
	}
	r, err := a.renderer(f.ndjson)
	return page, r, err
}

// decodeLink decodes one element of a link listing.
func decodeLink(raw json.RawMessage) (schema.Link, error) {
	var link schema.Link
	if err := json.Unmarshal(raw, &link); err != nil {
		return schema.Link{}, malformedRecord(raw, err)
	}
	return link, nil
}

// linkTypeLabel is the type of a link, with the predicate of a typed link after a colon.
func linkTypeLabel(linkType string, predicate *string) string {
	if predicate == nil || *predicate == "" {
		return linkType
	}
	return linkType + ":" + *predicate
}

// summaryOrEmpty returns the note, or an empty summary (printed as "-") for nil.
func summaryOrEmpty(note *schema.NoteSummary) schema.NoteSummary {
	if note == nil {
		return schema.NoteSummary{}
	}
	return *note
}

// decimal formats a score with places decimals.
func decimal(value float64, places int) string {
	return strconv.FormatFloat(value, 'f', places, 64)
}
