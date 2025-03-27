package action

import (
	"context"

	"github.com/pressly/goose/v3"
	"github.com/urfave/cli/v3"

	_ "modernc.org/sqlite"
	_ "thibaultleouay.dev/stargazers/migrations"
)

func Migrate(ctx context.Context, cmd *cli.Command) error {
	output := cmd.String("output")

	db, err := goose.OpenDBWithDriver("sqlite", output)
	if err != nil {
		return err
	}

	defer func() error {
		if err := db.Close(); err != nil {
			return err
			// log.Fatalf("goose: failed to close DB: %v", err)
		}
		return nil
	}()

	if err := goose.Up(db, "."); err != nil {
		return err
	}
	return nil
}
