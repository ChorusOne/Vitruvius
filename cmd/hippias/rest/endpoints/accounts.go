// Provide Account related endpoints.

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

// Handler is a small type that helps to make returning closures in the functions
// below less verbose.
type Handler = func(w http.ResponseWriter, r *http.Request)

// AccountList returns a list of addresses in Oasis format, these are returned
// as a JSON list of strings.
func AccountList(state types.State) Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		accounts := state.Api.Accounts()
		from, to := middleware.PaginateList(r, len(accounts))
		paginatedAccounts := accounts[from:to]
		if err := json.NewEncoder(w).Encode(paginatedAccounts); err != nil {
			log.Printf("%v", err)
		}
	}
}

// AccountListDescribed is similar to the above function, but also requests
// related account data at the same time. This produces a list of delegations
// and metadata along with the account.
func AccountListDescribed(state types.State) Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		addresses := state.Api.Accounts()
		pool, _ := state.Api.Pool()
		accountsMapped := make([]oasis.Account, len(addresses))
		from, to := middleware.PaginateList(r, len(addresses))
		for i, address := range addresses[from:to] {
			mappedAccount, err := state.Api.Account(address)
			if err != nil {
				continue
			}
			if len(mappedAccount.Delegations) > 0 {
				log.Printf("Looking Up: %s\n", address)
				log.Printf("Delegations for %s:\n", address.String())
				for _, delegation := range mappedAccount.Delegations {
					log.Printf("  To: %s\n", delegation.Validator)
					log.Printf("  Amount: %s\n", delegation.Amount)
					log.Printf("  Pool: %s\n", pool)
				}
			}
			accountsMapped[i] = *mappedAccount
		}
		if err := json.NewEncoder(w).Encode(accountsMapped); err != nil {
			log.Println(err)
		}
	}
}

// Account requests data for a specific account, but also deals with pulling
// related information and Oasis' weird account encoding.
func Account(state types.State) Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		if accountID := chi.URLParam(r, "accountID"); accountID != "" {
			key, _ := state.Api.DecodeKey(accountID)
			accountInfo, _ := state.Api.Account(key)
			if err := json.NewEncoder(w).Encode(accountInfo); err != nil {
				log.Println(err)
			}
		}
	}
}

// HistoryAccount mirrors an oasis Account, but includes a timestamp so the log
// of accounts can be consumed for graphing by Anthem.
type HistoryAccount struct {
	oasis.Account
	Date    string `json:"date"`
	Rewards string `json:"rewards"`
}

// AccountHistory provides a full history of every snapshot taken of a users
// account. Currently not limited as there's not much data, but a cap can be
// added later on.
func AccountHistory(state types.State) Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		if accountID := chi.URLParam(r, "accountID"); accountID != "" {
			var snapshotLength int
			var results *sql.Rows
			var err error

			// Get AccountHistory Length
			row, err := state.Dot.QueryRow(state.Db, "queryAccountHistoryLength", accountID)
			if err != nil {
				log.Printf("AccountHistory: Failed to query History length, %v", err)
				return
			}
			row.Scan(&snapshotLength)

			from, to := middleware.PaginateList(r, snapshotLength)
			if results, err = state.Dot.Query(state.Db, "queryAccountHistory", accountID, from, to); err != nil {
				log.Printf("AccountHistory: Failed to query History, %v", err)
				return
			}

			snapshots := make([]HistoryAccount, 0)
			for results.Next() {
				var id uint64
				var address string
				var balance string
				var stakingBalance string
				var debondingBalance string
				var rewards string
				var delegations string
				var isValidator bool
				var isDelegator bool
				var height int64
				var when string
				if err := results.Scan(&id, &address, &balance, &stakingBalance, &debondingBalance, &rewards, &delegations, &isValidator, &isDelegator, &height, &when); err != nil {
					log.Printf("AccountHistory: Failed to decode Account, %v", err)
					return
				}

				// Decode JSON Part
				var delegationsDecoded []oasis.Delegation
				json.Unmarshal([]byte(delegations), &delegationsDecoded)
				key, _ := state.Api.DecodeKey(address)

				var stakingJSON oasis.SharePool
				var debondingJSON oasis.SharePool
				json.Unmarshal([]byte(stakingBalance), &stakingJSON)
				json.Unmarshal([]byte(debondingBalance), &debondingJSON)

				// Decode Account as JSON
				snapshots = append(snapshots, HistoryAccount{
					Account: oasis.Account{
						Address:          key,
						Balance:          balance,
						Delegations:      delegationsDecoded,
						StakedBalance:    &stakingJSON,
						DebondingBalance: &debondingJSON,
						Height:           height,
						Meta: oasis.AccountMeta{
							IsValidator: isValidator,
							IsDelegator: isDelegator,
						},
					},
					Date:    when,
					Rewards: rewards,
				})
			}

			if err := json.NewEncoder(w).Encode(snapshots); err != nil {
				log.Println(err)
			}
		}
	}
}
