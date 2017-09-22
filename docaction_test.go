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
func TestDocActions01(t *testing.T) {
	var (
		actions = []string{"CREATE", "APPROVE", "REJECT", "RET_WITH_COMMENTS", "DISCARD"}
	)

	// Connect to the database.
	driver, connStr := "mysql", "js@/flow"
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

	// List document actions.
	t.Run("List", func(t *testing.T) {
		das, err := DocActions().List(0, 0)
		if err != nil {
			t.Errorf("error : %v", err)
		}

		for _, da := range das {
			fmt.Printf("%#v\n", da)
		}
	})

	// Register a few new document actions.
	t.Run("New", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Errorf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		for _, name := range actions {
			_, err = DocActions().New(tx, name)
			if err != nil {
				t.Errorf("error creating document action '%s' : %v\n", name, err)
			}
		}

		if err == nil {
			err = tx.Commit()
			if err != nil {
				t.Errorf("error committing transaction : %v\n", err)
			}
		}
	})

	// Retrieve a specified document action.
	t.Run("GetByID", func(t *testing.T) {
		da, err := DocActions().Get(1)
		if err != nil {
			t.Errorf("error getting document action '1' : %v\n", err)
		}

		fmt.Printf("%#v\n", da)
	})

	// Verify existence of a specified document action.
	t.Run("GetByName", func(t *testing.T) {
		da, err := DocActions().GetByName(actions[1])
		if err != nil {
			t.Errorf("error getting document action '%s' : %v\n", actions[1], err)
		}

		fmt.Printf("%#v\n", da)
	})

	// Rename the given document action to the specified new name.
	t.Run("RenameAction", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Errorf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		err = DocActions().Rename(tx, 1, "INITIALISE")
		if err != nil {
			t.Errorf("error renaming document action '1' : %v\n", err)
		}

		if err == nil {
			err = tx.Commit()
			if err != nil {
				t.Errorf("error committing transaction : %v\n", err)
			}
		}
	})

	// Rename the given document action to the specified old name.
	t.Run("UndoRename", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Errorf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		err = DocActions().Rename(tx, 1, "CREATE")
		if err != nil {
			t.Errorf("error renaming document action '1' : %v\n", err)
		}

		if err == nil {
			err = tx.Commit()
			if err != nil {
				t.Errorf("error committing transaction : %v\n", err)
			}
		}
	})
}
