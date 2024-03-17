package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/ProtonMail/gopenpgp/v2/helper"
	"github.com/gopasspw/gopass/pkg/gopass/api"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/urfave/cli/v2"
	"github.com/manifoldco/promptui"
)

const (
	name = "gopass-double-banger"
)

// Version is the released version of gopass.
var version string

func readPassword(prompt string) ([]byte, error) {
	prompter := promptui.Prompt{
		Label: prompt,
		Mask:  '*',
		Stdout: os.Stderr,
	}

	result, err := prompter.Run()
	if err != nil {
		return nil, err
	}

	return []byte(result), nil
}

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

	gp, err := api.New(ctx)
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
				// Check if at least one file argument is provided
				if c.NArg() < 1 {
				    return fmt.Errorf("at least one file is required")
				}

				password, err := readPassword(c.String("prompt"))

				if err != nil {
					return err
				}

				// Process each file argument
				for i := 0; i < c.NArg(); i++ {
					file := c.Args().Get(i)

					// Reading secrets by their name and revision from within the store.
					sec, err := gp.Get(ctx, file, "latest")
					if err != nil {
						panic(err)
					}

					message, err := helper.DecryptMessageWithPassword(password, string(sec.Bytes())) 
					if err != nil {
						fmt.Errorf("Failed to decrypt path: %s\n", file)
					}
					fmt.Print(message)
				}

				return nil
			},
			Flags: []cli.Flag{
				// Prompt for password request
				&cli.StringFlag{
					Name:    "prompt",
					Aliases: []string{"p"},
					Usage:   "Prompt for password",
					Value: "Password",
				},
			},
		},
		{
			// gopass-double-banger insert --prompt="Password" [secret-store-path] [plaintext-file]
			Name:  "insert",
			Usage: "Insert a nested secret",
			Description: "" +
				"This command will insert a nested secret into the store.",
			Action: func(c *cli.Context) error {
				if c.NArg() != 2 {
					return fmt.Errorf("two arguments are required: [secret-store-path] [plaintext-file]")
				}

				password, err := readPassword(c.String("prompt"))

				if err != nil {
					return err
				}

				secretStorePath := c.Args().Get(0)
				plaintextFile := c.Args().Get(1)

				// Read plaintext file
				plaintext, err := os.ReadFile(plaintextFile)
				if err != nil {
					return err
				}

				// Encrypt data with password
				armor, err := helper.EncryptMessageWithPassword(password, string(plaintext))
				if err != nil {
					return err
				}

				content := secrets.New()
				content.SetPassword(armor)

				// Write the encrypted data to the store
				if err := gp.Set(ctx, secretStorePath, content); err != nil {
					return err
				}

				return nil
			},
			Flags: []cli.Flag{
				// Prompt for password request
				&cli.StringFlag{
					Name:    "prompt",
					Aliases: []string{"p"},
					Usage:   "Prompt for password",
					Value: "Password",
				},
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
