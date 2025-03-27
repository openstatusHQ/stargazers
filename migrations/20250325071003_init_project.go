package migrations

import (
	"context"
	"database/sql"
	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upInitProject, downInitProject)
}

var createTableRepository = `
CREATE TABLE repository (
	id integer primary key,
	owner text NOT NULL,
	name text NOT NULL,
	created_at integer DEFAULT (unixepoch()),
	updated_at integer,
	UNIQUE(owner, name)
);
`
var dropTableRepository = `DROP TABLE repository;`

var createTableCompany = `
CREATE TABLE company (
    id integer primary key,
    avatar_url text,
    description text,
    email text,
    name text,
    location text,
    login text,
    members_ct integer,
    repositories_ct integer,
    website_url text,
    created_at integer DEFAULT (unixepoch()),
    updated_at integer,
    UNIQUE(login)
);
`
var dropTableCompany = `DROP TABLE company;`

var createTableUser = `
CREATE TABLE user (
    id integer primary key,
    avatar_url text,
    bio text,
    email text,
    enrichment_data text,
	followers_ct integer,
	following_ct integer,
    fullname text,
    is_stargazer integer,
    is_watcher integer,
    is_forker integer,
    login text NOT NULL,
    company_id integer references company(id),
    social_data text,
    created_at integer DEFAULT (unixepoch()),
    updated_at integer,
    UNIQUE(login)
);
`
var dropTableUser = `DROP TABLE user;`

var createTableUsersToRepositories = `
CREATE TABLE users_to_repositories (
	id integer primary key,
	user_id integer references user(id),
	repository_id integer references repository(id),
 	created_at integer DEFAULT (unixepoch())
);
`

var dropTableUsersToRepositories = `DROP TABLE users_to_repositories;`

var createTableSync = `
CREATE TABLE sync (
	id integer primary key,
	synced_type text,
	synced_data text,
	synced_at integer,
	created_at integer DEFAULT (unixepoch())
);`

var dropTableSync = `DROP TABLE sync;`

func upInitProject(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(createTableUser)
	if err != nil {
		return err
	}
	_, err = tx.Exec(createTableCompany)
	if err != nil {
		return err
	}
	_, err = tx.Exec(createTableRepository)
	if err != nil {
		return err
	}
	_, err = tx.Exec(createTableUsersToRepositories)
	if err != nil {
		return err
	}
	_, err = tx.Exec(createTableSync)
	if err != nil {
		return err
	}
	return nil
}

func downInitProject(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(dropTableUser)
	if err != nil {
		return err
	}
	_, err = tx.Exec(dropTableCompany)
	if err != nil {
		return err
	}
	_, err = tx.Exec(dropTableRepository)
	if err != nil {
		return err
	}
	_, err = tx.Exec(dropTableUsersToRepositories)
	if err != nil {
		return err
	}
	_, err = tx.Exec(dropTableSync)
	if err != nil {
		return err
	}

	return nil
}
