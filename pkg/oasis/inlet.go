// Given some DB, we want to aggregate queries and bulk write them in groups of
// transactions at some arbitrary limit. We will define a global instance so we
// can call into this from any application without large modification of the
// source.

// -----------------------------------------------------------------------------

package oasis

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// -----------------------------------------------------------------------------
// Types

const (
	TO_DATABASE = 1
	TO_DISK     = 2
)

type WriteMode = uint64

type Batch = []Query
type Query struct {
	Query string        `json:"query"`
	Args  []interface{} `json:"args"`
}

type Inlet struct {
	channel      chan<- Batch      // Channel to a goroutine running DB transactions in the background.
	createdAt    time.Time         // Used to associate logs with a single session.
	db           *sql.DB           // Underlying Database.
	errHandler   func(error)       // Function to handle errors in the background goroutine.
	mu           sync.Mutex        // Lock access to this structure in concurrent functions.
	queries      Batch             // Current batch of queries.
	syncAt       int               // Height to create a new batch at.
	writeHandler func(Batch) error // Function to handle query writes
	writeMode    WriteMode         // If an error occurs writing to the DB, enter failed state, and write to disk instead.
}

// -----------------------------------------------------------------------------
// Create a singleton instance and an initializer that makes sure to only run
// initialization once.

var (
	inlet *Inlet
	one   sync.Once
)

func InitInlet(db *sql.DB, syncAt int, errHandler func(error), writeHandler func(Batch) error) {
	one.Do(func() {
		queries := make(Batch, 0, syncAt)
		channel := make(chan Batch, 8)

		// Note: We are directly writing to a global pointer here, but we know
		// thanks to one.Do that this can only ever happen one time, by one
		// caller, so we know this is relatively safe.
		inlet = &Inlet{
			channel:      channel,
			createdAt:    time.Now(),
			db:           db,
			errHandler:   errHandler,
			mu:           sync.Mutex{},
			queries:      queries,
			syncAt:       syncAt,
			writeHandler: writeHandler,
			writeMode:    TO_DATABASE,
		}

		// Process Batches in the background with a loop over the channel. We
		// do this so we can guarantee ordering to processing each batch.
		go func() {
			for {
				select {
				case batch := <-channel:
					log.Println("Batch Starting")
					switch {
					case atomic.LoadUint64(&inlet.writeMode) == TO_DISK:
						log.Println("Batching to Disk")
						batchToDisk(batch)
					default:
						log.Println("Batching to DB")
						batchToDB(batch)
					}
				}
			}
		}()
	})
}

// -----------------------------------------------------------------------------
// Public API

func PushQuery(query string, args ...interface{}) {
	inlet.mu.Lock()
	defer inlet.mu.Unlock()

	// Append to batch, then forward batch if full.
	inlet.queries = append(inlet.queries, Query{query, args})
	log.Printf("Batch Size: %v\n", len(inlet.queries))
	if len(inlet.queries) == inlet.syncAt {
		log.Printf("Pushing Batch of %d Queries\n", len(inlet.queries))
		inlet.channel <- inlet.queries
		inlet.queries = make(Batch, 0, inlet.syncAt)
	}
}

// -----------------------------------------------------------------------------
// Private Functions

func batchToDB(queries Batch) {
	// tx, err := inlet.db.Begin()
	// if err != nil {
	// 	atomic.StoreUint64(&inlet.writeMode, TO_DISK)
	// 	log.Println("Transaction Begin Failed")
	// 	inlet.errHandler(err)
	// 	batchToDisk(queries)
	// 	return
	// }

	if err := inlet.writeHandler(queries); err != nil {
		atomic.StoreUint64(&inlet.writeMode, TO_DISK)
		log.Println("Batch Write Failed")
		inlet.errHandler(err)
		batchToDisk(queries)
	}

	// for _, q := range queries {
	// 	if _, err := tx.Exec(q.Query, q.Args...); err != nil {
	// 		atomic.StoreUint64(&inlet.writeMode, TO_DISK)
	// 		log.Println("Query Failed")
	// 		_ = tx.Rollback()
	// 		inlet.errHandler(err)
	// 		batchToDisk(queries)
	// 		return
	// 	}
	// }

	// log.Println("Persisting Transaction")
	// if err := tx.Commit(); err != nil {
	// 	atomic.StoreUint64(&inlet.writeMode, TO_DISK)
	// 	log.Println("Transaction Persist Failed")
	// 	inlet.errHandler(err)
	// 	batchToDisk(queries)
	// }
}

func getReference() *Inlet {
	return inlet
}

// There are cases where writing a batch may fail. There may be a networking
// issue, the database might be full, or a rare transaction collision. Instead of
// just exploding, we'll start writing batches to disk when this happens and name
// them sequentially so we can recover manually.

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func batchToDisk(queries Batch) {
	nanoseconds := time.Now().UnixNano()
	logFileName := fmt.Sprintf("inlet_%d_%d.json", inlet.createdAt.Unix(), nanoseconds)

	// If Anything here Fails, Panic and Crash
	file, err := os.Create(logFileName)
	defer file.Close()
	check(err)

	// Write queries as JSON.
	encoder := json.NewEncoder(file)
	for _, q := range queries {
		err := encoder.Encode(q)
		check(err)
	}
}
