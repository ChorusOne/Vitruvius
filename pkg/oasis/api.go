// API is an interface that describes a method of interfacing with an Oasis
// based blockchain. The details of how this works are abstracted away from the
// consumer of this package. The current and only implementation can be found
// in grpc.go

package oasis

// API implements this packages API interface. All methods automatically return
// data about the current synced blockchain height. See `AtHeight` which can be
// used to get an instance fixed to a specific height.
type API interface {
	// Utility Functions
	AtHeight(Height) API
	DecodeKey(string) (Address, error)

	// General Chain Information
	Account(Address) (*Account, error)
	AccountDelegations(Address) []Delegation
	Accounts() []Address
	Delegations() []Delegation
	GetBlock() (Block, Height)
	GetEvents() []StakingEvent
	GetGenesisState() (*Genesis, error)
	GetTransactions() []Transaction
	GetValidatorCommission(Address) (*Amount, *Amount, error)
	Pool() (*Pool, error)

	// Create Live Subscriptions to Blockchain Data
	WatchBlocks() (chan Block, error)
	WatchStakingEvents() (chan StakingEvent, error)
}
