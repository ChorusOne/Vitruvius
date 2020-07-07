// This file declares a BlockIterator interface, a primitive for executing some
// set of logic over a blockchain.

package extractor

import (
	"github.com/ChorusOne/Hippias/pkg/oasis"
)

// ExtractorSnapshot groups together anything needed for the extractor
// to work. BlockIterators will receive these.
type StateSnapshot struct {
	Block        oasis.Block
	Events       []oasis.StakingEvent
	Transactions []oasis.Transaction
}

// BlockIterator defines the interface for a type to implement
// to become a block iterator.
type BlockIterator interface {
	Process(StateSnapshot)
}
