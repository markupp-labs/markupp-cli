package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/output"
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// VaultPublisher reads, sets and removes the publication of a vault.
type VaultPublisher interface {
	GetPublication(ctx context.Context, id string) (io.ReadCloser, error)
	PublishVault(ctx context.Context, id, slug, title string) (io.ReadCloser, error)
	UnpublishVault(ctx context.Context, id string) error
}

const publishLong = `Publish the vault ID for reading without an account, under --slug, or change the
slug and title of a vault already published. The slug is unique across every
tenant: 3 to 63 characters among a-z, 0-9 and "-". An empty --title shows the vault
name. Only owners and admins of the tenant publish. Anyone can then read the vault
with --public SLUG on the notes and graph commands.`

func newVaultsPublicationCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:     "publication ID",
		Short:   "Show how a vault is published",
		Long:    requiresServer(featurePublication, "Show the slug, title and first publisher of the vault ID. A vault that is not published exits 5."),
		Example: "  mu vaults publication 3b241101-e2bb-4255-8caf-4136c566a962\n  mu vaults publication 3b241101-e2bb-4255-8caf-4136c566a962 --json | jq -r .slug",
		Args:    exactArgs("ID"),
		RunE:    app.showPublication,
	}
}

func newVaultsPublishCmd(app *App) *cobra.Command {
	var request schema.PublicationRequest
	cmd := &cobra.Command{
		Use:     "publish ID --slug SLUG",
		Short:   "Publish a vault for reading without an account",
		Long:    requiresServer(featurePublication, publishLong),
		Example: "  mu vaults publish 3b241101-e2bb-4255-8caf-4136c566a962 --slug docs --title \"Markupp docs\"",
		Args:    exactArgs("ID"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.publishVault(cmd.Context(), args[0], request)
		},
	}
	cmd.Flags().StringVar(&request.Slug, "slug", "", "slug of the public URL, required (3 to 63 of a-z, 0-9 and -)")
	cmd.Flags().StringVar(&request.Title, "title", "", "title of the public vault (default: the vault name)")
	return cmd
}

func newVaultsUnpublishCmd(app *App) *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:     "unpublish ID",
		Short:   "Stop the public reading of a vault",
		Long:    requiresServer(featurePublication, "Take the vault ID out of the public reads. Its notes stay; the slug becomes free."),
		Example: "  mu vaults unpublish 3b241101-e2bb-4255-8caf-4136c566a962 --yes",
		Args:    exactArgs("ID"),
		RunE:    func(cmd *cobra.Command, args []string) error { return app.unpublishVault(cmd.Context(), args[0], yes) },
	}
	addYesFlag(cmd, &yes, "the public URL stops working and anyone can take the slug")
	return cmd
}

func printPublication(p *output.Printer, raw json.RawMessage) error {
	publication, err := decodeRecord[schema.Publication](raw)
	if err != nil {
		return err
	}
	return p.Fields(
		output.Field{Key: "vault", Value: publication.VaultID}, output.Field{Key: "slug", Value: publication.Slug},
		output.Field{Key: "title", Value: publication.Title}, output.Field{Key: "published_by", Value: publication.PublishedBy},
		output.Field{Key: "published_at", Value: publication.PublishedAt},
	)
}

func (a *App) showPublication(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	return a.showFeature(ctx, featurePublication, printPublication, func(svc Services) (io.ReadCloser, error) {
		return svc.Publication.GetPublication(ctx, args[0])
	})
}

func (a *App) publishVault(ctx context.Context, id string, request schema.PublicationRequest) error {
	if request.Slug == "" {
		return failure.New(failure.Usage, "--slug", fmt.Sprintf("missing for vault %q, expected 3 to 63 characters among a-z, 0-9 and -", id))
	}
	return a.showFeature(ctx, featurePublication, printPublication, func(svc Services) (io.ReadCloser, error) {
		return svc.Publication.PublishVault(ctx, id, request.Slug, request.Title)
	})
}

func (a *App) unpublishVault(ctx context.Context, id string, yes bool) error {
	if err := requireYes(yes, "unpublish vault", id); err != nil {
		return err
	}
	unpublish := func(svc Services) error { return svc.Publication.UnpublishVault(ctx, id) }
	if err := a.execFeature(ctx, featurePublication, unpublish); err != nil {
		return err
	}
	a.confirmDone(fmt.Sprintf("unpublished vault %q", id))
	return nil
}
