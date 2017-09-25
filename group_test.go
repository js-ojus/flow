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

// Driver test function.
func TestGroups01(t *testing.T) {
	const genGrp = "research"

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

		_, err = tx.Exec(`DELETE FROM wf_group_users`)
		if err != nil {
			t.Fatalf("error running transaction : %v\n", err)
		}

		_, err = tx.Exec(`DELETE FROM wf_groups_master`)
		if err != nil {
			t.Fatalf("error running transaction : %v\n", err)
		}

		_, err = tx.Exec(`DELETE FROM users_master`)
		if err != nil {
			t.Fatalf("error running transaction : %v\n", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("error committing transaction : %v\n", err)
		}
	}()

	// Test local state.
	var u1, u2 int64
	var gs []*Group

	// Test CRL operations.
	t.Run("CRL", func(t *testing.T) {
		// Create required users.
		tx, err := db.Begin()
		if err != nil {
			t.Fatalf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		q := `
		INSERT INTO users_master(first_name, last_name, email, active)
		VALUES(?, ?, ?, ?)
		`
		res, err := tx.Exec(q, users[0].fname, users[0].lname, users[0].email, users[0].active)
		if err != nil {
			t.Fatalf("error running transaction : %v\n", err)
		}
		u1, _ = res.LastInsertId()

		q = `
		INSERT INTO users_master(first_name, last_name, email, active)
		VALUES(?, ?, ?, ?)
		`
		res, err = tx.Exec(q, users[1].fname, users[1].lname, users[1].email, users[1].active)
		if err != nil {
			t.Fatalf("error running transaction : %v\n", err)
		}
		u2, _ = res.LastInsertId()

		err = tx.Commit()
		if err != nil {
			t.Fatalf("error committing transaction : %v\n", err)
		}

		// Create required groups.
		tx, err = db.Begin()
		if err != nil {
			t.Fatalf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		_, err = Groups().NewSingleton(tx, UserID(u1))
		if err != nil {
			t.Fatalf("error adding singleton group for user '%d' : %v\n", u1, err)
		}
		_, err = Groups().NewSingleton(tx, UserID(u2))
		if err != nil {
			t.Fatalf("error adding singleton group for user '%d' : %v\n", u2, err)
		}
		_, err = Groups().New(tx, genGrp, "G")
		if err != nil {
			t.Fatalf("error adding group '%s' : %v\n", genGrp, err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("error committing transaction : %v\n", err)
		}

		// List groups.
		gs, err = Groups().List(0, 0)
		if err != nil {
			t.Fatalf("error : %v", err)
		}
		if len(gs) != 3 {
			t.Fatalf("listing groups -- expected : 3, got : %d\n", len(gs))
		}

		_, err = Groups().Get(gs[0].ID())
		if err != nil {
			t.Fatalf("error getting group : %v\n", err)
		}

		_, err = Users().SingletonGroupOf(UserID(u1))
		if err != nil {
			t.Fatalf("error querying groups of user '%d' : %v\n", u1, err)
		}

		_, err = Groups().SingletonUser(gs[1].ID())
		if err != nil {
			t.Fatalf("error querying singleton group '%d' : %v\n", gs[1].ID(), err)
		}

		// Test membership.
		ok, err := Groups().HasUser(gs[0].ID(), UserID(u1))
		if err != nil {
			t.Fatalf("error querying group users : %v\n", err)
		}

		if !ok {
			t.Fatalf("singleton user must be a part of its group")
		}
	})

	// Test update operations.
	t.Run("Update", func(t *testing.T) {
		// Test adding users.
		tx, err := db.Begin()
		if err != nil {
			t.Fatalf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		err = Groups().AddUser(tx, gs[0].ID(), UserID(u2))
		if err == nil {
			t.Fatalf("should have failed because group is singleton")
		}
		err = Groups().AddUser(tx, gs[2].ID(), UserID(u1))
		if err != nil {
			t.Fatalf("error adding user to general group : %v\n", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("error committing transaction : %v\n", err)
		}

		// Fetch groups a user is a member of.
		gids, err := Users().GroupsOf(UserID(u1))
		if err != nil {
			t.Fatalf("error querying groups of user '%d' : %v\n", 1, err)
		}
		if len(gids) != 2 {
			t.Fatalf("listing groups -- expected : 2, got : %d\n", len(gids))
		}
		gids, err = Users().GroupsOf(UserID(u2))
		if err != nil {
			t.Fatalf("error querying groups of user '%d' : %v\n", 1, err)
		}
		if len(gids) != 1 {
			t.Fatalf("listing groups -- expected : 1, got : %d\n", len(gids))
		}

		// Testing removing users.
		tx, err = db.Begin()
		if err != nil {
			t.Fatalf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		err = Groups().RemoveUser(tx, gs[0].ID(), UserID(u1))
		if err == nil {
			t.Fatalf("should have failed because group is singleton")
		}
		err = Groups().RemoveUser(tx, gs[2].ID(), UserID(u1))
		if err != nil {
			t.Fatalf("error removing user from general group : %v\n", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("error committing transaction : %v\n", err)
		}
	})

	// Test delete operations.
	t.Run("Delete", func(t *testing.T) {
		// Test deleting a singleton group.
		tx, err := db.Begin()
		if err != nil {
			t.Fatalf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		err = Groups().Delete(tx, gs[1].ID())
		if err == nil {
			t.Fatalf("should have failed because group is singleton")
		}
		err = Groups().Delete(tx, gs[2].ID())
		if err != nil {
			t.Fatalf("error deleting general group : %v\n", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("error committing transaction : %v\n", err)
		}
	})
}
