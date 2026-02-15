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
		dbURL       = flag.String("db", "", "Database URL (or set DATABASE_URL env)")
		libraryDir  = flag.String("library-dir", "", "Path to services-library root (contains compose/ and config/)")
		composeDir  = flag.String("compose-dir", "", "Path to compose directory (overrides library-dir)")
		projectRoot = flag.String("project-root", "", "Project root for resolving relative paths")
		configDir   = flag.String("config-dir", "", "Path to config directory (overrides library-dir)")
		countOnly   = flag.Bool("count-only", false, "Print service count and exit")
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

	if *libraryDir == "" {
		*libraryDir = os.Getenv("LIBRARY_DIR")
	}
	if *libraryDir == "" {
		for _, candidate := range []string{
			"/workspace/services-library",
			"../services-library",
		} {
			if info, err := os.Stat(candidate); err == nil && info.IsDir() {
				*libraryDir = candidate
				break
			}
		}
	}

	if *composeDir == "" {
		*composeDir = os.Getenv("COMPOSE_DIR")
	}
	if *composeDir == "" && *libraryDir != "" {
		*composeDir = *libraryDir + "/compose"
	}
	if *composeDir == "" {
		log.Fatal("compose-dir or library-dir is required")
	}

	if *configDir == "" && *libraryDir != "" {
		candidate := *libraryDir + "/config"
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			*configDir = candidate
		}
	}

	var importer *compose.Importer
	if *projectRoot != "" {
		importer = compose.NewImporterWithRoot(db, *composeDir, *projectRoot)
	} else {
		importer = compose.NewImporter(db, *composeDir)
	}

	log.Println("importing compose files...")
	if err := importer.ImportAll(); err != nil {
		log.Fatalf("import failed: %v", err)
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM services").Scan(&count)
	log.Printf("import complete: %d services", count)

	if *configDir != "" {
		importer.SetConfigDir(*configDir)
		fileCount, err := importer.ImportAllConfigFiles()
		if err != nil {
			log.Printf("warning: config file import: %v", err)
		} else {
			log.Printf("config files imported: %d files", fileCount)
		}

		log.Println("resolving config mount links...")
		if err := importer.ResolveConfigMountLinks(); err != nil {
			log.Printf("warning: config mount link resolution: %v", err)
		}
	}
}
