package main

import (
	"database/sql"
	"flag"
	"log"
	"os"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/pressly/goose/v3"
)

var (
	dbString      string
	migrationsDir string
	command       string
	migrationType string
)

func init() {

	os.Setenv("GOOSE_CREATE_TYPE", "sql")

	defaultDB := os.Getenv("DATABASE_URL")

	flag.StringVar(&dbString, "db", defaultDB, "host=localhost port=5432 user=postgres password=232003 dbname=mydb sslmode=disable")

	flag.StringVar(&migrationsDir, "dir", "internal/migrator", "Каталог с SQL-файлами миграций")

	flag.StringVar(&command, "cmd", "up", "up, down, status, create")

	flag.StringVar(&migrationType, "type", "go", " go / sql / none")

	flag.Usage = func() {

		log.Printf("Usage of %s:\n", os.Args[0])

		flag.PrintDefaults()
	}
}

func main() {

	flag.Parse()

	if command != "create" && dbString == "" {
		log.Fatal("Error: --db flag is required")
	}

	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {

		log.Fatalf("Error: migrations directory '%s' does not exist", migrationsDir)

	}

	if command != "up" && command != "down" && command != "status" && command != "create" {

		log.Fatalf("Error: invalid command '%s'. Use 'up', 'down', 'status', or 'create'", command)

	}

	log.SetFlags(0)

	log.SetOutput(os.Stdout)

	log.Printf("Starting migrations with command: %q", command)

	log.Printf("Using migrations from: %s", migrationsDir)

	log.Printf("Connecting to DB: %s", dbString)

	var db *sql.DB
	if command != "create" {
		var err error
		db, err = goose.OpenDBWithDriver("postgres", dbString)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer func() {
			if err := db.Close(); err != nil {
				log.Printf("Warning: failed to close DB connection: %v", err)
			}
		}()
	}

	if command == "create" {
		if len(flag.Args()) < 1 {

			log.Fatalf("Error: missing migration name for 'create' command")

		}

		name := flag.Args()[0]

		os.Setenv("GOOSE_CREATE_TYPE", migrationType)

		if err := goose.Run("create", nil, migrationsDir, name); err != nil {

			log.Fatalf("Failed to create migration: %v", err)

		}

		log.Printf("Created new migration: %s (type: %s)", name, migrationType)

		return
	}

	if err := goose.Run(command, db, migrationsDir); err != nil {

		log.Fatalf("Migration failed: %v", err)

	}

	log.Printf("Migrations applied successfully with command: %q", command)
}
