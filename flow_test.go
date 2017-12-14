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
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

// error0 expects only an error value as its argument.
func error0(err error) error {
	if err != nil {
		gt.Errorf("%v", err)
	}
	return err
}

// error1 expects a value and an error value as its arguments.
func error1(val1 interface{}, err error) interface{} {
	if err != nil {
		gt.Errorf("%v", err)
		return nil
	}
	return val1
}

// fatal0 expects only an error value as its argument.
func fatal0(err error) {
	if err != nil {
		gt.Fatalf("%v", err)
	}
}

// fatal1 expects a value and an error value as its arguments.
func fatal1(val1 interface{}, err error) interface{} {
	if err != nil {
		gt.Fatalf("%v", err)
	}
	return val1
}

// assertEqual compares the two given values for equality.  In case of
// a difference, it errors with the given message.
func assertEqual(expected, observed interface{}, msgs ...string) {
	if expected == observed {
		return
	}

	gt.Errorf("expected : '%v', observed : '%v'\n\t%s", expected, observed, strings.Join(msgs, "\n\t"))
}

// assertNotEqual compares the two given values for inequality.  In
// case of equality, it errors with the given message.
func assertNotEqual(expected, observed interface{}, msgs ...string) {
	if expected != observed {
		return
	}

	gt.Errorf("expected : '%v', observed : '%v'\n\t%s", expected, observed, strings.Join(msgs, "\n\t"))
}

// Initialise DB connection.
func TestFlowInit(t *testing.T) {
	gt = t

	// Connect to the database.
	driver, connStr := "mysql", "travis@/flow"
	tdb := fatal1(sql.Open(driver, connStr)).(*sql.DB)
	RegisterDB(tdb)
}

// Test-local state.
var gt *testing.T

var dtID1, dtID2 DocTypeID
var dsID1, dsID2, dsID3, dsID4, dsID5 DocStateID
var daID1, daID2, daID3, daID4, daID5, daID6, daID7, daID8, daID9 DocActionID
var wfID1, wfID2 WorkflowID

var roleID1, roleID2 RoleID
var uID1, uID2, uID3, uID4 UserID
var gID1, gID2, gID3, gID4, gID5, gID6 GroupID

