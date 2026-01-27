package main

import (
	"database/sql"
	"flag"
	"log"
	"os"

	"github.com/priz/devarch-api/internal/compose"
	_ "github.com/lib/pq"
)

func main() {
	var (
		dbURL      = flag.String("db", "", "Database URL (or set DATABASE_URL env)")
		composeDir = flag.String("compose-dir", "", "Path to compose directory")
	)
	flag.Parse()

	if *dbURL == "" {
		*dbURL = os.Getenv("DATABASE_URL")
	}
	if *dbURL == "" {
		*dbURL = "postgres://devarch:devarch@localhost:5432/devarch?sslmode=disable"
	}

	if *composeDir == "" {
		*composeDir = os.Getenv("COMPOSE_DIR")
	}
	if *composeDir == "" {
		log.Fatal("compose-dir is required")
	}

	db, err := sql.Open("postgres", *dbURL)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping: %v", err)
	}

	importer := compose.NewImporter(db, *composeDir)

	log.Println("importing compose files...")
	if err := importer.ImportAll(); err != nil {
		log.Fatalf("import failed: %v", err)
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM services").Scan(&count)
	log.Printf("import complete: %d services", count)
}
