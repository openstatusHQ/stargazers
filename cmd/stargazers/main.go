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
		UsageText: "stargazers -github-token <token> -owner <owner> -name <repo>",
		Version:   "0.0.1",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "github-token",
				Usage:    "Your GitHub access token",
				Aliases:  []string{"t"},
				Sources:  cli.EnvVars("GITHUB_TOKEN"),
				Required: true,
			},

			&cli.StringFlag{
				Name:        "output",
				Usage:       "The output file",
				DefaultText: "stargazers.csv",
				// Aliases:     []string{"o"},
			},
		},
		Commands: []*cli.Command{
			{
				Name:   "insights",
				Usage:  "Get the stargazers insights",
				Action: action.Stargazers,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "owner",
						Usage:    "The owner/organization of the repository",
						Aliases:  []string{"o"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     "name",
						Usage:    "The repository name",
						Aliases:  []string{"n"},
						Required: true,
					},
				},
			},
			{
				Name:   "company",
				Usage:  "Get the stargazers company",
				Action: action.Company,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "input",
						Value: "stargazers.csv",
					},
				},
			},
		},
	}
	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