// Create operations.
func TestFlowCreate(t *testing.T) {
	gt = t

	t.Run("DocTypes", func(t *testing.T) {
		tx := fatal1(db.Begin()).(*sql.Tx)
		defer tx.Rollback()

		dtID1 = fatal1(DocTypes.New(tx, "Stor Request")).(DocTypeID)
		dtID2 = fatal1(DocTypes.New(tx, "Compute Request")).(DocTypeID)

		fatal0(tx.Commit())
	})

	t.Run("DocStates", func(t *testing.T) {
		tx := fatal1(db.Begin()).(*sql.Tx)
		defer tx.Rollback()

		dsID1 = fatal1(DocStates.New(tx, "Initial")).(DocStateID)
		dsID2 = fatal1(DocStates.New(tx, "Pending Approval")).(DocStateID)
		dsID3 = fatal1(DocStates.New(tx, "Approved")).(DocStateID)
		dsID4 = fatal1(DocStates.New(tx, "Rejected")).(DocStateID)
		dsID5 = fatal1(DocStates.New(tx, "Discarded")).(DocStateID)

		fatal0(tx.Commit())
	})

	t.Run("DocActions", func(t *testing.T) {
		tx := fatal1(db.Begin()).(*sql.Tx)
		defer tx.Rollback()

		daID1 = fatal1(DocActions.New(tx, "Initialise", false)).(DocActionID)
		daID2 = fatal1(DocActions.New(tx, "New", false)).(DocActionID)
		daID3 = fatal1(DocActions.New(tx, "Get", false)).(DocActionID)
		daID4 = fatal1(DocActions.New(tx, "Update", true)).(DocActionID)
		daID5 = fatal1(DocActions.New(tx, "Delete", true)).(DocActionID)
		daID6 = fatal1(DocActions.New(tx, "Approve", false)).(DocActionID)
		daID7 = fatal1(DocActions.New(tx, "Reject", false)).(DocActionID)
		daID8 = fatal1(DocActions.New(tx, "Return", false)).(DocActionID)
		daID9 = fatal1(DocActions.New(tx, "Discard", true)).(DocActionID)

		fatal0(tx.Commit())
	})

	t.Run("Workflows", func(t *testing.T) {
		tx := fatal1(db.Begin()).(*sql.Tx)
		defer tx.Rollback()

		wfID1 = fatal1(Workflows.New(tx, "Storage Management", dtID1, dsID1)).(WorkflowID)
		wfID2 = error1(Workflows.New(tx, "Compute Management", dtID2, dsID1)).(WorkflowID)

		fatal0(tx.Commit())
	})

	t.Run("Users", func(t *testing.T) {
		tx := fatal1(db.Begin()).(*sql.Tx)
		defer tx.Rollback()

		res, err := tx.Exec(`INSERT INTO users_master(first_name, last_name, email, active)
			VALUES('FN 1', 'LN 1', 'email1@example.com', 1)`)
		if err != nil {
			t.Fatalf("%v\n", err)
		}
		uid, _ := res.LastInsertId()
		uID1 = UserID(uid)
		gID1 = fatal1(Groups.NewSingleton(tx, uID1)).(GroupID)

		res, err = tx.Exec(`INSERT INTO users_master(first_name, last_name, email, active)
			VALUES('FN 2', 'LN 2', 'email2@example.com', 1)`)
		if err != nil {
			t.Fatalf("%v\n", err)
		}
		uid, _ = res.LastInsertId()
		uID2 = UserID(uid)
		gID2 = fatal1(Groups.NewSingleton(tx, uID2)).(GroupID)

		res, err = tx.Exec(`INSERT INTO users_master(first_name, last_name, email, active)
			VALUES('FN 3', 'LN 3', 'email3@example.com', 1)`)
		if err != nil {
			t.Errorf("%v\n", err)
		}
		uid, _ = res.LastInsertId()
		uID3 = UserID(uid)
		gID3 = fatal1(Groups.NewSingleton(tx, uID3)).(GroupID)

		res, err = tx.Exec(`INSERT INTO users_master(first_name, last_name, email, active)
			VALUES('FN 4', 'LN 4', 'email4@example.com', 1)`)
		if err != nil {
			t.Fatalf("%v\n", err)
		}
		uid, _ = res.LastInsertId()
		uID4 = UserID(uid)
		gID4 = fatal1(Groups.NewSingleton(tx, uID4)).(GroupID)

		fatal0(tx.Commit())
	})

	t.Run("Groups", func(t *testing.T) {
		tx := fatal1(db.Begin()).(*sql.Tx)
		defer tx.Rollback()

		gID5 = fatal1(Groups.New(tx, "Analysts", "G")).(GroupID)
		gID6 = fatal1(Groups.New(tx, "Managers", "G")).(GroupID)

		fatal0(tx.Commit())
	})

	t.Run("GroupsAddUsers", func(t *testing.T) {
		tx := fatal1(db.Begin()).(*sql.Tx)
		defer tx.Rollback()

		fatal0(Groups.AddUser(tx, gID5, uID1))
		fatal0(Groups.AddUser(tx, gID5, uID2))
		fatal0(Groups.AddUser(tx, gID5, uID3))

		fatal0(Groups.AddUser(tx, gID6, uID2))
		fatal0(Groups.AddUser(tx, gID6, uID3))
		fatal0(Groups.AddUser(tx, gID6, uID4))

		fatal0(tx.Commit())
	})

	t.Run("Roles", func(t *testing.T) {
		tx := fatal1(db.Begin()).(*sql.Tx)
		defer tx.Rollback()

		roleID1 = fatal1(Roles.New(tx, "Research Analyst")).(RoleID)
		roleID2 = fatal1(Roles.New(tx, "Manager")).(RoleID)

		fatal0(tx.Commit())
	})

	t.Run("RolesAddPermissions", func(t *testing.T) {
		tx := fatal1(db.Begin()).(*sql.Tx)
		defer tx.Rollback()

		fatal0(Roles.AddPermissions(tx, roleID1, dtID1, []DocActionID{daID1, daID2, daID3, daID4, daID8, daID9}))
		fatal0(Roles.AddPermissions(tx, roleID2, dtID1, []DocActionID{daID1, daID2, daID3, daID4, daID5, daID6, daID7, daID8, daID9}))

		fatal0(tx.Commit())
	})
}

