package config

// Environment variables read by mu.
const (
	EnvURL    = "MARKUPP_URL"
	EnvToken  = "MARKUPP_TOKEN"
	EnvTenant = "MARKUPP_TENANT"
	EnvVault  = "MARKUPP_VAULT"
)

// Source says where a resolved value came from.
type Source string

// The sources, from the strongest to the weakest. FromNone means unset.
const (
	FromNone Source = ""
	FromFlag Source = "flag"
	FromEnv  Source = "env"
	FromFile Source = "file"
)

// Value is a resolved setting and its source.
type Value struct {
	Value  string
	Source Source
}

// Overrides are the values given by flags; empty means not given.
type Overrides struct {
	URL    string
	Tenant string
	Vault  string
}

// Resolved holds every setting after precedence. Token only ever comes from
// MARKUPP_TOKEN, a raw access token that is never refreshed; the stored login is
// Settings.Auth.
type Resolved struct {
	URL    Value
	Tenant Value
	Vault  Value
	Token  Value
}

// Resolve applies the precedence flag > environment > file to every setting.
func Resolve(flags Overrides, lookup LookupEnv, file Settings) Resolved {
	return Resolved{
		URL:    pick(flags.URL, envValue(lookup, EnvURL), file.URL),
		Tenant: pick(flags.Tenant, envValue(lookup, EnvTenant), file.Tenant),
		Vault:  pick(flags.Vault, envValue(lookup, EnvVault), file.Vault),
		Token:  pick("", envValue(lookup, EnvToken), ""),
	}
}

func pick(flag, env, file string) Value {
	switch {
	case flag != "":
		return Value{Value: flag, Source: FromFlag}
	case env != "":
		return Value{Value: env, Source: FromEnv}
	case file != "":
		return Value{Value: file, Source: FromFile}
	}
	return Value{}
}
