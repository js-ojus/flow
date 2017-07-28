// (c) Copyright 2015 JONNALAGADDA Srinivas
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
	"errors"
	"fmt"
	"strings"
)

// GroupID is the type of unique group identifiers.
type GroupID int64

// Group represents a specified collection of users.  A user belongs
// to zero or more groups.
type Group struct {
	id        GroupID // Globally-unique ID
	name      string  // Globally-unique name
	groupType string  // Is this a user-specific group? Etc.
}

// NewSingletonGroup creates a singleton group associated with the
// given user.  The e-mail address of the user is used as the name of
// the group.  This serves as the linking identifier.
func NewSingletonGroup(otx *sql.Tx, uid UserID, email string) (GroupID, error) {
	var tx *sql.Tx
	if otx == nil {
		tx, err := db.Begin()
		if err != nil {
			return 0, err
		}
		defer tx.Rollback()
	} else {
		tx = otx
	}

	res, err := tx.Exec("INSERT INTO wf_groups_master(name, group_type) VALUES(?, ?)", email, "S")
	if err != nil {
		return 0, err
	}
	var gid int64
	gid, err = res.LastInsertId()
	if err != nil {
		return 0, err
	}

	res, err = tx.Exec("INSERT INTO wf_group_users(group_id, user_id) VALUES(?, ?)", gid, uid)
	if err != nil {
		return 0, err
	}
	_, err = res.LastInsertId()
	if err != nil {
		return 0, err
	}

	if otx == nil {
		err = tx.Commit()
		if err != nil {
			return 0, err
		}
	}

	return GroupID(gid), nil
}

// NewGroup creates a new group that can be populated with users
// later.
func NewGroup(otx *sql.Tx, name string, gtype string) (GroupID, error) {
	name = strings.TrimSpace(name)
	gtype = strings.TrimSpace(gtype)
	if name == "" || gtype == "" {
		return 0, errors.New("group name and type must not be empty")
	}
	switch gtype {
	case "G": // General
	// Nothing to do

	default:
		return 0, errors.New("unknown group type")
	}

	var tx *sql.Tx
	if otx == nil {
		tx, err := db.Begin()
		if err != nil {
			return 0, err
		}
		defer tx.Rollback()
	} else {
		tx = otx
	}

	res, err := tx.Exec("INSERT INTO wf_groups_master(name, group_type) VALUES(?, ?)", name, gtype)
	if err != nil {
		return 0, err
	}
	var gid int64
	gid, err = res.LastInsertId()
	if err != nil {
		return 0, err
	}

	if otx == nil {
		err = tx.Commit()
		if err != nil {
			return 0, err
		}
	}

	return GroupID(gid), nil
}

// GetGroup initialises the group by reading from database.
func GetGroup(gid GroupID) (*Group, error) {
	if gid <= 0 {
		return nil, errors.New("group ID should be a positive integer")
	}

	var tid GroupID
	var name string
	var gtype string
	row := db.QueryRow("SELECT id, name, group_type FROM groups_master WHERE id = ?", gid)
	err := row.Scan(&tid, &name, &gtype)
	if err != nil {
		return nil, err
	}
	if gtype == "S" {
		q := `
		SELECT status FROM wf_users_master_v
		WHERE id = (SELECT user_id FROM wf_group_users WHERE group_id = ?)
		`
		var active bool
		row = db.QueryRow(q, gid)
		err = row.Scan(&active)
		if err != nil {
			return nil, err
		}
		if !active {
			return nil, errors.New("user corresponding to this singleton group is currently inactive")
		}
	}

	g := &Group{id: gid, name: name, groupType: gtype}
	return g, nil
}

