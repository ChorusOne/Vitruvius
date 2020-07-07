// Provide Events related endpoints

package endpoints

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/ChorusOne/Hippias/cmd/hippias/rest/middleware"
	"github.com/ChorusOne/Hippias/cmd/hippias/types"
	"github.com/ChorusOne/Hippias/pkg/oasis"
	"github.com/go-chi/chi"
)

// EventList returns a list of all events related to the Oasis chain. This list
// is presented in a discriminated union format.
func EventList(state types.State) Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		var results *sql.Rows
		var err error
		pagination := middleware.GetPagination(r)

		// Check if we should filter by account.
		if accountID := chi.URLParam(r, "accountID"); accountID != "" {
			if results, err = state.Dot.Query(state.Db, "queryAllEventsFiltered", "%"+accountID+"%", (pagination.Page * pagination.Limit), pagination.Limit); err != nil {
				log.Printf("EventList failed to query events, %v", err)
				return
			}
		} else {
			log.Printf("EventList: Querying 3\n")
			if results, err = state.Dot.Query(state.Db, "queryAllEvents", (pagination.Page * pagination.Limit), pagination.Limit); err != nil {
				log.Printf("EventList failed to query events, %v", err)
				return
			}
		}

		events := make([]oasis.StakingEvent, 0)
		for results.Next() {
			var kind string
			var when string
			var height int64
			var payload string

			// Scan Row into Parts
			if err := results.Scan(&height, &when, &kind, &payload); err != nil {
				log.Printf("EventList: Failed to decode Event from DB (%v)", err)
				return
			}

			// Construct Event
			log.Printf("Event: %v\n", payload)
			switch kind {
			case "transfer":
				var decoded oasis.TransferEvent
				err := json.Unmarshal([]byte(payload), &decoded)
				log.Printf("Unmarshal Error: %v", err)
				events = append(events, oasis.StakingEvent{Transfer: &decoded})
			case "burn":
				var decoded oasis.BurnEvent
				err := json.Unmarshal([]byte(payload), &decoded)
				log.Printf("Unmarshal Error: %v", err)
				events = append(events, oasis.StakingEvent{Burn: &decoded})
			case "escrow":
				var decoded oasis.EscrowEvent
				err := json.Unmarshal([]byte(payload), &decoded)
				log.Printf("Unmarshal Error: %v", err)
				events = append(events, oasis.StakingEvent{Escrow: &decoded})
			}
		}

		if err := json.NewEncoder(w).Encode(events); err != nil {
			log.Printf("%v", err)
		}
	}
}

type RpcTransaction struct {
	Hash     string      `json:"hash"`      // Hash of Transaction Bytes
	Fee      string      `json:"fee"`       // Amount Paid for Tx
	GasPrice uint64      `json:"gas_price"` // Implied Gas Price
	Gas      uint64      `json:"gas"`       // Max Gas Allowed
	Method   string      `json:"method"`    // Which Method does this Tx invoke.
	Sender   string      `json:"sender"`    // Who submitted the Tx
	When     string      `json:"date"`      // When the Tx was processed.
	Height   uint64      `json:"height"`    // Height tX was seen at.
	Payload  interface{} `json:"data"`      // Actual Payload of TX
}

func TransactionList(state types.State) Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		var results *sql.Rows
		var err error
		pagination := middleware.GetPagination(r)

		// We might receive a Transaction Hash as a paran.
		txHash := r.URL.Query().Get("hash")
		if txHash != "" {
			TransactionListByHash(state, txHash)(w, r)
			return
		}

		// Check if we should filter by account.
		if accountID := chi.URLParam(r, "accountID"); accountID != "" || txHash != "" {
			log.Println(accountID)
			if results, err = state.Dot.Query(state.Db, "queryAllTransactionsFiltered", "%"+accountID+"%", (pagination.Page * pagination.Limit), pagination.Limit); err != nil {
				log.Printf("TransactionList failed to query events, %v", err)
				return
			}
		} else {
			if results, err = state.Dot.Query(state.Db, "queryAllTransactions", (pagination.Page * pagination.Limit), pagination.Limit); err != nil {
				log.Printf("TransactionList failed to query events, %v", err)
				return
			}
		}

		transactions := make([]RpcTransaction, 0)
		for results.Next() {
			var id uint64
			var method string
			var payload string
			var height uint64
			var when string
			var sender string
			var fee string
			var gas uint64
			var gasPrice uint64
			var hash string

			// Scan Row into Parts
			if err := results.Scan(&id, &when, &fee, &gas, &gasPrice, &hash, &height, &method, &payload, &sender); err != nil {
				log.Printf("TransactionList: Failed to decode Event from DB (%v)", err)
				return
			}

			var decoded interface{}
			json.Unmarshal([]byte(payload), &decoded)
			transactions = append(transactions, RpcTransaction{
				Fee:      fee,
				Gas:      gas,
				GasPrice: gasPrice,
				Hash:     hash,
				Height:   height,
				Method:   method,
				Payload:  decoded,
				Sender:   sender,
				When:     when,
			})
		}

		if err := json.NewEncoder(w).Encode(transactions); err != nil {
			log.Printf("%v", err)
		}
	}
}

func TransactionListByHash(state types.State, txHash string) Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Parsing: %v", txHash)
		row, err := state.Dot.QueryRow(state.Db, "querySpecificTransaction", txHash)
		if err != nil {
			log.Printf("TransactionListByHash: No Query for Hash: %v\n", txHash)
			return
		}

		var id uint64
		var method string
		var payload string
		var height uint64
		var when string
		var sender string
		var fee string
		var gas uint64
		var gasPrice uint64
		var hash string

		if err := row.Scan(&id, &when, &fee, &gas, &gasPrice, &hash, &height, &method, &payload, &sender); err != nil {
			log.Printf("TransactionListByHash: Failed to decode Event from DB (%v)", err)
			return
		}

		var decoded interface{}
		json.Unmarshal([]byte(payload), &decoded)
		if err := json.NewEncoder(w).Encode(RpcTransaction{
			Fee:      fee,
			Gas:      gas,
			GasPrice: gasPrice,
			Hash:     hash,
			Height:   height,
			Method:   method,
			Payload:  decoded,
			Sender:   sender,
			When:     when,
		}); err != nil {
			log.Printf("%v", err)
		}
	}
}
