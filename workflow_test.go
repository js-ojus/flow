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
func TestWorkflows01(t *testing.T) {
	const (
		wflowName = "DATA_PLAT:STOR_REQ"
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

	// List workflows.
	t.Run("List", func(t *testing.T) {
		ws, err := Workflows().List(0, 0)
		if err != nil {
			t.Errorf("error : %v", err)
		}

		for _, w := range ws {
			fmt.Printf("%#v\n", w)
		}
	})

	// Register a few new workflows.
	t.Run("New", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Errorf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		_, err = Workflows().New(tx, wflowName, 3, 1)
		if err != nil {
			t.Errorf("error creating workflow : %v\n", err)
		}

		if err == nil {
			err = tx.Commit()
			if err != nil {
				t.Errorf("error committing transaction : %v\n", err)
			}
		}
	})

	// Retrieve a specified workflow.
	t.Run("GetByID", func(t *testing.T) {
		w, err := Workflows().Get(1)
		if err != nil {
			t.Errorf("error getting workflow '1' : %v\n", err)
		}

		fmt.Printf("%#v\n", w)
	})

	// Verify existence of a specified workflow.
	t.Run("GetByName", func(t *testing.T) {
		w, err := Workflows().GetByName(wflowName)
		if err != nil {
			t.Errorf("error getting workflow : %v\n", err)
		}

		fmt.Printf("%#v\n", w)
	})

}
