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
func TestDocTypes01(t *testing.T) {
	const (
		dtypeStorReq = "DATA_PLAT:STOR_REQ"
		dtypeStorRel = "DATA_PLAT:STOR_REL"
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

	// List document types.
	t.Run("List", func(t *testing.T) {
		dts, err := DocTypes().List(0, 0)
		if err != nil {
			t.Errorf("error : %v", err)
		}

		for _, dt := range dts {
			fmt.Printf("%#v\n", dt)
		}
	})

	// Register a few new document types.
	t.Run("New", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Errorf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		_, err = DocTypes().New(tx, dtypeStorReq)
		if err != nil {
			t.Errorf("error creating document type '%s' : %v\n", dtypeStorReq, err)
		}
		_, err = DocTypes().New(tx, dtypeStorRel)
		if err != nil {
			t.Errorf("error creating document type '%s' : %v\n", dtypeStorRel, err)
		}

		if err == nil {
			err = tx.Commit()
			if err != nil {
				t.Errorf("error committing transaction : %v\n", err)
			}
		}
	})

	// Retrieve a specified document type.
	t.Run("GetByID", func(t *testing.T) {
		dt, err := DocTypes().Get(3)
		if err != nil {
			t.Errorf("error getting document type '3' : %v\n", err)
		}

		fmt.Printf("%#v\n", dt)
	})

	// Verify existence of a specified document type.
	t.Run("GetByName", func(t *testing.T) {
		dt, err := DocTypes().GetByName(dtypeStorRel)
		if err != nil {
			t.Errorf("error getting document type '%s' : %v\n", dtypeStorRel, err)
		}

		fmt.Printf("%#v\n", dt)
	})

	// Rename the given document type to the specified new name.
	t.Run("RenameType", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Errorf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		err = DocTypes().Rename(tx, 3, "DATA_PLAT:STOR_DEL")
		if err != nil {
			t.Errorf("error renaming document type '3' : %v\n", err)
		}

		if err == nil {
			err = tx.Commit()
			if err != nil {
				t.Errorf("error committing transaction : %v\n", err)
			}
		}
	})

	// Rename the given document type to the specified old name.
	t.Run("UndoRename", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Errorf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		err = DocTypes().Rename(tx, 3, "DATA_PLAT:STOR_REQ")
		if err != nil {
			t.Errorf("error renaming document type '3' : %v\n", err)
		}

		if err == nil {
			err = tx.Commit()
			if err != nil {
				t.Errorf("error committing transaction : %v\n", err)
			}
		}
	})
}
