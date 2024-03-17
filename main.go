package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/gopasspw/gopass/pkg/gopass/api"
	"github.com/urfave/cli/v2"
)

const (
	name = "gopass-double-banger"
)

// Version is the released version of gopass.
var version string

func main() {
	ctx := context.Background()

	// trap Ctrl+C and call cancel on the context
	ctx, cancel := context.WithCancel(ctx)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	defer func() {
		signal.Stop(sigChan)
		cancel()
	}()
	go func() {
		select {
		case <-sigChan:
			cancel()
		case <-ctx.Done():
		}
	}()

	_, err := api.New(ctx)
	if err != nil {
		fmt.Printf("Failed to initialize gopass API: %s\n", err)
		os.Exit(1)
	}

	app := cli.NewApp()
	app.Name = name
	app.Version = getVersion().String()
	app.Usage = "Gopass integration for nested secrets"
	app.EnableBashCompletion = true
	app.Commands = []*cli.Command{
		{
			Name:  "show",
			Usage: "Decrypt and show a nested secret",
			Description: "" +
				"This command will decrypt a nested secret and show it.",
			Action: func(c *cli.Context) error {
				return nil
			},
			Flags: []cli.Flag{
			},
		},
		{
			Name:  "insert",
			Usage: "Insert a nested secret",
			Description: "" +
				"This command will insert a nested secret into the store.",
			Action: func(c *cli.Context) error {
				return nil
			},
			Flags: []cli.Flag{
			},
		},
		{
			Name: "version",
			Action: func(c *cli.Context) error {
				cli.VersionPrinter(c)

				return nil
			},
		},
	}

	if err := app.RunContext(ctx, os.Args); err != nil {
		log.Fatal(err)
	}
}
