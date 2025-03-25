package db

import (
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	_ "thibaultleouay.dev/stargazers/migrations"
)

func New() *sqlx.DB {

	db, err := sqlx.Open("sqlite", "file:./db")

	if err != nil {
		panic(err)
	}

	if err := goose.SetDialect("sqlite"); err != nil {
		panic(err)
	}

	if err := goose.Up(db.DB, "."); err != nil {
		panic(err)
	}
	return db

}
