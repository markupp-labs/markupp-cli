// Command mu is the command line client of the Markupp server.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/markupp-labs/markupp-cli/internal/cli"
	"github.com/markupp-labs/markupp-cli/internal/config"
)

// version is set at build time with -ldflags "-X main.version=...".
var version = "dev"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	code := cli.Execute(ctx, cli.Deps{
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		Stdin:   os.Stdin,
		Env:     os.LookupEnv,
		Store:   config.DefaultStore(os.LookupEnv),
		Version: version,
	}, os.Args[1:])
	stop()
	os.Exit(code)
}
