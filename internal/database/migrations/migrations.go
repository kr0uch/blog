package migrations

import (
	"database/sql"
	"errors"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
)

func Up(db *sql.DB) (int, error) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return -1, err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return -1, err
	}

	version, _, err := m.Version()
	if err != nil {
		return -1, err
	}
	err = m.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return int(version), nil
		}
		return -1, err
	}

	version, _, err = m.Version()
	if err != nil {
		return 0, err
	}
	return int(version), nil
}
