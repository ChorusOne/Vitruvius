// State wraps shared data, such as database connections and
// Oasis API connections, so that components of the system
// can avoid global sharing.

package types

import (
	"database/sql"
	"io/ioutil"

	"github.com/ChorusOne/Hippias/pkg/oasis"
	"github.com/gchaincl/dotsql"
)

// State contains all the shared external resources that endpoints are using.
// Essentially API injection.
type State struct {
	Api oasis.API
	Db  *sql.DB
	Dot *dotsql.DotSql
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

// NewState is just a public constructor. It also does some filesystem heavy
// lifting to construct the merged dotsql instance of all .sql files under
// sql/queries
func NewState(api oasis.API, db *sql.DB) State {
	files, err := ioutil.ReadDir("sql/queries")
	check(err)

	// Create and Merge dot instances for every SQL file found.
	var dot *dotsql.DotSql
	for _, file := range files {
		dotFile, err := dotsql.LoadFromFile("sql/queries/" + file.Name())
		check(err)
		if dot == nil {
			dot = dotFile
		} else {
			dot = dotsql.Merge(dot, dotFile)
		}
	}

	return State{api, db, dot}
}
