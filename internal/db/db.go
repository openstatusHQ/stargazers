package db

import (
	"path/filepath"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	_ "thibaultleouay.dev/stargazers/migrations"
)

func New(path string) *sqlx.DB {

	fn := filepath.Join(".", path)

	db, err := sqlx.Open("sqlite", fn)

	if err != nil {
		panic(err)
	}

	if err := goose.SetDialect("sqlite"); err != nil {
		panic(err)
	}
	goose.SetLogger(goose.NopLogger())
	if err := goose.Up(db.DB, "."); err != nil {
		panic(err)
	}
	return db

}
