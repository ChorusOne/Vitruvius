// Define our Extractor local types, these allow us to parse data from Oasis
// upsteram into a constant format of our own. This will give us a stable API
// mostly immune to Oasis upstream.

package oasis

import (
	"time"

	"github.com/oasisprotocol/oasis-core/go/common/quantity"
	"github.com/oasisprotocol/oasis-core/go/staking/api"
)

// Useful Aliases
// -----------------------------------------------------------------------------

// Account is an alias of the Oasis quantity type, this makes it easy for us to
// change the representation within the API.
type Amount = quantity.Quantity

// Height of the chain is a simple integer.
type Height = int64

// Pool represents the total quantity of currency in the shared pool of rewards.
type Pool = quantity.Quantity

// Concrete API Types
// -----------------------------------------------------------------------------

type Genesis struct {
	Serialized []byte
}

// Address represents the unique identifier for an account, for now we just
// alias the underlying PublicKey but can change it to an encoding later.
type Address = api.Address

// Account is a type that contains all the relevent information that we care
// about in Anthem for an account, this includes delegations.
type SharePool struct {
	Balance     string `json:"balance"`
	TotalShares string `json:"shares"`
}

type Account struct {
	Address          Address      `json:"address,omitempty"`
	Balance          string       `json:"balance,omitempty"`
	DebondingBalance *SharePool   `json:"debonding_balance,omitempty"`
	Delegations      []Delegation `json:"delegations,omitempty"`
	Height           Height       `json:"height,omitempty"`
	Meta             AccountMeta  `json:"meta,omitempty"`
	StakedBalance    *SharePool   `json:"staked_balance,omitempty"`
}

// AccountMeta contains flags to make it easier for Anthem to display
// additional information about a specific account.
type AccountMeta struct {
	IsValidator bool `json:"is_validator"`
	IsDelegator bool `json:"is_delegator"`
}

// Delegation represents what Oasis calls an escrow, a sum of money from
// one account bound to a Validator.
type Delegation struct {
	Delegator Address `json:"delegator"`
	Validator Address `json:"validator"`
	Amount    Amount  `json:"amount"`
}

// Block is a type that wraps up information about a block on the Oasis
// chain. As Oasis stores this in CBOR, we can use this wrapper type to
// store decoded data.
type Block struct {
	ChainID string
	Height  Height
	Time    time.Time
}

// Transaction Types
// -----------------------------------------------------------------------------

type TransferTx struct {
	From   Address `json:"from"`
	To     Address `json:"to"`
	Tokens Amount  `json:"tokens"`
}

type RegisterEntityTx struct {
	ID                     Address   `json:"id"`
	Nodes                  []Address `json:"nodes"`
	AllowEntitySignedNodes bool      `json:"allow_entity_signed_nodes"`
}

type DeregisterEntityTx struct{}

type RegisterNodeTx struct {
	ID         Address `json:"id"`
	Entity     Address `json:"entity_id"`
	Expiration uint64  `json:"expiration"`
}

type UnfreezeNodeTx struct {
	ID Address `json:"id"`
}

type RegisterRuntimeTx struct {
	ID      []byte `json:"id"`
	Version string `json:"version"`
}

type BurnTx struct {
	From   Address `json:"from"`
	Tokens Amount  `json:"tokens"`
}

type AddEscrowTx struct {
	To     Address `json:"to"`
	Tokens Amount  `json:"tokens"`
}

type ReclaimEscrowTx struct {
	From   Address `json:"from"`
	Shares Amount  `json:"shares"`
}

type Rate struct {
	Start Height `json:"start"`
	Rate  Amount `json:"rate"`
}

type Bound struct {
	Start   Height `json:"start"`
	RateMin Amount `json:"rate_min"`
	RateMax Amount `json:"rate_max"`
}

type AmendCommissionScheduleTx struct {
	Rates  []Rate  `json:"rates"`
	Bounds []Bound `json:"bounds"`
}

