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
func TestRoles01(t *testing.T) {
	const (
		roleAdmin = "ADMIN"
		roleUser  = "USER"
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

	// List roles.
	t.Run("List", func(t *testing.T) {
		dts, err := Roles().List(0, 0)
		if err != nil {
			t.Errorf("error : %v", err)
		}

		for _, dt := range dts {
			fmt.Printf("%#v\n", dt)
		}
	})

	// Register a few new roles.
	t.Run("New", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Errorf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		_, err = Roles().New(tx, roleAdmin)
		if err != nil {
			t.Errorf("error creating role '%s' : %v\n", roleAdmin, err)
		}
		_, err = Roles().New(tx, roleUser)
		if err != nil {
			t.Errorf("error creating role '%s' : %v\n", roleUser, err)
		}

		if err == nil {
			err = tx.Commit()
			if err != nil {
				t.Errorf("error committing transaction : %v\n", err)
			}
		}
	})

	// Retrieve a specified role.
	t.Run("GetByID", func(t *testing.T) {
		dt, err := Roles().Get(2)
		if err != nil {
			t.Errorf("error getting role '2' : %v\n", err)
		}

		fmt.Printf("%#v\n", dt)
	})

	// Verify existence of a specified role.
	t.Run("GetByName", func(t *testing.T) {
		dt, err := Roles().GetByName(roleAdmin)
		if err != nil {
			t.Errorf("error getting role '%s' : %v\n", roleAdmin, err)
		}

		fmt.Printf("%#v\n", dt)
	})

	// Rename the given role to the specified new name.
	t.Run("RenameRole", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Errorf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		err = Roles().Rename(tx, 1, "Administrator")
		if err != nil {
			t.Errorf("error renaming role '3' : %v\n", err)
		}

		if err == nil {
			err = tx.Commit()
			if err != nil {
				t.Errorf("error committing transaction : %v\n", err)
			}
		}
	})

	// Rename the given role to the specified old name.
	t.Run("UndoRename", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Errorf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		err = Roles().Rename(tx, 1, roleAdmin)
		if err != nil {
			t.Errorf("error renaming role '3' : %v\n", err)
		}

		if err == nil {
			err = tx.Commit()
			if err != nil {
				t.Errorf("error committing transaction : %v\n", err)
			}
		}
	})
}
