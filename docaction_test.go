// (c) Copyright 2015-2017 JONNALAGADDA Srinivas
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package flow

import (
	"database/sql"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

var (
	actions = []string{"CREATE", "APPROVE", "REJECT", "RET_WITH_COMMENTS", "DISCARD"}
)

// Driver test function.
func TestDocActions01(t *testing.T) {
	// Connect to the database.
	driver, connStr := "mysql", "travis@/flow"
	db, err := sql.Open(driver, connStr)
	if err != nil {
		t.Fatalf("could not connect to database : %v\n", err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		t.Fatalf("could not ping the database : %v\n", err)
	}
	RegisterDB(db)

	// Tear down.
	defer func() {
		tx, err := db.Begin()
		if err != nil {
			t.Fatalf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		_, err = tx.Exec(`DELETE FROM wf_docactions_master`)
		if err != nil {
			t.Fatalf("error running transaction : %v\n", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("error committing transaction : %v\n", err)
		}
	}()

	// Test CRUL operations.
	t.Run("CRUL", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Fatalf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		var da DocActionID
		for _, name := range actions {
			da, err = DocActions().New(tx, name)
			if err != nil {
				t.Fatalf("error creating document action '%s' : %v\n", name, err)
			}
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("error committing transaction : %v\n", err)
		}

		// Test read operations.
		_, err = DocActions().Get(da)
		if err != nil {
			t.Fatalf("error getting document action : %v\n", err)
		}

		_, err = DocActions().GetByName(actions[1])
		if err != nil {
			t.Fatalf("error getting document action '%s' : %v\n", actions[1], err)
		}

		_, err = DocActions().List(0, 0)
		if err != nil {
			t.Fatalf("error : %v", err)
		}

		// Test renaming.
		tx, err = db.Begin()
		if err != nil {
			t.Fatalf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		err = DocActions().Rename(tx, da, "INITIALISE")
		if err != nil {
			t.Fatalf("error renaming document action : %v\n", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("error committing transaction : %v\n", err)
		}

		tx, err = db.Begin()
		if err != nil {
			t.Fatalf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		err = DocActions().Rename(tx, da, actions[len(actions)-1])
		if err != nil {
			t.Fatalf("error renaming document action : %v\n", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("error committing transaction : %v\n", err)
		}
	})
}
