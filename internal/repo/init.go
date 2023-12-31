package repo

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mahmoud-eskandari/gojen/internal/config"
)

var DB *sql.DB

func Connect() error {
	dbc := config.Conf.DB
	dbc.Schema = "information_schema"

	db, err := sql.Open("mysql", dbc.String())

	if err != nil {
		return err
	}

	DB = db
	return nil
}
