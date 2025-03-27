package action

import (
	"context"
	"fmt"
	"os"

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
	tx := database.MustBegin()
	for _, repo := range c.Repositories {
		tx.MustExec("INSERT INTO repository(owner, name) Values ($1, $2)", repo.Owner, repo.Name)
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "Your database is ready")
	return nil
}
