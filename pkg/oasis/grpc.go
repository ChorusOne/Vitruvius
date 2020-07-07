// This file implements this packages API inteface backed by  Oasis gRPC API.
//
// NOTE:
// This file re-uses the Oasis gRPC client adaptor, as it is fully configured
// to provide their CBOR serialization configuration. This is to make sure we
// are using the same serialization as the actual Oasis codebase is currently
// using to store data.

package oasis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/oasisprotocol/oasis-core/go/common/cbor"
	"github.com/oasisprotocol/oasis-core/go/common/crypto/hash"
	"github.com/oasisprotocol/oasis-core/go/common/crypto/signature"
	"github.com/oasisprotocol/oasis-core/go/consensus/api/transaction"
	"google.golang.org/grpc"

	grpcOasis "github.com/oasisprotocol/oasis-core/go/common/grpc"
	consensus "github.com/oasisprotocol/oasis-core/go/consensus/api"
	staking "github.com/oasisprotocol/oasis-core/go/staking/api"
	tmtypes "github.com/tendermint/tendermint/types"
)

// Types
// ------------------------------------------------------------------------------

// Oasis wraps the state required to maintain a connection with the real Oasis
// gRPC API.
type Oasis struct {
	conn  *grpc.ClientConn // Stores a live connection to an Oasis Chain.
	State *chainState      // Stores the current chain state, stays in sync with the blockchain.
}

// Enforce Interface
var _ API = &Oasis{}

// chainState stores all the data that changes as the extractor sync with the
// oasis chain. It's a separate struct from `Oasis` so that we can swap the
// entire state out atomically.
type chainState struct {
	Block    *Block
	Height   Height
	Snapshot *staking.Genesis
}

// Utility Functions
// ------------------------------------------------------------------------------

func decodeBlockAsTendermint(block *consensus.Block) Block {
	var tendermintBlock tmtypes.Block
	if err := cbor.Unmarshal(block.Meta, &tendermintBlock); err != nil {
		log.Fatalln("Failed to decode Tendermint Metadata")
	}

	return Block{
		Height:  block.Height,
		ChainID: tendermintBlock.Header.ChainID,
		Time:    tendermintBlock.Header.Time,
	}
}

// syncChain attempts to copy current chain state into the local state.
func (oasis *Oasis) syncChain(block *Block) error {
	ctx := context.Background()
	api := staking.NewStakingClient(oasis.conn)
	gen, err := api.StateToGenesis(ctx, block.Height)
	if err != nil {
		log.Fatalf("syncChain: Failed to sync, %v\n", err)
		return err
	}

	// Update New State Atomically
	currState := (*unsafe.Pointer)(unsafe.Pointer(&oasis.State))
	nextState := unsafe.Pointer(&chainState{
		Block:    block,
		Height:   block.Height,
		Snapshot: gen,
	})

	atomic.StorePointer(currState, nextState)
	return nil
}

// freezeChain will atomically clone the current state and fix it at the
// current height.
func (oasis *Oasis) freezeChain() *Oasis {
	currentState := (*unsafe.Pointer)(unsafe.Pointer(&oasis.State))
	return &Oasis{
		conn:  oasis.conn,
		State: (*chainState)(atomic.LoadPointer(currentState)),
	}
}

// API implementation for Oasis
// ------------------------------------------------------------------------------

// NewOasis tries to open a gRPC connection with an existing oasis-core unix
// socket. If it finds one, it spawns a goroutine that keeps track of the
// current height of the chain.
func NewOasis(address string) (*Oasis, error) {
	// If the argument is a file, assume It's a UNIX socket.
	if _, err := os.Stat(address); err == nil {
		address = "unix:" + address
	}

	// Dial RPC, Insecure By Default
	conn, err := grpcOasis.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	// Initial State -- we haven't observed any blocks yet, so everything in
	// state is left unset (nil) by default.
	oasis := &Oasis{conn: conn}

	// In order to allow the API to behave as if it is always currently
	// querying the top block, we'll sync the object in the background with the
	// tip of the chain.
	waitChan := make(chan int)
	waitOnce := sync.Once{}
	go func() {
		if channel, err := oasis.WatchBlocks(); err == nil {
			for {
				select {
				case block := <-channel:
					oasis.syncChain(&block)
					waitOnce.Do(func() { waitChan <- 0 })
				}
			}
		}

		log.Fatalln("NewOasis: Could not open WatchBlocks channel")
	}()

	// Wait for at least one block to sync before we allow the API to be
	// considered ready.
	select {
	case _ = <-waitChan:
		close(waitChan)
	}

	return oasis, nil
}

