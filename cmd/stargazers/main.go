package main

import (
	"context"
	"log"
	"os"

	"github.com/urfave/cli/v3"
	"thibaultleouay.dev/stargazers/internal/action"
)

func main() {

	app := &cli.Command{
		Name:      "stargazers",
		Usage:     "Get your insights from your Stargazers",
		UsageText: "stargazers [global options] command [command options] [arguments...]",
		Version:   "0.0.1",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Usage:   "Your config file",
				Aliases: []string{"c"},
				Value:   "stargazers.yaml",
			},
		},
		Commands: []*cli.Command{
			{
				Name:   "insights",
				Usage:  "Get the stargazers insights",
				Action: action.Stargazers,
			},
			{
				Name:   "sync",
				Usage:  "Sync your data",
				Action: action.Sync,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "github-token",
						Usage:    "Your GitHub access token",
						Aliases:  []string{"t"},
						Sources:  cli.EnvVars("GITHUB_TOKEN"),
						Required: true,
					},
				},
			},
			{
				Name:   "init",
				Usage:  "Initialize the project",
				Action: action.Init,
			},
			{
				Name:   "migrate",
				Usage:  "Migrate the database",
				Action: action.Migrate,
			},
		},
	}
	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
