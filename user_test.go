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

var users = []struct {
	fname  string
	lname  string
	email  string
	active byte
}{
	{"Fname1", "Lname1", "user1@domain.com", 1},
	{"Fname2", "Lname2", "user2@domain.com", 1},
}

// Driver test function.
func TestUsers01(t *testing.T) {
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

		_, err = tx.Exec(`DELETE FROM users_master`)
		if err != nil {
			t.Fatalf("error running transaction : %v\n", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("error committing transaction : %v\n", err)
		}
	}()

	// Test CRL operations.
	t.Run("CRL", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Fatalf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		var u1, u2 int64
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
		_, err = tx.Exec(q, users[1].fname, users[1].lname, users[1].email, users[1].active)
		if err != nil {
			t.Fatalf("error running transaction : %v\n", err)
		}
		u2, _ = res.LastInsertId()

		err = tx.Commit()
		if err != nil {
			t.Fatalf("error committing transaction : %v\n", err)
		}

		// Test reading.
		_, err = Users.Get(UserID(u1))
		if err != nil {
			t.Fatalf("error getting user : %v\n", err)
		}

		_, err = Users.GetByEmail(users[1].email)
		if err != nil {
			t.Fatalf("error getting user : %v\n", err)
		}

		_, err = Users.List(0, 0)
		if err != nil {
			t.Fatalf("error : %v", err)
		}

		active, err := Users.IsActive(UserID(u2))
		if err != nil {
			t.Fatalf("error geting status of user : %v\n", err)
		}
		if !active {
			t.Fatalf("user status -- expected : 1, got : %v\n", active)
		}
	})
}