// Utilities
// -----------------------------------------------------------------------------

func (oasis *Oasis) AtHeight(height Height) API {
	ctx := context.Background()
	api := consensus.NewConsensusClient(oasis.conn)
	block, err := api.GetBlock(ctx, height)
	if err != nil {
		log.Fatalf("Attempted to process non-existing Block, %v\n", err)
		return nil
	}

	tendermintBlock := decodeBlockAsTendermint(block)
	newAPI := Oasis{conn: oasis.conn}
	newAPI.syncChain(&tendermintBlock)
	return oasis
}

// DecodeKey is a small helper to decode Oasis' internal encoded keys to
// something we can work with locally.
func (oasis *Oasis) DecodeKey(id string) (Address, error) {
	var address Address
	err := address.UnmarshalText([]byte(id))
	return address, err
}

// Now fixes the current API state at a fixed height, this prevents drift if
// the chain produces a new block while the API is being used.
func (oasis *Oasis) Now() API {
	return &Oasis{
		conn:  oasis.conn,
		State: oasis.State,
	}
}

// API
// -----------------------------------------------------------------------------

// Account extracts a full snapshot state of an account from the genesis block
// at a given height.
func (oasis *Oasis) Account(id Address) (*Account, error) {
	self := oasis.freezeChain()

	// Attempt to retrieve Account from Genesis State.
	account, ok := self.State.Snapshot.Ledger[id]
	if !ok {
		return nil, fmt.Errorf("No Account with ID: %v", id)
	}

	// Convert to internal representations
	stakedBalance := SharePool{
		Balance:     account.Escrow.Active.Balance.String(),
		TotalShares: account.Escrow.Active.TotalShares.String(),
	}

	debondingBalance := SharePool{
		Balance:     account.Escrow.Debonding.Balance.String(),
		TotalShares: account.Escrow.Debonding.TotalShares.String(),
	}

	// Delegations
	delegations := self.AccountDelegations(id)

	return &Account{
		Address:          id,
		Balance:          account.General.Balance.String(),
		StakedBalance:    &stakedBalance,
		DebondingBalance: &debondingBalance,
		Height:           self.State.Height,
		Delegations:      delegations,
		Meta: AccountMeta{
			IsValidator: false,
			IsDelegator: false,
		},
	}, nil
}

func (oasis *Oasis) AccountDelegations(id Address) []Delegation {
	self := oasis.freezeChain()
	account, err := self.State.Snapshot.Delegations[id]
	if !err {
		return nil
	}

	// Convert all delegations for this account into our local types.
	delegations := []Delegation{}
	for delegatee, delegation := range account {
		delegations = append(delegations, Delegation{
			Delegator: id,
			Validator: delegatee,
			Amount:    delegation.Shares,
		})
	}

	return delegations
}

func (oasis *Oasis) Accounts() []Address {
	self := oasis.freezeChain()
	ctx := context.Background()
	api := staking.NewStakingClient(self.conn)
	addresses, err := api.Addresses(ctx, self.State.Height)

	if err != nil {
		fmt.Printf("%+v", err)
		return nil
	}

	return addresses
}

func (oasis *Oasis) Delegations() []Delegation {
	self := oasis.freezeChain()

	// For Each Account...
	delegations := []Delegation{}
	for account := range self.State.Snapshot.Delegations {
		// ... and each Delegation that Account has done. Append to list.
		for delegatee := range self.State.Snapshot.Delegations[account] {
			delegations = append(delegations, Delegation{
				Delegator: account,
				Validator: delegatee,
				Amount:    self.State.Snapshot.Delegations[account][delegatee].Shares,
			})
		}
	}
	return delegations
}

func (oasis *Oasis) GetBlock() (Block, Height) {
	self := oasis.freezeChain()
	return *self.State.Block, self.State.Height
}

func (oasis *Oasis) GetEvents() []StakingEvent {
	self := oasis.freezeChain()
	ctx := context.Background()
	api := staking.NewStakingClient(self.conn)
	events, err := api.GetEvents(ctx, self.State.Height)
	if err != nil {
		fmt.Printf("GetEvents: failed to fetch txs, %v", err)
		return nil
	}

	stakingEvents := []StakingEvent{}
	for _, event := range events {
		stakingEvent := convertEvent(&event)
		stakingEvents = append(stakingEvents, stakingEvent)
	}

	return stakingEvents
}

