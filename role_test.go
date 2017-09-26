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

var (
	roles = []string{"ADMIN", "USER"}
)

// Driver test function.
func TestRoles01(t *testing.T) {
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

		_, err = tx.Exec(`DELETE FROM wf_role_docactions`)
		if err != nil {
			t.Fatalf("error running transaction : %v\n", err)
		}

		_, err = tx.Exec(`DELETE FROM wf_roles_master`)
		if err != nil {
			t.Fatalf("error running transaction : %v\n", err)
		}

		_, err = tx.Exec(`DELETE FROM wf_docactions_master`)
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

	// Test local state.
	var adminID RoleID
	var userID RoleID

	// Test create operations.
	t.Run("New", func(t *testing.T) {
		// Create a couple of roles.
		tx, err := db.Begin()
		if err != nil {
			t.Fatalf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		adminID, err = Roles().New(tx, roles[0])
		if err != nil {
			t.Fatalf("error creating role '%s' : %v\n", roles[0], err)
		}
		userID, err = Roles().New(tx, roles[1])
		if err != nil {
			t.Fatalf("error creating role '%s' : %v\n", roles[1], err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("error committing transaction : %v\n", err)
		}
	})

	// Test reading.
	t.Run("Read", func(t *testing.T) {
		_, err := Roles().List(0, 0)
		if err != nil {
			t.Fatalf("error : %v", err)
		}

		_, err = Roles().Get(adminID)
		if err != nil {
			t.Fatalf("error getting role '%d' : %v\n", adminID, err)
		}

		_, err = Roles().GetByName(roles[1])
		if err != nil {
			t.Fatalf("error getting role '%s' : %v\n", roles[1], err)
		}
	})

	// Test renaming.
	t.Run("Update", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Fatalf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		err = Roles().Rename(tx, adminID, "Administrator")
		if err != nil {
			t.Fatalf("error renaming role '%d' : %v\n", adminID, err)
		}
		err = Roles().Rename(tx, adminID, roles[0])
		if err != nil {
			t.Fatalf("error renaming role '%d' : %v\n", adminID, err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("error committing transaction : %v\n", err)
		}
	})

	// Test permissions operations.
	t.Run("Permissions", func(t *testing.T) {
		// Add a permission.
		tx, err := db.Begin()
		if err != nil {
			t.Fatalf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		dtypeStorReqID, err := DocTypes().New(tx, dtypeStorReq)
		if err != nil {
			t.Fatalf("error creating document type '%s' : %v\n", dtypeStorReq, err)
		}
		var da DocActionID
		for _, name := range actions {
			da, err = DocActions().New(tx, name)
			if err != nil {
				t.Fatalf("error creating document action '%s' : %v\n", name, err)
			}
		}

		err = Roles().AddPermission(tx, adminID, dtypeStorReqID, da)
		if err != nil {
			t.Fatalf("error adding permission : %v\n", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("error committing transaction : %v\n", err)
		}

		// List permissions.
		perms, err := Roles().Permissions(adminID)
		if err != nil {
			t.Fatalf("unable to query permissions : %v\n", err)
		}
		if len(perms) != 1 {
			t.Fatalf("permission doctype count -- expected : 1, got : %d\n", len(perms))
		}

		ok, err := Roles().HasPermission(adminID, dtypeStorReqID, da)
		if err != nil {
			t.Fatalf("unable to query permission : %v\n", err)
		}
		if !ok {
			t.Fatalf("permission on doctype '%d' for action '%d' -- extected : true, got : %v\n", dtypeStorReqID, da, ok)
		}

		// Remove a permission.
		tx, err = db.Begin()
		if err != nil {
			t.Fatalf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		err = Roles().RemovePermission(tx, adminID, dtypeStorReqID, da)
		if err != nil {
			t.Fatalf("error removing permission : %v\n", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("error committing transaction : %v\n", err)
		}

		// Verify removal.
		ok, err = Roles().HasPermission(adminID, dtypeStorReqID, da)
		if err != nil {
			t.Fatalf("unable to query permission : %v\n", err)
		}
		if ok {
			t.Fatalf("permission on doctype '%d' for action '%d' -- extected : false, got : %v\n", dtypeStorReqID, da, ok)
		}
	})

	// Test deleting.
	t.Run("Delete", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Fatalf("error starting transaction : %v\n", err)
		}
		defer tx.Rollback()

		err = Roles().Delete(tx, userID)
		if err != nil {
			t.Fatalf("error deleting role '%d' : %v\n", adminID, err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("error committing transaction : %v\n", err)
		}
	})
}
