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
	states = []string{"DRAFT", "REQ_PENDING", "APPROVED", "REJECTED"}
)

// Driver test function.
func TestDocStates01(t *testing.T) {
	var dtypeStorReqID DocTypeID
	var dtypeStorRelID DocTypeID

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

		_, err = tx.Exec(`DELETE FROM wf_docstates_master WHERE id > 1`)
		if err != nil {
			t.Fatalf("error running transaction : %v\n", err)
		}

		_, err = tx.Exec(`DELETE FROM wf_doctypes_master`)
		if err != nil {
			t.Fatalf("error running transaction : %v\n", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("error committing transaction : %v\n", err)
		}
	}()

	// Test life cycle.
	t.Run("CRUL", func(t *testing.T) {
		// Insert the required document types.
		tx, err := db.Begin()
		if err != nil {
			t.Fatalf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		dtypeStorReqID, err = DocTypes.New(tx, dtypeStorReq)
		if err != nil {
			t.Fatalf("error creating document type '%s' : %v\n", dtypeStorReq, err)
		}
		dtypeStorRelID, err = DocTypes.New(tx, dtypeStorRel)
		if err != nil {
			t.Fatalf("error creating document type '%s' : %v\n", dtypeStorRel, err)
		}

		// Add document states.
		var ds DocStateID
		for _, name := range states {
			ds, err = DocStates.New(tx, name)
			if err != nil {
				t.Fatalf("error creating document state '%s' : %v\n", name, err)
			}
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("error committing transaction : %v\n", err)
		}

		// Test reading.
		_, err = DocStates.Get(ds)
		if err != nil {
			t.Fatalf("error getting document state '1' : %v\n", err)
		}

		_, err = DocStates.GetByName(states[1])
		if err != nil {
			t.Fatalf("error getting document type:state '%s' : %v\n", states[1], err)
		}

		_, err = DocStates.List(0, 0)
		if err != nil {
			t.Fatalf("error listing document types : %v\n", err)
		}

		// Rename the given document state to the specified new name.
		tx, err = db.Begin()
		if err != nil {
			t.Fatalf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		err = DocStates.Rename(tx, ds, "INITIAL")
		if err != nil {
			t.Fatalf("error renaming document state '1' : %v\n", err)
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

		err = DocStates.Rename(tx, ds, states[len(states)-1])
		if err != nil {
			t.Fatalf("error renaming document state '1' : %v\n", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("error committing transaction : %v\n", err)
		}
	})
}
