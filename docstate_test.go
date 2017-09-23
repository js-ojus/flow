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
	"fmt"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

// Driver test function.
func TestDocStates01(t *testing.T) {
	const (
		dtypeStorReqID = 3
		dtypeStorRelID = 4
	)
	var (
		storReqStates = []string{"REQ_PENDING", "APPROVED", "REJECTED", "RET_WITH_COMMENTS"}
		storRelStates = []string{"REQ_PENDING", "APPROVED", "REJECTED", "RET_WITH_COMMENTS"}
	)

	// Connect to the database.
	driver, connStr := "mysql", "travis@/flow"
	db, err := sql.Open(driver, connStr)
	if err != nil {
		t.Errorf("could not connect to database : %v\n", err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		t.Errorf("could not ping the database : %v\n", err)
	}
	RegisterDB(db)

	// List document states.
	t.Run("List", func(t *testing.T) {
		dss, err := DocStates().List(0, 0)
		if err != nil {
			t.Errorf("error : %v", err)
		}

		for _, ds := range dss {
			fmt.Printf("%#v\n", ds)
		}
	})

	// Register a few new document states.
	t.Run("New", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Errorf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		for _, name := range storReqStates {
			_, err = DocStates().New(tx, dtypeStorReqID, name)
			if err != nil {
				t.Errorf("error creating document type:state '%d:%s' : %v\n", dtypeStorReqID, name, err)
			}
		}
		for _, name := range storRelStates {
			_, err = DocStates().New(tx, dtypeStorRelID, name)
			if err != nil {
				t.Errorf("error creating document type:state '%d:%s' : %v\n", dtypeStorRelID, name, err)
			}
		}

		if err == nil {
			err = tx.Commit()
			if err != nil {
				t.Errorf("error committing transaction : %v\n", err)
			}
		}
	})

	// Retrieve a specified document state.
	t.Run("GetByID", func(t *testing.T) {
		ds, err := DocStates().Get(1)
		if err != nil {
			t.Errorf("error getting document state '1' : %v\n", err)
		}

		fmt.Printf("%#v\n", ds)
	})

	// Verify existence of a specified document state.
	t.Run("GetByName", func(t *testing.T) {
		ds, err := DocStates().GetByName(dtypeStorReqID, storReqStates[1])
		if err != nil {
			t.Errorf("error getting document type:state '%d:%s' : %v\n", dtypeStorReqID, storReqStates[1], err)
		}

		fmt.Printf("%#v\n", ds)
	})

	// Rename the given document state to the specified new name.
	t.Run("RenameState", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Errorf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		err = DocStates().Rename(tx, 1, "INITIAL")
		if err != nil {
			t.Errorf("error renaming document state '1' : %v\n", err)
		}

		if err == nil {
			err = tx.Commit()
			if err != nil {
				t.Errorf("error committing transaction : %v\n", err)
			}
		}
	})

	// Rename the given document state to the specified old name.
	t.Run("UndoRename", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Errorf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		err = DocStates().Rename(tx, 1, "REQ_PENDING")
		if err != nil {
			t.Errorf("error renaming document state '1' : %v\n", err)
		}

		if err == nil {
			err = tx.Commit()
			if err != nil {
				t.Errorf("error committing transaction : %v\n", err)
			}
		}
	})
}