// GetGenesisState will fetch a full snapshot of the chains state in a format
// Oasis uses to bootstrap network ugprades. Storing this should allow us to
// easily pull historical data without going through gRPC.
func (oasis *Oasis) GetGenesisState() (*Genesis, error) {
	self := oasis.freezeChain()
	ctx := context.Background()
	api := staking.NewStakingClient(self.conn)

	var err error
	var genesis *staking.Genesis
	var encoded []byte

	if genesis, err = api.StateToGenesis(ctx, self.State.Height); err != nil {
		return nil, fmt.Errorf("Failed to fetch Genesis for height %v: %w", self.State.Height, err)
	}

	if encoded, err = json.Marshal(&genesis); err != nil {
		return nil, fmt.Errorf("Failed to encode Genesis for height %v: %w", self.State.Height, err)
	}

	return &Genesis{encoded}, nil
}

// GetTransactions pulls decoded Oasis transactions from gRPC.
func (oasis *Oasis) GetTransactions() []Transaction {
	self := oasis.freezeChain()
	ctx := context.Background()
	api := consensus.NewConsensusClient(self.conn)
	txs, err := api.GetTransactions(ctx, self.State.Height)
	if err != nil {
		fmt.Printf("GetTransactions: failed to fetch txs, %v", err)
		return nil
	}

	decodedTxs := []Transaction{}
	for _, tx := range txs {
		var signedTx signature.Signed
		var decodeTx transaction.Transaction
		if err := cbor.Unmarshal(tx, &signedTx); err != nil {
			log.Printf("GetTransactions: Unmarshal Signed failed, %v", err)
			return nil
		}

		if err := cbor.Unmarshal(signedTx.Blob, &decodeTx); err != nil {
			log.Printf("GetTransactions: Unmarshal Blob failed, %v", err)
			return nil
		}

		// Create Base Tx Structure
		tx := Transaction{
			Method:   string(decodeTx.Method),
			Sender:   staking.NewAddress(signedTx.Signature.PublicKey),
			Hash:     hash.NewFromBytes(tx).String(),
			Fee:      decodeTx.Fee.Amount,
			Gas:      uint64(decodeTx.Fee.Gas),
			GasPrice: *decodeTx.Fee.GasPrice(),
		}

		// Convert various encoded blobs into JSON we can store in the database
		// for Anthem. This code is quite meaty, I'd love to split it out into
		// another module.
		// TODO: Is there a code generator we can use for this?
		switch decodeTx.Method {
		case staking.MethodTransfer:
			var cborPayload staking.Transfer
			if err := cbor.Unmarshal(decodeTx.Body, &cborPayload); err != nil {
				log.Printf("GetTransactions: Unmarshal MethodTransfer Failed, %v", err)
				return nil
			}
			tx.Payload = &TransferTx{
				From:   staking.NewAddress(signedTx.Signature.PublicKey),
				To:     cborPayload.To,
				Tokens: cborPayload.Tokens,
			}

		case staking.MethodAddEscrow:
			var cborPayload staking.Escrow
			if err := cbor.Unmarshal(decodeTx.Body, &cborPayload); err != nil {
				log.Printf("GetTransactions: Unmarshal MethodAddEscrow Failed, %v", err)
				return nil
			}
			tx.Payload = &AddEscrowTx{
				To:     cborPayload.Account,
				Tokens: cborPayload.Tokens,
			}

		case staking.MethodBurn:
			var cborPayload staking.Burn
			if err := cbor.Unmarshal(decodeTx.Body, &cborPayload); err != nil {
				log.Printf("GetTransactions: Unmarshal MethodBurn Failed, %v", err)
				return nil
			}
			tx.Payload = &BurnTx{
				From:   staking.NewAddress(signedTx.Signature.PublicKey),
				Tokens: cborPayload.Tokens,
			}

		case staking.MethodReclaimEscrow:
			var cborPayload staking.ReclaimEscrow
			if err := cbor.Unmarshal(decodeTx.Body, &cborPayload); err != nil {
				log.Printf("GetTransactions: Unmarshal MethodReclaimEscrow Failed, %v", err)
				return nil
			}
			tx.Payload = &ReclaimEscrowTx{
				From:   cborPayload.Account,
				Shares: cborPayload.Shares,
			}

		default:
			tx.Payload = &UnknownTx{
				Payload: decodeTx.Body,
			}
		}

		decodedTxs = append(decodedTxs, tx)
	}

	return decodedTxs
}

