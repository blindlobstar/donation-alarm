package database

import (
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jmoiron/sqlx"
)

type Repo struct {
	DB *sqlx.DB
}

func (r Repo) Migrate() {
	d, err := postgres.WithInstance(r.DB.DB, &postgres.Config{})
	if err != nil {
		log.Fatal("Error migrating database")
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		d,
	)
	if err != nil {
		log.Fatal("Error migrating database")
	}

	if err := m.Up(); err != nil {
		log.Fatal("Error migrating database")
	}
}
