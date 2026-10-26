package cli

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/markupp-labs/markupp-cli/internal/config"
	"github.com/markupp-labs/markupp-cli/internal/output"
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

func printStatusText(p *output.Printer, view statusView) error {
	fields, err := statusFields(view)
	if err != nil {
		return err
	}
	if err := p.Fields(fields...); err != nil {
		return err
	}
	if view.Tenants == nil {
		return nil
	}
	if err := p.Line(""); err != nil {
		return err
	}
	return printTenants(p, view.Tenants)
}

func statusFields(view statusView) ([]output.Field, error) {
	fields := []output.Field{{Key: "server", Value: view.Server}}
	if view.Health != nil {
		var health schema.Health
		if err := json.Unmarshal(view.Health, &health); err != nil {
			return nil, malformedRecord(view.Health, err)
		}
		fields = append(fields, output.Field{Key: "health", Value: health.Status})
	}
	fields = append(fields, output.Field{Key: "token", Value: describeCredential(view.Token)})
	if view.Account == nil {
		return fields, nil
	}
	var user schema.User
	if err := json.Unmarshal(view.Account, &user); err != nil {
		return nil, malformedRecord(view.Account, err)
	}
	return append(fields, output.Field{Key: "account", Value: user.Email}, output.Field{Key: "user id", Value: user.ID}), nil
}

func describeCredential(cred credentialView) string {
	switch cred.Source {
	case credentialFromEnv:
		return config.EnvToken + " from the environment, never refreshed"
	case credentialFromLogin:
		return "login to " + cred.Issuer + ", expires " + cred.ExpiresAt.Format(time.RFC3339)
	}
	return "none; run mu login"
}

func printTenants(p *output.Printer, raw json.RawMessage) error {
	var tenants []schema.MyTenant
	if err := json.Unmarshal(raw, &tenants); err != nil {
		return malformedRecord(raw, err)
	}
	table := p.Table(output.Column{Name: "TENANT", Width: 36}, output.Column{Name: "NAME", Width: 24},
		output.Column{Name: "ROLE", Width: 6}, output.Column{Name: "NOTES"})
	if err := table.Header(); err != nil {
		return err
	}
	for _, tenant := range tenants {
		if err := table.Row(tenant.ID, tenant.Name, tenant.Role, strconv.Itoa(tenant.NotesCount)); err != nil {
			return err
		}
	}
	return nil
}
