package main

import (
	"flag"
	"log"

	"order-services/internal/db"
)

func main() {
	var dir string
	flag.StringVar(&dir, "dir", "", "path to the migrations directory (defaults to MIGRATIONS_DIR or ./migrations)")
	flag.Parse()

	gormDB, err := db.ConnectPostgres()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	sqlDB, err := gormDB.DB()
	if err == nil {
		defer sqlDB.Close()
	}

	if err := db.RunMigrations(gormDB, dir); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
}
