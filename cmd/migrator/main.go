package main

import (
	"errors"
	"flag"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var dbAddress, dbName, dbUser, dbPassword, migrationsPath string

	flag.StringVar(&dbAddress, "db-address", "localhost:5432", "database address")
	flag.StringVar(&dbName, "db-name", "", "database name")
	flag.StringVar(&dbUser, "db-user", "", "database user")
	flag.StringVar(&dbPassword, "db-password", "", "database password")
	flag.StringVar(&migrationsPath, "migrations-path", "", "path to folder with migrations user")

	flag.Parse()

	if dbAddress == "" {
		panic("need db-address")
	}
	if dbName == "" {
		panic("need db-name")
	}
	if dbUser == "" {
		panic("need db-user")
	}
	if dbPassword == "" {
		panic("need db-password")
	}
	if migrationsPath == "" {
		panic("need migrationsPath")
	}

	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", dbUser, dbPassword, dbAddress, dbName),
	)

	if err != nil {
		panic(err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")
			return
		}

		panic(err)
	}
}