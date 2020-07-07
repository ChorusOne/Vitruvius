// This is the default block iterator, it will attempt to take snapshots once a
// day on the first block of the day.

package extractor

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/ChorusOne/Hippias/cmd/hippias/types"
	"github.com/ChorusOne/Hippias/pkg/oasis"
)

var (
	_ BlockIterator = &SnapshotIterator{}
)

// SnapshotIterator is an iterator that attempts to take daily snapshots of
type SnapshotIterator struct {
	lastObserved time.Time
	config       *types.Config
	state        types.State
}

func NewSnapshotIterator(config *types.Config, state types.State) *SnapshotIterator {
	return &SnapshotIterator{
		config: config,
		state:  state,
	}
}

func (self *SnapshotIterator) Process(snapshot StateSnapshot) {
	snapshotTransactions(self.config, self.state, snapshot.Block, snapshot.Transactions)
	snapshotEvents(self.config, self.state, snapshot.Block, snapshot.Events)
	snapshotCommission(self.config, self.state, snapshot.Block)
	if isDailyBlock(self.lastObserved, snapshot.Block.Time) {
		go snapshotState(self.config, self.state, snapshot.Block)
	}

	self.lastObserved = snapshot.Block.Time
}

// Internal Extractor Functions
// -----------------------------------------------------------------------------

// isDailyBlock checks that the block being observed, is the first block of the
// day. To do this check, the last observed block time must be passed in.
func isDailyBlock(previous time.Time, next time.Time) bool {
	midnight := time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, time.UTC)
	return previous.Before(midnight) && (next.After(midnight) || next.Equal(midnight))
}

// snapshotCommission collects Commission information for validators over time.
func snapshotCommission(config *types.Config, state types.State, block oasis.Block) {
	frozenAPI := state.Api.AtHeight(block.Height)
	addresses := frozenAPI.Accounts()

	for _, address := range addresses {
		commission, base, err := frozenAPI.GetValidatorCommission(address)
		if commission != nil && !commission.IsZero() {
			log.Printf("Commission: %v: %v, %v, %v", address, commission, base, err)
			var hundred oasis.Amount
			hundred.FromInt64(100)
			commission.Mul(&hundred)
			commission.Quo(base)
			if _, err = state.Dot.Exec(state.Db, "insertValidatorCommission",
				commission.String(),
				address.String(),
				block.Height,
			); err != nil {
				log.Printf("Commission Insert for %v Failed, %v", address, err)
			}
		}
	}
}

// snapshotState persists the entire current state of all accounts with nonzero
// balance on the oasis network. This is quite slow so this is done only once a
// day, for the first block of the day.
func snapshotState(config *types.Config, state types.State, block oasis.Block) {
	log.Printf("Snapshot Triggered at %s", block.Time)
	frozenAPI := state.Api.AtHeight(block.Height)
	now := time.Now()

	// Extract all accounts from the state.
	addresses := frozenAPI.Accounts()
	for i, address := range addresses {
		mappedAccount, err := frozenAPI.Account(address)

		// Extract all delegations from the state.
		// allDelegations := frozenAPI.Delegations()
		if err != nil {
			log.Printf("Failed to Retrieve Account: %s, %v", address, err)
			continue
		}

		accountDelegations := frozenAPI.AccountDelegations(address)
		encodedDelegations, err := json.Marshal(&accountDelegations)
		if err != nil {
			log.Printf("Failed to Encode Delegations: %s", address)
			log.Printf("%v", err)
			continue
		}

		stakedJson, err := json.Marshal(&mappedAccount.StakedBalance)
		if err != nil {
			log.Printf("Failed to Encode Balance: %s", address)
			log.Printf("%v", err)
			continue
		}

		debondingJson, err := json.Marshal(&mappedAccount.DebondingBalance)
		if err != nil {
			log.Printf("Failed to Encode Balance: %s", address)
			log.Printf("%v", err)
			continue
		}

		// Get Last Reward Balance
		row, err := state.Dot.QueryRow(state.Db, "queryGetLastRewardBalance",
			address.String(),
			block.Height,
		)
		if err != nil {
			log.Printf("Failed to Fetch Reward Balance: %s", address)
			log.Printf("%v", err)
			continue
		}

		var tokens int
		err = row.Scan(&tokens)
		if err != nil {
			log.Printf("User has no Reward Entries, %s: %v", address, err)
			continue
		}

		// Snapshot Account Itself
		oasis.PushQuery("insertSnapshot",
			address.String(),
			mappedAccount.Balance,
			stakedJson,
			debondingJson,
			tokens,
			encodedDelegations,
			false,
			false,
			block.Height,
			block.Time,
		)

		log.Println("Wrote ok!")

		// Print Progress
		if i%(len(addresses)/20) == 0 {
			log.Printf("Snapshot %d%% complete", 5*(i/(len(addresses)/20)))
		}
	}

	elapsed := time.Since(now)
	log.Printf("Snapshot finished, took: %s", elapsed)
}

