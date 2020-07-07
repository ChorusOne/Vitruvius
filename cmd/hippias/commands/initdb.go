// This file provides a command to apply the required schema in a postgres
// database. This currently does not handle migrations, and so also provides
// the ability to drop all schema as a stop-gap.
//
// The result of this file is the following two useful commands:
//
// ```bash
// $ # Write Schema into database specified in VIRT_DB env var.
// $ vitruvius initdb
// $
// $ # Same thing, but drop all existing Schema first.
// $ vitruvius initdb --reset
// ```

package commands

import (
	"database/sql"
	"io/ioutil"
	"log"

	"github.com/spf13/cobra"

	"github.com/ChorusOne/Hippias/cmd/hippias/types"
)

// Command line flag variables.
var (
	VarReset bool
)

// InitDB creates the cobra struct and flags required for the `initdb` command.
func InitDB(config *types.Config) *cobra.Command {
	command := &cobra.Command{
		Use:   "initdb",
		Short: "Initialize Postgres",
		Long:  "",
		Run:   InitDBHandler(config),
	}

	command.Flags().BoolVarP(&VarReset, "reset", "r", false, "wipe database first")
	return command
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

// InitDBHandler does the heavy lifting of connecting (wiping) and writing to a
// postgres instance.
func InitDBHandler(config *types.Config) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		var err error
		defer func() {
			if r := recover(); r != nil {
				log.Fatalf("InitDB failed, %v", r)
			}
		}()

		// Attempt to open the SQL definition file. It's expected to be in
		// the current working directory.
		schema, err := ioutil.ReadFile("sql/schema.sql")
		check(err)

		// Connect to DB specified by VIRT_DB env var.
		con, err := sql.Open("postgres", config.DatabasePath)
		check(err)
		defer con.Close()

		// If `--reset` specified, wipe the DB schema.
		if VarReset {
			_, err = con.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;")
			check(err)
		}

		// Execute schema write in one single round-trip.
		_, err = con.Exec(string(schema))
		check(err)

		log.Printf("Success! Schema written.")
	}
}
