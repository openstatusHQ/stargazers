package main

import (
	"context"
	"log"
	"os"

	"github.com/urfave/cli/v3"
	"thibaultleouay.dev/stargazers/internal/stargazers"
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
			&cli.StringFlag{
				Name:        "file",
				Usage:       "The output file",
				Aliases:     []string{"f"},
				Value:       "stargazers.csv",
				DefaultText: "stargazers.csv",
			},
		},
		Action: stargazers.Stargazers,
	}
	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
