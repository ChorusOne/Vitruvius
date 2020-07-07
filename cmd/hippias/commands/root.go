// This file provides the default command. Therefore this command can be
// considere the entry point to the app. This default command will start an
// extractor goroutine to continuously pull and store data from an Oasis chain,
// and a REST service to query it.

package commands

import (
	"database/sql"
	"log"

	"github.com/spf13/cobra"

	"github.com/ChorusOne/Hippias/cmd/hippias/extractor"
	"github.com/ChorusOne/Hippias/cmd/hippias/rest"
	"github.com/ChorusOne/Hippias/cmd/hippias/types"
	"github.com/ChorusOne/Hippias/pkg/oasis"
)

// Root creates the cobra struct required to wrap the root command.
func Root(config *types.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "hippias",
		Short: "Extractor for the Oasis blockchain.",
		Long:  "",
		Run:   RootHandler(config),
	}
}

// RootHandler wraps the main functionality of this app, it will spawn two
// goroutines (rest & extractor) then block forever.
func RootHandler(config *types.Config) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		var err error
		var api *oasis.Oasis
		var con *sql.DB

		// Initialize Oasis API, gRPC is hidden/managed by the oasis package.
		if api, err = oasis.NewOasis(config.OasisSocket); err != nil {
			log.Printf("Failed to initialize Oasis API, %v", err)
			return
		}

		// Connect to the Database, Postgres is the only supported DB right now.
		if con, err = sql.Open("postgres", config.DatabasePath); err != nil {
			log.Printf("Failed to open postgres connection, %v", err)
			return
		}

		// Construct Shared State. The dotsql construction happens here.
		state := types.NewState(api, con)

		// Setup Inlet to manage batching queries to the database.
		oasis.InitInlet(con, 1,
			func(err error) {
				log.Printf("Inlet: error occurred, %v", err)
			},
			func(batch oasis.Batch) error {
				for _, q := range batch {
					log.Printf("Execing Query: %v, %v\n", q.Query, q.Args)
					if _, err := state.Dot.Exec(state.Db, q.Query, q.Args...); err != nil {
						log.Fatalf("TRACK: Fuck")
						return err
					}
				}
				return nil
			},
		)

		go rest.StartAPI(config, state)
		go extractor.StartExtractor(config, state)

		select {}
	}
}