// Entity listing.
func TestFlowList(t *testing.T) {
	gt = t
	var res interface{}

	t.Run("DocTypes", func(t *testing.T) {
		var dts []*DocType
		if res = error1(DocTypes.List(0, 0)); res == nil {
			return
		}
		dts = res.([]*DocType)
		assertEqual(2, len(dts))
	})

	t.Run("DocStates", func(t *testing.T) {
		var dss []*DocState
		if res = error1(DocStates.List(0, 0)); res == nil {
			return
		}
		dss = res.([]*DocState)
		// There is a pre-defined reserved state for children.
		assertEqual(6, len(dss))
	})

	t.Run("DocActions", func(t *testing.T) {
		var das []*DocAction
		if res = error1(DocActions.List(0, 0)); res == nil {
			return
		}
		das = res.([]*DocAction)
		assertEqual(9, len(das))
	})

	t.Run("Workflows", func(t *testing.T) {
		if res = error1(Workflows.List(0, 0)); res == nil {
			return
		}
		wfs := res.([]*Workflow)
		assertEqual(2, len(wfs))
	})

	t.Run("Users", func(t *testing.T) {
		if res = error1(Users.List("", 0, 0)); res == nil {
			return
		}
		us := res.([]*User)
		assertEqual(4, len(us))

		if res = error1(Users.List("LN 4", 0, 0)); res == nil {
			return
		}
		us = res.([]*User)
		assertEqual(1, len(us))
	})

	t.Run("Groups", func(t *testing.T) {
		var gs []*Group
		if res = error1(Groups.List(0, 0)); res == nil {
			return
		}
		gs = res.([]*Group)
		assertEqual(6, len(gs))
	})

	t.Run("Roles", func(t *testing.T) {
		var rs []*Role
		if res = error1(Roles.List(0, 0)); res == nil {
			return
		}
		rs = res.([]*Role)
		// There are two pre-defined roles for administrators.
		assertEqual(4, len(rs))
	})
}