func (oasis *Oasis) GetValidatorCommission(id Address) (*Amount, *Amount, error) {
	self := oasis.freezeChain()
	ctx := context.Background()
	api := consensus.NewConsensusClient(self.conn)

	// Attempt to retrieve Account from Genesis State.
	account, ok := self.State.Snapshot.Ledger[id]
	if !ok {
		return nil, nil, fmt.Errorf("No Account with ID: %v", id)
	}

	// Get Current Commission Numerator
	epochTime, _ := api.GetEpoch(ctx, self.State.Height)
	rate := account.Escrow.CommissionSchedule.CurrentRate(epochTime)
	return rate, staking.CommissionRateDenominator, nil
}

func (oasis *Oasis) Pool() (*Pool, error) {
	self := oasis.freezeChain()
	ctx := context.Background()
	api := staking.NewStakingClient(self.conn)
	pool, err := api.CommonPool(ctx, self.State.Height)
	if err != nil {
		return nil, err
	}
	return pool, nil
}

// Watchers
// -----------------------------------------------------------------------------

// WatchBlocks wraps the WatchBlocks provided by the Oasis API, and decodes the
// underlying Tendermint header transparently.
func (oasis *Oasis) WatchBlocks() (chan Block, error) {
	self := oasis.freezeChain()
	ctx := context.Background()
	api := consensus.NewConsensusClient(self.conn)
	channel, _, err := api.WatchBlocks(ctx)
	if err != nil {
		return nil, err
	}

	proxyChannel := make(chan Block)

	// Spawn Proxy Goroutine
	go func() {
		for {
			select {
			case block := <-channel:
				tendermintBlock := decodeBlockAsTendermint(block)
				proxyChannel <- tendermintBlock
			}
		}
	}()

	return proxyChannel, nil
}

func (oasis *Oasis) WatchStakingEvents() (chan StakingEvent, error) {
	self := oasis.freezeChain()
	ctx := context.Background()
	api := staking.NewStakingClient(self.conn)

	var eventChannel <-chan *staking.Event
	var err error

	// Watch All Channels
	if eventChannel, _, err = api.WatchEvents(ctx); err != nil {
		return nil, fmt.Errorf("WatchStakingEvents: failed WatchEvents, %w", err)
	}

	proxyChannel := make(chan StakingEvent)

	go func() {
		for {
			select {
			case msg := <-eventChannel:
				proxyChannel <- convertEvent(msg)
			}
		}
	}()

	return proxyChannel, nil
}

func convertEvent(event *staking.Event) StakingEvent {
	switch {
	case event.Transfer != nil:
		return StakingEvent{
			Transfer: &TransferEvent{
				Hash:   event.TxHash.String(),
				From:   event.Transfer.From,
				To:     event.Transfer.To,
				Tokens: event.Transfer.Tokens,
			},
		}

	case event.Burn != nil:
		return StakingEvent{
			Burn: &BurnEvent{
				Hash:   event.TxHash.String(),
				Owner:  event.Burn.Owner,
				Tokens: event.Burn.Tokens,
			},
		}

	case event.Escrow != nil:
		addEvent := (*AddEscrowEvent)(nil)
		takeEvent := (*TakeEscrowEvent)(nil)
		reclaimEvent := (*ReclaimEscrowEvent)(nil)

		switch {
		case event.Escrow.Add != nil:
			addEvent = &AddEscrowEvent{
				Owner:  event.Escrow.Add.Owner,
				Escrow: event.Escrow.Add.Escrow,
				Tokens: event.Escrow.Add.Tokens,
				Hash:   event.TxHash.String(),
			}
		case event.Escrow.Take != nil:
			takeEvent = &TakeEscrowEvent{
				Owner:  event.Escrow.Take.Owner,
				Tokens: event.Escrow.Take.Tokens,
				Hash:   event.TxHash.String(),
			}
		case event.Escrow.Reclaim != nil:
			reclaimEvent = &ReclaimEscrowEvent{
				Owner:  event.Escrow.Reclaim.Owner,
				Escrow: event.Escrow.Reclaim.Escrow,
				Tokens: event.Escrow.Reclaim.Tokens,
				Hash:   event.TxHash.String(),
			}
		}

		return StakingEvent{
			Escrow: &EscrowEvent{
				Add:     addEvent,
				Take:    takeEvent,
				Reclaim: reclaimEvent,
			},
		}
	}

	return StakingEvent{
		Unknown: &UnknownEvent{},
	}
}
