package main

import (
	config "auth/configs"
	"auth/migrations/migrations"
	"database/sql"
	"flag"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func init() {
	config.InitConfig()
}

func main() {
	cmd := flag.String("cmd", "", "migration command: up | down | reset")
	flag.Usage = func() {
		_, err := fmt.Fprintf(flag.CommandLine.Output(),
			`Usage:
		  -cmd up      : apply all pending migrations
		  -cmd down    : roll back the last migration
		  -cmd reset   : roll back all applied migrations
		
		Examples:
		  %s -cmd up
		  %s -cmd down
		  %s -cmd reset
		`, os.Args[0], os.Args[0], os.Args[0])
		if err != nil {
			return
		}
	}
	flag.Parse()

	if *cmd == "" {
		flag.Usage()
		os.Exit(2)
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		config.Config.DB.Postgres.Username,
		config.Config.DB.Postgres.Password,
		config.Config.DB.Postgres.Host,
		config.Config.DB.Postgres.Port,
		config.Config.DB.Postgres.Database,
		config.Config.DB.Postgres.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		fmt.Printf("failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if cerr := db.Close(); cerr != nil {
			fmt.Printf("failed to close database connection: %v\n", cerr)
			os.Exit(1)
		}
	}()

	goose.SetBaseFS(migrations.FS)

	if err := goose.SetDialect("postgres"); err != nil {
		fmt.Printf("failed to set goose dialect: %v", err)
		os.Exit(1)
	}

	switch *cmd {
	case "up":
		if err := goose.Up(db, "."); err != nil {
			fmt.Printf("failed to apply migrations (up): %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Migrations applied successfully.")
	case "down":
		if err := goose.Down(db, "."); err != nil {
			fmt.Printf("failed to apply migrations (down): %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Last migration rolled back successfully.")
	case "reset":
		if err := goose.Reset(db, "."); err != nil {
			fmt.Printf("failed to reset all migrations: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("All migrations rolled back (reset) successfully.")
	default:
		flag.Usage()
		os.Exit(2)
	}
}
