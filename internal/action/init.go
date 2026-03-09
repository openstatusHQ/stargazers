package action

import (
	"context"
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/urfave/cli/v3"
	"thibaultleouay.dev/stargazers/internal/config"
	"thibaultleouay.dev/stargazers/internal/db"
)

func Init(ctx context.Context, cmd *cli.Command) error {
	fmt.Fprintln(os.Stderr, "Initializing the project")
	output := cmd.String("output")
	database := db.New(output)
	path := cmd.String("config")
	c, err := config.ReadConfig(path)
	if err != nil {
		return err
	}
	if err := doInit(database, c); err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "Your database is ready")
	return nil
}

func doInit(database *sqlx.DB, cfg *config.Config) error {
	tx := database.MustBegin()
	for _, repo := range cfg.Repositories {
		tx.MustExec("INSERT INTO repository(owner, name) Values ($1, $2)", repo.Owner, repo.Name)
	}
	return tx.Commit()
}
