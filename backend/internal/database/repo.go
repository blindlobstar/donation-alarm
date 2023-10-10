package database

import (
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
)

type Repo struct {
	DB *sqlx.DB
}

func (r Repo) Migrate() {
	d, err := postgres.WithInstance(r.DB.DB, &postgres.Config{})
	if err != nil {
		log.Fatalf("Error migrating database: %v\n", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		d,
	)
	if err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Error migrating database: %v\n", err)
	}

	if err := m.Up(); err != nil {
		log.Fatalf("Error migrating database: %v\n", err)
	}
}
