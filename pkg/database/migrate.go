package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(db *sql.DB) {

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		panic(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://./internal/migrations",
		"postgres",
		driver,
	)

	if err != nil {
		panic(err)
	}

	err = m.Up()

	if err != nil && err != migrate.ErrNoChange {
		panic(err)
	}

	fmt.Println("✅ migrations applied")
}
