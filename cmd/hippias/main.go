package main

import (
	"fmt"
	"log"

	"github.com/ChorusOne/Hippias/cmd/hippias/commands"
	"github.com/ChorusOne/Hippias/cmd/hippias/types"
	"github.com/golang-migrate/migrate/v4"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func migrateDB(config *types.Config) {
	m, err := migrate.New(
		"file://sql/migrations",
		config.DatabasePath,
	)

	if err != nil {
		log.Fatalf("Migration Failed: %v\n", err)
	}

	// Migrate all the way until tip.
	if err := m.Up(); err != nil {
		if err != migrate.ErrNoChange {
			log.Fatalf("Migration Failed: %v\n", err)
		}
	}
}

func main() {
	// Setup CLI Commands & Handlers
	config := types.ConfigFromEnv()
	rootCommand := commands.Root(&config)
	rootCommand.AddCommand(commands.Version(&config))
	rootCommand.AddCommand(commands.InitDB(&config))

	// Process DB Migrations before entering main logic.
	migrateDB(&config)

	if err := rootCommand.Execute(); err != nil {
		fmt.Printf("Error: %v", err)
	}
}