type UnknownTx struct {
	Payload []byte `json:"-"`
}

// Transaction represents a deserialized oasis transaction, this format
// is also CBOR serialized within Oasis.
type Transaction struct {
	Fee      Amount      `json:"fee"`       // Amount Paid for Tx
	GasPrice Amount      `json:"gas_price"` // Implied Gas Price
	Hash     string      `json:"hash"`      // Hash of transaction bytes
	Gas      uint64      `json:"gas"`       // Max Gas Allowed
	Method   string      `json:"method"`    // Which Method does this Tx invoke.
	Sender   Address     `json:"sender"`    // Who submitted the Tx
	Payload  interface{} `json:"data"`      // Actual Payload of TX
}

type TransactionPayload struct {
	TransferTx                *TransferTx                `json:"transfer,omitempty"`
	RegisterEntityTx          *RegisterEntityTx          `json:"register_entity,omitempty"`
	DeregisterEntityTx        *DeregisterEntityTx        `json:"deregister_entity,omitempty"`
	RegisterNodeTx            *RegisterNodeTx            `json:"register_node,omitempty"`
	UnfreezeNodeTx            *UnfreezeNodeTx            `json:"unfreeze_node,omitempty"`
	RegisterRuntimeTx         *RegisterRuntimeTx         `json:"register_runtime,omitempty"`
	BurnTx                    *BurnTx                    `json:"burn,omitempty"`
	AddEscrowTx               *AddEscrowTx               `json:"add_escrow,omitempty"`
	ReclaimEscrowTx           *ReclaimEscrowTx           `json:"reclaim_escrow,omitempty"`
	AmendCommissionScheduleTx *AmendCommissionScheduleTx `json:"amend_commission_schedule,omitempty"`
	UnknownTx                 *UnknownTx                 `json:"unknown,omitempty"`
}

// Event Types
// -----------------------------------------------------------------------------

type TransferEvent struct {
	From   Address   `json:"from"`
	To     Address   `json:"to"`
	Tokens Amount    `json:"tokens"`
	Height Height    `json:"height,omitempty"`
	Date   time.Time `json:"date,omitempty"`
	Hash   string    `json:"hash,omitempty"`
}

type BurnEvent struct {
	Owner  Address           `json:"owner"`
	Tokens quantity.Quantity `json:"tokens"`
	Height Height            `json:"height,omitempty"`
	Date   time.Time         `json:"date,omitempty"`
	Hash   string            `json:"hash,omitempty"`
}

type EscrowEvent struct {
	Add     *AddEscrowEvent     `json:"add,omitempty"`
	Take    *TakeEscrowEvent    `json:"take,omitempty"`
	Reclaim *ReclaimEscrowEvent `json:"reclaim,omitempty"`
}

type UnknownEvent struct{}

type StakingEvent struct {
	Transfer *TransferEvent `json:"transfer,omitempty"`
	Burn     *BurnEvent     `json:"burn,omitempty"`
	Escrow   *EscrowEvent   `json:"escrow,omitempty"`
	Unknown  *UnknownEvent  `json:"unknown,omitempty"`
}

// EscrowEvent Discriminants

type AddEscrowEvent struct {
	Owner  Address           `json:"owner"`
	Escrow Address           `json:"escrow"`
	Tokens quantity.Quantity `json:"tokens"`
	Height Height            `json:"height,omitempty"`
	Hash   string            `json:"hash,omitempty"`
}

type TakeEscrowEvent struct {
	Owner  Address           `json:"owner"`
	Tokens quantity.Quantity `json:"tokens"`
	Height Height            `json:"height,omitempty"`
	Hash   string            `json:"hash,omitempty"`
}

type ReclaimEscrowEvent struct {
	Owner  Address           `json:"owner"`
	Escrow Address           `json:"escrow"`
	Tokens quantity.Quantity `json:"tokens"`
	Height Height            `json:"height,omitempty"`
	Hash   string            `json:"hash,omitempty"`
}
