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

const (
	dtypeStorReq = "DATA_PLAT:STOR_REQ"
	dtypeStorRel = "DATA_PLAT:STOR_REL"
)

// Driver test function.
func TestDocTypes01(t *testing.T) {
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
		// Test creation.
		tx, err := db.Begin()
		if err != nil {
			t.Fatalf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		dtypeStorReqID, err := DocTypes().New(tx, dtypeStorReq)
		if err != nil {
			t.Fatalf("error creating document type '%s' : %v\n", dtypeStorReq, err)
		}
		_, err = DocTypes().New(tx, dtypeStorRel)
		if err != nil {
			t.Fatalf("error creating document type '%s' : %v\n", dtypeStorRel, err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("error committing transaction : %v\n", err)
		}

		// Test reading.
		_, err = DocTypes().Get(dtypeStorReqID)
		if err != nil {
			t.Fatalf("error getting document type : %v\n", err)
		}

		_, err = DocTypes().GetByName(dtypeStorRel)
		if err != nil {
			t.Fatalf("error getting document type '%s' : %v\n", dtypeStorRel, err)
		}

		_, err = DocTypes().List(0, 0)
		if err != nil {
			t.Fatalf("error : %v", err)
		}

		_, err = DocStates().List(0, 0)
		if err != nil {
			t.Fatalf("error : %v", err)
		}

		// Test renaming.
		tx, err = db.Begin()
		if err != nil {
			t.Fatalf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		err = DocTypes().Rename(tx, dtypeStorReqID, "DATA_PLAT:STOR_DEL")
		if err != nil {
			t.Fatalf("error renaming document type : %v\n", err)
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

		err = DocTypes().Rename(tx, dtypeStorReqID, "DATA_PLAT:STOR_REQ")
		if err != nil {
			t.Fatalf("error renaming document type : %v\n", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("error committing transaction : %v\n", err)
		}
	})
}
