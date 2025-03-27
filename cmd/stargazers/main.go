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
		Usage:     "Get your insights from your GitHub repositories",
		UsageText: "stargazers [global options] command [command options] [arguments...]",
		Version:   "0.0.1",

		Commands: []*cli.Command{
			{
				Name:  "repo",
				Usage: "Manage your repositories",
				Commands: []*cli.Command{
					{
						Name:  "add",
						Usage: "Add a repository",
					},
					{
						Name:  "delete",
						Usage: "Delete a repository",
					},
					{
						Name:  "view",
						Usage: "View all your repositories",
						Action: action.RepoView,
					},
				},
			},
			{
				Name:   "sync",
				Usage:  "Sync all your repositories",
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
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "config",
						Usage:   "Your config file",
						Aliases: []string{"c"},
						Value:   "stargazers.yaml",
					},
				},
			},
			{
				Name:   "migrate",
				Usage:  "Migrate the database",
				Action: action.Migrate,
				Hidden: true,
			},
		},
		Flags: []cli.Flag{
            &cli.StringFlag{
                Name:  "output",
                Value: "db",
                Aliases: []string{"o"},
                Usage: "The name of the output Database",
            },
        },
	}
	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
