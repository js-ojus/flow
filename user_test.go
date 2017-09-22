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
func TestUsers01(t *testing.T) {
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

	// List users.
	t.Run("List", func(t *testing.T) {
		us, err := Users().List(0, 0)
		if err != nil {
			t.Errorf("error : %v", err)
		}

		for _, u := range us {
			fmt.Printf("%#v\n", u)
		}
	})

	// Retrieve a specified user.
	t.Run("GetByID", func(t *testing.T) {
		u, err := Users().Get(1)
		if err != nil {
			t.Errorf("error getting user '1' : %v\n", err)
		}

		fmt.Printf("%#v\n", u)
	})

	// Check the status of a user.
	t.Run("IsActive", func(t *testing.T) {
		status, err := Users().IsActive(1)
		if err != nil {
			t.Errorf("error geting status of user '1' : %v\n", err)
		}

		fmt.Printf("%#v\n", status)
	})

	// Fetch groups a user is a member of.
	t.Run("GroupsOf", func(t *testing.T) {
		gs, err := Users().GroupsOf(1)
		if err != nil {
			t.Errorf("error querying groups of user '%d' : %v\n", 1, err)
		}

		for _, g := range gs {
			fmt.Printf("group : %d\n", g)
		}
	})

	// Fetch singleton group of a user.
	t.Run("SingletonGroupOf", func(t *testing.T) {
		g, err := Users().SingletonGroupOf(2)
		if err != nil {
			t.Errorf("error querying groups of user '%d' : %v\n", 1, err)
		}

		fmt.Printf("group : %d\n", g)
	})
}
