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
func TestGroups01(t *testing.T) {
	const genGrp = "research"

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

	// List groups.
	t.Run("List", func(t *testing.T) {
		gs, err := Groups().List(0, 0)
		if err != nil {
			t.Errorf("error : %v", err)
		}

		for _, g := range gs {
			fmt.Printf("%#v\n", g)
		}
	})

	// Create a singleton group.
	t.Run("CreateSingleton", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Errorf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		_, err = Groups().NewSingleton(tx, 1)
		if err != nil {
			t.Errorf("error adding singleton group for user '%d' : %v\n", 1, err)
		}

		if err == nil {
			err = tx.Commit()
			if err != nil {
				t.Errorf("error committing transaction : %v\n", err)
			}
		}
	})

	// Create a general group.
	t.Run("CreateGeneral", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Errorf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		_, err = Groups().New(tx, genGrp, "G")
		if err != nil {
			t.Errorf("error adding group '%s' : %v\n", genGrp, err)
		}

		if err == nil {
			err = tx.Commit()
			if err != nil {
				t.Errorf("error committing transaction : %v\n", err)
			}
		}
	})

	// Retrieve a specified group.
	t.Run("GetByID", func(t *testing.T) {
		g, err := Groups().Get(3)
		if err != nil {
			t.Errorf("error getting group '1' : %v\n", err)
		}

		fmt.Printf("%#v\n", g)
	})

	// Delete a singleton group.
	t.Run("DeleteSingleton", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Errorf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		err = Groups().Delete(tx, 3)
		if err == nil {
			t.Errorf("error deleting singleton group : %d\n", 3)
		}

		if err == nil {
			err = tx.Commit()
			if err != nil {
				t.Errorf("error committing transaction : %v\n", err)
			}
		}
	})

	// Delete a singleton group.
	t.Run("DeleteGeneral1", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Errorf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		err = Groups().Delete(tx, 3)
		if err == nil {
			t.Errorf("error deleting general group '%d' : %v\n", 3, err)
		}

		if err == nil {
			err = tx.Commit()
			if err != nil {
				t.Errorf("error committing transaction : %v\n", err)
			}
		}
	})

	// Delete a singleton group.
	t.Run("DeleteGeneral2", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Errorf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		err = Groups().Delete(tx, 5)
		if err != nil {
			t.Errorf("error deleting general group '%d' : %v\n", 5, err)
		}

		if err == nil {
			err = tx.Commit()
			if err != nil {
				t.Errorf("error committing transaction : %v\n", err)
			}
		}
	})

	// Query if a user is part of the given group.
	t.Run("HasUser", func(t *testing.T) {
		ok, err := Groups().HasUser(3, 1)
		if err != nil {
			t.Errorf("error querying group users : %v\n", err)
		}

		if !ok {
			t.Errorf("singleton user must be a part of its group")
		}
	})

	// Query singleton user.
	t.Run("SingletonUser", func(t *testing.T) {
		uid, err := Groups().SingletonUser(3)
		if err != nil {
			t.Errorf("error querying singleton user : %v\n", err)
		}

		fmt.Printf("single user : %d\n", uid)
	})

	// Add a user to a group.
	t.Run("AddUser1", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Errorf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		err = Groups().AddUser(tx, 3, 2)
		if err == nil {
			t.Errorf("should have failed because group is singleton")
		}

		if err == nil {
			err = tx.Commit()
			if err != nil {
				t.Errorf("error committing transaction : %v\n", err)
			}
		}
	})

	// Add a user to a group.
	t.Run("AddUser2", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Errorf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		err = Groups().AddUser(tx, 6, 1)
		if err != nil {
			t.Errorf("error adding user to general group : %v\n", err)
		}
		err = Groups().AddUser(tx, 6, 2)
		if err != nil {
			t.Errorf("error adding user to general group : %v\n", err)
		}

		if err == nil {
			err = tx.Commit()
			if err != nil {
				t.Errorf("error committing transaction : %v\n", err)
			}
		}
	})

	// Remove a user from a group.
	t.Run("RemoveUser1", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Errorf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		err = Groups().RemoveUser(tx, 3, 1)
		if err == nil {
			t.Errorf("should have failed because group is singleton")
		}

		if err == nil {
			err = tx.Commit()
			if err != nil {
				t.Errorf("error committing transaction : %v\n", err)
			}
		}
	})

	// Remove a user from a group.
	t.Run("RemoveUser2", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Errorf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		err = Groups().RemoveUser(tx, 6, 1)
		if err != nil {
			t.Errorf("error removing user from general group : %v\n", err)
		}
		err = Groups().RemoveUser(tx, 6, 2)
		if err != nil {
			t.Errorf("error removing user from general group : %v\n", err)
		}

		if err == nil {
			err = tx.Commit()
			if err != nil {
				t.Errorf("error committing transaction : %v\n", err)
			}
		}
	})
}
