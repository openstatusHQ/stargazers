package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upAddSyncCursor, downAddSyncCursor)
}

func upAddSyncCursor(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec("ALTER TABLE repository ADD COLUMN last_cursor text")
	if err != nil {
		return err
	}
	_, err = tx.Exec("ALTER TABLE repository ADD COLUMN last_synced_at integer")
	if err != nil {
		return err
	}
	_, err = tx.Exec("CREATE UNIQUE INDEX idx_user_repo ON users_to_repositories(user_id, repository_id)")
	if err != nil {
		return err
	}
	return nil
}

func downAddSyncCursor(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec("DROP INDEX idx_user_repo")
	if err != nil {
		return err
	}
	_, err = tx.Exec("ALTER TABLE repository DROP COLUMN last_cursor")
	if err != nil {
		return err
	}
	_, err = tx.Exec("ALTER TABLE repository DROP COLUMN last_synced_at")
	if err != nil {
		return err
	}
	return nil
}
