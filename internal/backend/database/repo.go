package database

import "github.com/jmoiron/sqlx"

type Repo struct {
	DB *sqlx.DB
}
