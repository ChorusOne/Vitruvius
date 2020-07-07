// This file implements a long running extractor that persists data about the
// chain on a regular interval.

package extractor

import (
	"fmt"
	"log"

	"github.com/ChorusOne/Hippias/cmd/hippias/types"
	"github.com/ChorusOne/Hippias/pkg/oasis"
)

// Main Extractor Logic
// -----------------------------------------------------------------------------

func StartExtractor(config *types.Config, state types.State) error {
	log.Print("Extractor Starting")

	var err error
	var blocks chan oasis.Block

	// Use a Blocks iterator in order to pump the event loop that runs
	// the extractor.
	if blocks, err = state.Api.WatchBlocks(); err != nil {
		return fmt.Errorf("StartExtractor: WatchBlocks failed, %w", err)
	}

	// Track the last observed block time, to make intelligent decisions around
	// when we should create snapshots. We'll pull the latest from our database
	// so we don't resync from 0 when the application is killed.
	var lastHeight oasis.Height
	if row, err := state.Dot.QueryRow(state.Db, "queryLatestSyncHeight"); err == nil {
		row.Scan(&lastHeight)
	}

	// Setup Block Iterators
	iterators := []BlockIterator{
		NewSnapshotIterator(config, state),
	}

	log.Printf("Starting Sync from %d\n", lastHeight)

	for {
		select {
		case msg := <-blocks:
			log.Printf("Height %d Observed. Last was %d. Timestamp: %s\n", msg.Height, lastHeight, msg.Time)

			// Oasis gRPC is unreliable, and skips blocks. So here we'll track
			// how far ahead the chain may have skipped without us knowing.
			blockDistance := msg.Height - lastHeight - 1

			if blockDistance > 1 {
				log.Printf("Blocks Skipped: %d", blockDistance)
			}

			// For each Block we know we've skipped (hopefully only ever 1 at a
			// time) we run block processors.
			for ; blockDistance >= 0; blockDistance-- {
				currentBlock := msg.Height - blockDistance
				log.Printf("Processing Block %d\n", currentBlock)

				api := state.Api.AtHeight(currentBlock)
				block, _ := api.GetBlock()
				events := api.GetEvents()
				transactions := api.GetTransactions()

				for _, iterator := range iterators {
					iterator.Process(StateSnapshot{
						Block:        block,
						Events:       events,
						Transactions: transactions,
					})
				}

				// Update Height
				state.Dot.Exec(state.Db, "updateLatestSyncHeight", block.Height)
			}

			lastHeight = msg.Height
		}
	}
}
