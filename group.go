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
	"errors"
	"fmt"
	"math"
	"strings"
)

// GroupID is the type of unique group identifiers.
type GroupID int64

// Group represents a specified collection of users.  A user belongs
// to zero or more groups.
type Group struct {
	id    GroupID // Globally-unique ID
	name  string  // Globally-unique name
	gtype string  // Is this a user-specific group? Etc.
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
	return g.gtype
}

// Unexported type, only for convenience methods.
type _Groups struct{}

var _groups *_Groups

func init() {
	_groups = &_Groups{}
}

// Groups provides a resource-like interface to groups in the system.
func Groups() *_Groups {
	return _groups
}

// NewSingleton creates a singleton group associated with the given
// user.  The e-mail address of the user is used as the name of the
// group.  This serves as the linking identifier.
func (gs *_Groups) NewSingleton(otx *sql.Tx, uid UserID) (GroupID, error) {
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

	q := `
	INSERT INTO wf_groups_master(name, group_type)
	SELECT u.email, 'S'
	FROM wf_users_master u
	WHERE u.id = ?
	`
	res, err := tx.Exec(q, uid)
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

// New creates a new group that can be populated with users later.
func (gs *_Groups) New(otx *sql.Tx, name string, gtype string) (GroupID, error) {
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
	var id int64
	id, err = res.LastInsertId()
	if err != nil {
		return 0, err
	}

	if otx == nil {
		err = tx.Commit()
		if err != nil {
			return 0, err
		}
	}

	return GroupID(id), nil
}

// List answers a subset of the groups, based on the input
// specification.
//
// Result set begins with ID >= `offset`, and has not more than
// `limit` elements.  A value of `0` for `offset` fetches from the
// beginning, while a value of `0` for `limit` fetches until the end.
func (gs *_Groups) List(offset, limit int64) ([]*Group, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.New("offset and limit must be non-negative integers")
	}
	if limit == 0 {
		limit = math.MaxInt64
	}

	q := `
	SELECT id, name, group_type
	FROM wf_groups_master
	ORDER BY id
	LIMIT ? OFFSET ?
	`
	rows, err := db.Query(q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ary := make([]*Group, 0, 10)
	for rows.Next() {
		var g Group
		err = rows.Scan(&g.id, &g.name, &g.gtype)
		if err != nil {
			return nil, err
		}
		ary = append(ary, &g)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ary, nil
}

// Get initialises the group by reading from database.
func (gs *_Groups) Get(id GroupID) (*Group, error) {
	if id <= 0 {
		return nil, errors.New("group ID should be a positive integer")
	}

	var elem Group
	row := db.QueryRow("SELECT id, name, group_type FROM wf_groups_master WHERE id = ?", id)
	err := row.Scan(&elem.id, &elem.name, &elem.gtype)
	if err != nil {
		return nil, err
	}
	if elem.gtype == "S" {
		q := `
		SELECT active FROM wf_users_master
		WHERE id = (SELECT user_id FROM wf_group_users WHERE group_id = ?)
		`
		var active bool
		row = db.QueryRow(q, id)
		err = row.Scan(&active)
		if err != nil {
			return nil, err
		}
		if !active {
			return nil, errors.New("user corresponding to this singleton group is currently inactive")
		}
	}

	return &elem, nil
}

// Delete deletes the given group from the system, if no access
// context is actively using it.
func (gs *_Groups) Delete(otx *sql.Tx, id GroupID) error {
	if id <= 0 {
		return errors.New("group ID must be a positive integer")
	}

	row := db.QueryRow("SELECT group_type FROM wf_groups_master WHERE id = ?", id)
	var gtype string
	err := row.Scan(&gtype)
	if err != nil {
		return err
	}
	if gtype == "S" {
		return errors.New("singleton groups cannot be deleted")
	}

	row = db.QueryRow("SELECT COUNT(*) FROM wf_access_contexts WHERE group_id = ?", id)
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

	_, err = tx.Exec("DELETE FROM wf_group_users WHERE group_id = ?", id)
	if err != nil {
		return err
	}
	res, err := tx.Exec("DELETE FROM wf_groups_master WHERE id = ?", id)
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

// HasUser answers `true` if this group includes the given user;
// `false` otherwise.
func (gs *_Groups) HasUser(gid GroupID, uid UserID) (bool, error) {
	q := `
	SELECT id FROM wf_group_users
	WHERE group_id = ?
	AND user_id = ?
	ORDER BY id
	LIMIT 1
	`
	var id int64
	row := db.QueryRow(q, gid, uid)
	err := row.Scan(&id)
	switch {
	case err == sql.ErrNoRows:
		return false, errors.New("given user is not part of the specified group")

	case err != nil:
		return false, err

	default:
		return true, nil
	}
}

// SingletonUser answer the user ID of the corresponding user, if this
// group is a singleton group.
func (gs *_Groups) SingletonUser(gid GroupID) (UserID, error) {
	q := `
	SELECT gus.user_id FROM wf_group_users gus
	JOIN wf_groups_master gm ON gus.group_id = gm.id
	WHERE gm.id = ?
	AND gm.group_type = 'S'
	ORDER BY gus.id
	LIMIT 1
	`
	var uid UserID
	row := db.QueryRow(q, gid)
	err := row.Scan(&uid)
	switch {
	case err == sql.ErrNoRows:
		return 0, errors.New("given group is not a singleton group")

	case err != nil:
		return 0, err

	default:
		return uid, nil
	}
}

// AddUser adds the given user as a member of this group.
func (gs *_Groups) AddUser(otx *sql.Tx, gid GroupID, uid UserID) error {
	if gid <= 0 || uid <= 0 {
		return errors.New("group ID and user ID must be positive integers")
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

	var gtype string
	row := tx.QueryRow("SELECT group_type FROM wf_groups_master WHERE id = ?", gid)
	err := row.Scan(&gtype)
	if err != nil {
		return err
	}
	if gtype == "S" {
		return errors.New("cannot add users to singleton groups")
	}

	_, err = tx.Exec("INSERT INTO wf_group_users(group_id, user_id) VALUES(?, ?)", gid, uid)
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
func (gs *_Groups) RemoveUser(otx *sql.Tx, gid GroupID, uid UserID) error {
	if gid <= 0 || uid <= 0 {
		return errors.New("group ID and user ID must be positive integers")
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

	var gtype string
	row := tx.QueryRow("SELECT group_type FROM wf_groups_master WHERE id = ?", gid)
	err := row.Scan(&gtype)
	if err != nil {
		return err
	}
	if gtype == "S" {
		return errors.New("cannot remove users from singleton groups")
	}

	res, err := tx.Exec("DELETE FROM wf_group_users WHERE group_id = ? AND user_id = ?", gid, uid)
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