// Retrieval of individual entities.
func TestFlowGet(t *testing.T) {
	gt = t
	var res interface{}

	t.Run("DocTypes", func(t *testing.T) {
		var dt *DocType
		if res = error1(DocTypes.GetByName("Compute Request")); res == nil {
			return
		}
		dt = res.(*DocType)
		assertEqual("Compute Request", dt.Name)

		var dt2 *DocType
		if res = error1(DocTypes.Get(dt.ID)); res == nil {
			return
		}
		dt2 = res.(*DocType)
		assertEqual("Compute Request", dt2.Name)
	})

	t.Run("DocStates", func(t *testing.T) {
		var ds *DocState
		if res = error1(DocStates.GetByName("Approved")); res == nil {
			return
		}
		ds = res.(*DocState)
		assertEqual("Approved", ds.Name)

		var ds2 *DocState
		if res = error1(DocStates.Get(ds.ID)); res == nil {
			return
		}
		ds2 = res.(*DocState)
		assertEqual("Approved", ds2.Name)
	})

	t.Run("DocActions", func(t *testing.T) {
		var da *DocAction
		if res = error1(DocActions.GetByName("Reject")); res == nil {
			return
		}
		da = res.(*DocAction)
		assertEqual("Reject", da.Name)

		var da2 *DocAction
		if res = error1(DocActions.Get(da.ID)); res == nil {
			return
		}
		da2 = res.(*DocAction)
		assertEqual("Reject", da2.Name)
	})

	t.Run("Workflows", func(t *testing.T) {
		if res = error1(Workflows.GetByDocType(dtID1)); res == nil {
			return
		}
		wf := res.(*Workflow)
		assertEqual("Storage Management", wf.Name)
		assertEqual(wfID1, wf.ID)
	})

	t.Run("Groups", func(t *testing.T) {
		var g *Group
		if res = error1(Groups.Get(gID1)); res == nil {
			return
		}
		g = res.(*Group)

		var u *User
		if res = error1(Groups.SingletonUser(gID1)); res == nil {
			return
		}
		u = res.(*User)

		assertEqual(u.Email, g.Name, "singleton group name should match corresponding user's e-mail")

		if res = error1(Groups.HasUser(gID6, uID4)); res == nil {
			return
		}
		ok := res.(bool)
		assertEqual(true, ok)
	})

	t.Run("Roles", func(t *testing.T) {
		var dt *DocType
		if res = error1(DocTypes.Get(dtID1)); res == nil {
			return
		}
		dt = res.(*DocType)

		if res = error1(Roles.Permissions(roleID1)); res == nil {
			return
		}
		perms := res.(map[string]struct {
			DocTypeID DocTypeID
			Actions   []*DocAction
		})
		assertEqual(1, len(perms))
		assertEqual(6, len(perms[dt.Name].Actions))

		if res = error1(Roles.HasPermission(roleID2, dtID1, daID6)); res == nil {
			return
		}
		assertEqual(true, res.(bool))
	})
}

// Entity update operations.
func TestFlowUpdate(t *testing.T) {
	gt = t
	var res interface{}

	t.Run("DocTypeRename", func(t *testing.T) {
		tx := fatal1(db.Begin()).(*sql.Tx)
		defer tx.Rollback()

		if err := error0(DocTypes.Rename(tx, dtID1, "Storage Request")); err != nil {
			return
		}

		fatal0(tx.Commit())

		if res = error1(DocTypes.Get(dtID1)); res == nil {
			return
		}
		obj := res.(*DocType)
		assertEqual("Storage Request", obj.Name)
	})

	t.Run("DocStateRename", func(t *testing.T) {
		tx := fatal1(db.Begin()).(*sql.Tx)
		defer tx.Rollback()

		if err := error0(DocStates.Rename(tx, dsID1, "Draft")); err != nil {
			return
		}

		fatal0(tx.Commit())

		if res = error1(DocStates.Get(dsID1)); res == nil {
			return
		}
		obj := res.(*DocState)
		assertEqual("Draft", obj.Name)
	})

	t.Run("DocActionRename", func(t *testing.T) {
		tx := fatal1(db.Begin()).(*sql.Tx)
		defer tx.Rollback()

		if res = error0(DocActions.Rename(tx, daID1, "List")); res != nil {
			return
		}

		fatal0(tx.Commit())

		if res = error1(DocActions.Get(daID1)); res == 0 {
			return
		}
		obj := res.(*DocAction)
		assertEqual("List", obj.Name)
	})

	t.Run("WorkflowsSetActive", func(t *testing.T) {
		tx := fatal1(db.Begin()).(*sql.Tx)
		defer tx.Rollback()

		if res = error0(Workflows.SetActive(tx, wfID1, false)); res != nil {
			return
		}

		fatal0(tx.Commit())

		if res = error1(Workflows.Get(wfID1)); res == nil {
			return
		}
		wf := res.(*Workflow)
		assertEqual(false, wf.Active)

		tx = fatal1(db.Begin()).(*sql.Tx)
		defer tx.Rollback()

		if res = error0(Workflows.SetActive(tx, wfID1, true)); res != nil {
			return
		}

		fatal0(tx.Commit())

		if res = error1(Workflows.Get(wfID1)); res == nil {
			return
		}
		wf = res.(*Workflow)
		assertEqual(true, wf.Active)
	})

	t.Run("GroupRename", func(t *testing.T) {
		tx := fatal1(db.Begin()).(*sql.Tx)
		defer tx.Rollback()

		if res = error0(Groups.Rename(tx, gID5, "Research Associates")); res != nil {
			return
		}

		fatal0(tx.Commit())

		if res = error1(Groups.Get(gID5)); res == 0 {
			return
		}
		obj := res.(*Group)
		assertEqual("Research Associates", obj.Name)
	})

	t.Run("GroupsDeleteUsers", func(t *testing.T) {
		tx := fatal1(db.Begin()).(*sql.Tx)
		defer tx.Rollback()

		error0(Groups.RemoveUser(tx, gID5, uID3))

		fatal0(tx.Commit())

		if res = error1(Groups.Users(gID5)); res == nil {
			return
		}
		objs := res.([]*User)
		assertEqual(2, len(objs))
	})

	t.Run("RolesRename", func(t *testing.T) {
		tx := fatal1(db.Begin()).(*sql.Tx)
		defer tx.Rollback()

		if err := error0(Roles.Rename(tx, roleID1, "Analyst")); err != nil {
			return
		}

		fatal0(tx.Commit())

		if res = error1(Roles.Get(roleID1)); res == 0 {
			return
		}
		obj := res.(*Role)
		assertEqual("Analyst", obj.Name)
	})

	t.Run("RolesDeletePerm", func(t *testing.T) {
		tx := fatal1(db.Begin()).(*sql.Tx)
		defer tx.Rollback()

		if err := error0(Roles.RemovePermissions(tx, roleID1, dtID1, []DocActionID{daID8})); err != nil {
			return
		}

		fatal0(tx.Commit())

		if res = error1(Roles.HasPermission(roleID1, dtID1, daID8)); res == nil {
			return
		}
		assertEqual(false, res.(bool))
	})
}