func snapshotTransactions(config *types.Config, state types.State, block oasis.Block, txs []oasis.Transaction) {
	for _, tx := range txs {
		var encodedTx []byte
		var err error
		if encodedTx, err = json.Marshal(&tx.Payload); err != nil {
			log.Printf("Failed to Encode Tx: %v, %v", tx.Method, err)
			continue
		}

		if _, err = state.Dot.Exec(state.Db, "insertTransaction",
			tx.Method,
			encodedTx,
			block.Height,
			block.Time,
			tx.Sender.String(),
			tx.Fee.String(),
			tx.Gas,
			tx.GasPrice.String(),
			tx.Hash,
		); err != nil {
			log.Printf("Failed to Persist Tx: %v, %v", tx.Method, err)
			continue
		}

		log.Printf("Persisted Tx: %v", tx.Method)
	}
}

func snapshotFullState(config *types.Config, state types.State, block oasis.Block) {
	log.Printf("Full Snapshot entire Oasis State at height %d", block.Height)
	now := time.Now()
	elapsed := time.Since(now)
	log.Printf("Full Snapshot finished, took: %s", elapsed)
}

// snapshotEvents persists each individual event that occurs on the network.
func snapshotEvents(config *types.Config, state types.State, block oasis.Block, events []oasis.StakingEvent) {
	for _, event := range events {
		log.Printf("Event Observed: %v", event)

		switch {
		// Transfer events occur when balance is moved from one address balance to
		// another.
		case event.Transfer != nil:
			if _, err := state.Dot.Exec(state.Db, "insertTransfer",
				event.Transfer.From.String(),
				event.Transfer.To.String(),
				event.Transfer.Tokens.String(),
				event.Transfer.Hash,
				block.Height,
				block.Time,
			); err != nil {
				fmt.Printf("Failed Transfer Insert: %v\n", err)
			}

		// Burn events occur when someone is slashed, this tells us who was slashed
		// and by how much.
		case event.Burn != nil:
			if _, err := state.Dot.Exec(state.Db, "insertBurn",
				event.Burn.Owner.String(),
				event.Burn.Tokens.String(),
				event.Burn.Hash,
				block.Height,
				block.Time,
			); err != nil {
				fmt.Printf("Failed Burn Insert: %v\n", err)
			}

		// Escrow occurs whenever a delegation is modified.
		case event.Escrow != nil:
			switch {
			case event.Escrow.Add != nil:
				if _, err := state.Dot.Exec(state.Db, "insertEscrowEvent", "add",
					event.Escrow.Add.Owner.String(),
					event.Escrow.Add.Escrow.String(),
					event.Escrow.Add.Tokens.String(),
					event.Escrow.Add.Hash,
					block.Height,
					block.Time,
				); err != nil {
					fmt.Printf("Failed Escrow Add Insert: %v\n", err)
				}

			case event.Escrow.Take != nil:
				if _, err := state.Dot.Exec(state.Db, "insertEscrowEvent", "take",
					event.Escrow.Take.Owner.String(),
					"",
					event.Escrow.Take.Tokens.String(),
					event.Escrow.Take.Hash,
					block.Height,
					block.Time,
				); err != nil {
					fmt.Printf("Failed Escrow Take Insert: %v\n", err)
				}

			case event.Escrow.Reclaim != nil:
				if _, err := state.Dot.Exec(state.Db, "insertEscrowEvent", "add",
					event.Escrow.Reclaim.Owner.String(),
					event.Escrow.Reclaim.Escrow.String(),
					event.Escrow.Reclaim.Tokens.String(),
					event.Escrow.Reclaim.Hash,
					block.Height,
					block.Time,
				); err != nil {
					fmt.Printf("Failed Escrow Reclaim Insert: %v\n", err)
				}
			}
		}
	}
}