// DeleteGroup deletes the given group from the system, if no access
// context is actively using it.
func DeleteGroup(otx *sql.Tx, gid GroupID) error {
	if gid <= 0 {
		return errors.New("group ID must be a positive integer")
	}

	row := db.QueryRow("SELECT group_type FROM wf_groups_master WHERE group_id = ?", gid)
	var gtype string
	err := row.Scan(&gtype)
	if err != nil {
		return err
	}
	if gtype == "S" {
		return errors.New("singleton groups cannot be deleted")
	}

	row = db.QueryRow("SELECT COUNT(*) FROM wf_access_contexts WHERE group_id = ?", gid)
	var n int64
	err = row.Scan(&n)
	if n > 0 {
		return errors.New("group is being used in at least one access context; cannot delete")
	}

	var tx *sql.Tx
	if otx == nil {
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()
	} else {
		tx = otx
	}

	_, err = tx.Exec("DELETE FROM wf_group_users WHERE group_id = ?", gid)
	if err != nil {
		return err
	}
	res, err := tx.Exec("DELETE FROM wf_groups_master WHERE id = ?", gid)
	if err != nil {
		return err
	}
	n, err = res.RowsAffected()
	if n != 1 {
		return fmt.Errorf("expected number of affected rows : 1; actual affected : %d", n)
	}

	if otx == nil {
		err = tx.Commit()
		if err != nil {
			return err
		}
	}

	return nil
}

// ID answers this group's identifier.
func (g *Group) ID() GroupID {
	return g.id
}

// Name answers this group's name.
func (g *Group) Name() string {
	return g.name
}

// GroupType answers the nature of this group.  For instance, if this
// group was auto-created as the native group of a user account, or is
// a collection of users, etc.
func (g *Group) GroupType() string {
	return g.groupType
}

// HasUser answers `true` if this group includes the given user;
// `false` otherwise.
func (g *Group) HasUser(uid UserID) (bool, error) {
	var count int64
	row := db.QueryRow("SELECT COUNT(*) FROM wf_group_users WHERE group_id = ? AND user_id = ? LIMIT 1", g.id, uid)
	err := row.Scan(&count)
	if err != nil {
		return false, err
	}

	if count == 0 {
		return false, nil
	}
	return true, nil
}

// SingletonUser answer the user ID of the corresponding user, if this
// group is a singleton group.
func (g *Group) SingletonUser() (UserID, error) {
	if g.groupType != "S" {
		return 0, errors.New("this group is not a singleton group")
	}

	var uid UserID
	row := db.QueryRow("SELECT user_id FROM wf_group_users WHERE group_id = ?", g.id)
	err := row.Scan(&uid)
	if err != nil {
		return 0, err
	}
	return uid, nil
}

// AddUser adds the given user as a member of this group.
func (g *Group) AddUser(otx *sql.Tx, uid UserID) error {
	if g.groupType == "S" {
		return errors.New("cannot add users to a singleton group")
	}
	if uid <= 0 {
		return errors.New("user ID must be a positive integer")
	}

	var tx *sql.Tx
	if otx == nil {
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()
	} else {
		tx = otx
	}

	_, err := tx.Exec("INSERT INTO wf_group_users(group_id, user_id) VALUES(?, ?)", g.id, uid)
	if err != nil {
		return err
	}
	if otx == nil {
		err = tx.Commit()
		if err != nil {
			return err
		}
	}

	return nil
}

// RemoveUser removes the given user from this group, if the user is a
// member of the group.  This operation is idempotent.
func (g *Group) RemoveUser(otx *sql.Tx, uid UserID) error {
	if g.groupType == "S" {
		return errors.New("cannot delete users from a singleton group")
	}
	if uid <= 0 {
		return errors.New("user ID must be a positive integer")
	}

	var tx *sql.Tx
	if otx == nil {
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()
	} else {
		tx = otx
	}

	res, err := tx.Exec("DELETE FROM wf_group_users WHERE group_id = ? AND user_id = ?", g.id, uid)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("expected number of affected rows : 1; actual affected : %d", n)
	}

	if otx == nil {
		err = tx.Commit()
		if err != nil {
			return err
		}
	}

	return nil
}