// Entity deletion operations.
func TestFlowDelete(t *testing.T) {
	gt = t
	var res interface{}

	t.Run("GroupsDelete", func(t *testing.T) {
		tx := fatal1(db.Begin()).(*sql.Tx)
		defer tx.Rollback()

		assertNotEqual(nil, Groups.Delete(tx, gID1), "it should not be possible to delete a singleton group")
		assertEqual(nil, Groups.Delete(tx, gID6), "it should be possible to delete a general group")

		fatal0(tx.Commit())
	})

	t.Run("RolesDelete", func(t *testing.T) {
		tx := fatal1(db.Begin()).(*sql.Tx)
		defer tx.Rollback()

		assertEqual(nil, Roles.Delete(tx, roleID1))

		fatal0(tx.Commit())

		if res = error1(Roles.List(0, 0)); res == nil {
			return
		}
		objs := res.([]*Role)
		// There are two pre-defined roles for administrators.
		assertEqual(3, len(objs))
	})
}

// Tear down.
func TestFlowTearDown(t *testing.T) {
	gt = t

	tx := fatal1(db.Begin()).(*sql.Tx)
	defer tx.Rollback()

	error1(tx.Exec(`DELETE FROM wf_ac_group_roles`))
	error1(tx.Exec(`DELETE FROM wf_ac_group_hierarchy`))
	error1(tx.Exec(`DELETE FROM wf_access_contexts`))

	error1(tx.Exec(`DELETE FROM wf_group_users`))
	error1(tx.Exec(`DELETE FROM wf_groups_master`))
	error1(tx.Exec(`DELETE FROM users_master`))
	error1(tx.Exec(`DELETE FROM wf_role_docactions`))
	error1(tx.Exec(`DELETE FROM wf_roles_master WHERE id > 2`))

	error1(tx.Exec(`DELETE FROM wf_workflow_nodes`))
	error1(tx.Exec(`DELETE FROM wf_workflows`))
	error1(tx.Exec(`DELETE FROM wf_docactions_master`))
	error1(tx.Exec(`DELETE FROM wf_docstates_master WHERE id > 1`))
	error1(tx.Exec(`DELETE FROM wf_doctypes_master`))

	fatal0(tx.Commit())
}
