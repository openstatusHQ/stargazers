package action

import (
	"context"

	"github.com/pressly/goose/v3"
	"github.com/urfave/cli/v3"

	_ "modernc.org/sqlite"
	_ "thibaultleouay.dev/stargazers/migrations"
)

func Migrate(ctx context.Context, cmd *cli.Command) error {

	db, err := goose.OpenDBWithDriver("sqlite", "file:./db")
	if err != nil {
		return err
		// log.Fatalf("goose: failed to open DB: %v", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			return
			// log.Fatalf("goose: failed to close DB: %v", err)
		}
	}()

	if err := goose.Up(db, "."); err != nil {
		return err
	}
	return nil
}
