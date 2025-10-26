package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var (
		dbURL   string
		path    string
		command string // up|down|version
	)

	flag.StringVar(&dbURL, "db-url", "", "Postgres URL, e.g. postgres://user:pass@host:5432/db?sslmode=disable")
	flag.StringVar(&path, "path", "./migrations", "path to migrations folder")
	flag.StringVar(&command, "command", "up", "command: up|down|version")
	flag.Parse()

	if dbURL == "" {
		log.Fatal("-db-url is required")
	}
	if path == "" {
		log.Fatal("-path is required")
	}

	src := "file://" + strings.TrimRight(path, "/")
	m, err := migrate.New(src, dbURL)
	if err != nil {
		log.Fatalf("init migrator: %v", err)
	}

	switch command {
	case "up":
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatalf("migrate up: %v", err)
		}
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")
		} else {
			fmt.Println("migrations applied")
		}

	case "down":
		if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatalf("migrate down: %v", err)
		}
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to rollback")
		} else {
			fmt.Println("rolled back all migrations")
		}

	case "version":
		v, dirty, err := m.Version()
		if errors.Is(err, migrate.ErrNilVersion) {
			fmt.Println("version: none (database empty)")
			os.Exit(0)
		}
		if err != nil {
			log.Fatalf("version: %v", err)
		}
		fmt.Printf("version: %d (dirty=%v)\n", v, dirty)

	default:
		log.Fatalf("unknown -command: %q (use up|down|version)", command)
	}
}
