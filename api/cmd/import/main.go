package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/priz/devarch-api/internal/compose"
	_ "github.com/lib/pq"
)

func main() {
	var (
		dbURL      = flag.String("db", "", "Database URL (or set DATABASE_URL env)")
		composeDir = flag.String("compose-dir", "", "Path to compose directory")
		countOnly  = flag.Bool("count-only", false, "Print service count and exit")
	)
	flag.Parse()

	if *dbURL == "" {
		*dbURL = os.Getenv("DATABASE_URL")
	}
	if *dbURL == "" {
		*dbURL = "postgres://devarch:devarch@localhost:5432/devarch?sslmode=disable"
	}

	db, err := sql.Open("postgres", *dbURL)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping: %v", err)
	}

	if *countOnly {
		var count int
		db.QueryRow("SELECT COUNT(*) FROM services").Scan(&count)
		fmt.Println(count)
		return
	}

	if *composeDir == "" {
		*composeDir = os.Getenv("COMPOSE_DIR")
	}
	if *composeDir == "" {
		log.Fatal("compose-dir is required")
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
