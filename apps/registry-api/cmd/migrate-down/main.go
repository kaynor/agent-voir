package main

import (
	"log"
	"os"

	"github.com/agentvoir/agentvoir/apps/registry-api/internal/postgres"
)

func main() {
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		log.Fatal("POSTGRES_DSN is required")
	}

	if err := postgres.MigrateDown(dsn); err != nil {
		log.Fatalf("migration rollback failed: %v", err)
	}
	log.Println("postgres migrations rolled back successfully")
}
