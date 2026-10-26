package config

import "testing"

func TestResolvePrecedenceFlagEnvFile(t *testing.T) {
	file := Settings{URL: "https://file", Tenant: "file-t", Vault: "file-v"}
	env := fakeEnv{EnvURL: "https://env", EnvTenant: "env-t"}
	got := Resolve(Overrides{URL: "https://flag"}, env.lookup, file)
	want := Resolved{
		URL:    Value{Value: "https://flag", Source: FromFlag},
		Tenant: Value{Value: "env-t", Source: FromEnv},
		Vault:  Value{Value: "file-v", Source: FromFile},
	}
	if got != want {
		t.Errorf("Resolve = %+v, want %+v", got, want)
	}
}

func TestResolveTokenComesOnlyFromEnvironment(t *testing.T) {
	file := Settings{Auth: &Credential{AccessToken: "stored"}}
	got := Resolve(Overrides{}, fakeEnv{EnvToken: " raw "}.lookup, file)
	if got.Token != (Value{Value: "raw", Source: FromEnv}) {
		t.Errorf("Token = %+v, want raw from env", got.Token)
	}
	if none := Resolve(Overrides{}, fakeEnv{}.lookup, file); none.Token != (Value{}) {
		t.Errorf("Token without env = %+v, want unset", none.Token)
	}
}

func TestResolveIgnoresBlankEnvironment(t *testing.T) {
	got := Resolve(Overrides{}, fakeEnv{EnvVault: "  "}.lookup, Settings{})
	if got.Vault != (Value{}) {
		t.Errorf("Vault = %+v, want unset", got.Vault)
	}
}
