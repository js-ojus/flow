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
	wflowName = "DATA_PLAT:STOR_REQ"
)

// Driver test function.
func TestWorkflows01(t *testing.T) {
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

		_, err = tx.Exec(`DELETE FROM wf_workflows`)
		if err != nil {
			t.Fatalf("error running transaction : %v\n", err)
		}

		_, err = tx.Exec(`DELETE FROM wf_docstates_master`)
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

	// Test-local state.
	var dtypeStorReqID DocTypeID
	var dstateID DocStateID
	var wid WorkflowID

	// Register a few new workflows.
	t.Run("Create", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Fatalf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		dtypeStorReqID, err = DocTypes.New(tx, dtypeStorReq)
		if err != nil {
			t.Fatalf("error creating document type '%s' : %v\n", dtypeStorReq, err)
		}
		for _, name := range storReqStates {
			dstateID, err = DocStates.New(tx, dtypeStorReqID, name)
			if err != nil {
				t.Fatalf("error creating document type:state '%d:%s' : %v\n", dtypeStorReqID, name, err)
			}
		}

		wid, err = Workflows.New(tx, wflowName, dtypeStorReqID, dstateID)
		if err != nil {
			t.Fatalf("error creating workflow : %v\n", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("error committing transaction : %v\n", err)
		}
	})

	// Test reading.
	t.Run("Read", func(t *testing.T) {
		_, err = Workflows.Get(wid)
		if err != nil {
			t.Fatalf("error getting workflow : %v\n", err)
		}

		_, err = Workflows.GetByName(wflowName)
		if err != nil {
			t.Fatalf("error getting workflow : %v\n", err)
		}

		_, err = Workflows.List(0, 0)
		if err != nil {
			t.Fatalf("error : %v", err)
		}
	})

	// Test renaming.
	t.Run("Rename", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Fatalf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		err = Workflows.Rename(tx, wid, "TEST_WFLOW")
		if err != nil {
			t.Fatalf("error renaming workflow : %v\n", err)
		}
		err = Workflows.Rename(tx, wid, wflowName)
		if err != nil {
			t.Fatalf("error renaming workflow : %v\n", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("error committing transaction : %v\n", err)
		}
	})

	// Test activation and inactivation.
	t.Run("Active", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Fatalf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		err = Workflows.SetActive(tx, wid, false)
		if err != nil {
			t.Fatalf("error inactivating workflow : %v\n", err)
		}
		err = Workflows.SetActive(tx, wid, true)
		if err != nil {
			t.Fatalf("error activating workflow : %v\n", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("error committing transaction : %v\n", err)
		}
	})
}
