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
	"math"
	"strings"
)

// AccessContextID is the type unique access context identifiers.
type AccessContextID int64

// AccessContext is a namespace that provides an environment for
// workflow execution.
//
// It is an environment in which users are mapped into an hierarchy
// that determines certain aspects of workflow control. This
// hierarchy, usually, but not necessarily, reflects an organogram. In
// each access context, applicable groups are mapped to their
// respective intended permissions.  This mapping happens through
// roles.
//
// Each workflow that operates on a document type is given an
// associated access context.  This context is used to determine
// workflow routing, possible branching and rendezvous points.
//
// Please note that the same workflow may operate independently in
// multiple unrelated access contexts.  Thus, a workflow is not
// limited to one access context.  Conversely, an access context can
// have several workflows operating on it, for various document types.
// Therefore, the relationship between workflows and access contexts
// is M:N.
//
// For complex organisational requirements, the name of the access
// context can be made hierarchical with a suitable delimiter.  For
// example:
//
//     - IN:south:HYD:BR-101
//     - sbu-08/client-0249/prj-006348
type AccessContext struct {
	ID         AccessContextID           `json:"ID"`                   // Unique identifier of this access context
	Name       string                    `json:"Name"`                 // Globally-unique namespace; can be a department, project, location, branch, etc.
	Active     bool                      `json:"Active"`               // Can a workflow be initiated in this context?
	GroupRoles map[GroupID]*AcGroupRoles `json:"GroupRoles,omitempty"` // Mapping of groups to their roles.
	UserLevels map[UserID]AcUserLevel    `json:"UserLevels,omitempty"` // Mapping of users to their hierarchy.
}

// AcGroupRoles holds the information of the various roles that each
// group has been assigned in this access context.
type AcGroupRoles struct {
	Group string `json:"Group"` // Group whose roles this represents
	Roles []Role `json:"Roles"` // Map holds the role assignments to groups
}

// AcUserLevel holds the information of the level assigned to a user
// in the hierarchy of this access context.
type AcUserLevel struct {
	User  User  `json:"User"`  // User whose level this represents
	Level int64 `json:"Level"` // Level assigned to this user in this access context
}

// Unexported type, only for convenience methods.
type _AccessContexts struct{}

var _accCtxs *_AccessContexts

func init() {
	_accCtxs = &_AccessContexts{}
}

// AccessContexts provides a resource-like interface to access
// contexts in the system.
func AccessContexts() *_AccessContexts {
	return _accCtxs
}

// New creates a new access context with the globally-unique name
// given.
func (acs *_AccessContexts) New(otx *sql.Tx, name string) (AccessContextID, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, errors.New("access context name should be non-empty")
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

	q := `INSERT INTO wf_access_contexts(name) VALUES(?)`
	res, err := tx.Exec(q, name)
	if err != nil {
		return 0, err
	}
	acID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	if otx == nil {
		err := tx.Commit()
		if err != nil {
			return 0, err
		}
	}

	return AccessContextID(acID), nil
}

// List answers a list of access contexts defined in the system.
//
// Result set begins with ID >= `offset`, and has not more than
// `limit` elements.  A value of `0` for `offset` fetches from the
// beginning, while a value of `0` for `limit` fetches until the end.
func (acs *_AccessContexts) List(offset, limit int64) ([]*AccessContext, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.New("offset and limit should be non-negative integers")
	}
	if limit == 0 {
		limit = math.MaxInt64
	}

	q := `
	SELECT id, name
	FROM wf_access_contexts
	ORDER BY id
	LIMIT ? OFFSET ?
	`
	rows, err := db.Query(q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ary := make([]*AccessContext, 0, 10)
	for rows.Next() {
		var elem AccessContext
		err = rows.Scan(&elem.ID, &elem.Name)
		if err != nil {
			return nil, err
		}
		ary = append(ary, &elem)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ary, nil
}

// Get fetches the requested access context that determines how the
// workflows that operate in its context run.
func (acs *_AccessContexts) Get(id AccessContextID) (*AccessContext, error) {
	q := `SELECT id, name FROM wf_access_contexts WHERE id = ?`
	res := db.QueryRow(q, id)
	var elem AccessContext
	err := res.Scan(&elem.ID, &elem.Name)
	if err != nil {
		return nil, err
	}

	q = `
	SELECT agrs.group_id, gm.name, agrs.role_id, rm.name
	FROM wf_ac_group_roles agrs
	JOIN wf_groups_master gm ON gm.id = agrs.group_id
	JOIN wf_roles_master rm ON rm.id = agrs.role_id
	WHERE agrs.ac_id = ?
	ORDER BY agrs.group_id
	`
	rows, err := db.Query(q, id)
	if err != nil {
		return nil, err
	}

	elem.GroupRoles = make(map[GroupID]*AcGroupRoles)
	var curGid int64
	for rows.Next() {
		var gid int64
		var gname string
		var role Role
		err = rows.Scan(&gid, &gname, &role.ID, &role.Name)
		if err != nil {
			return nil, err
		}

		var gr *AcGroupRoles
		if curGid == gid {
			gr = elem.GroupRoles[GroupID(gid)]
		} else {
			gr = &AcGroupRoles{Group: gname, Roles: make([]Role, 0, 4)}
			curGid = gid
		}
		gr.Roles = append(gr.Roles, role)
		elem.GroupRoles[GroupID(gid)] = gr
	}
	if rows.Err() != nil {
		return nil, err
	}

	q = `
	SELECT auls.user_id, um.first_name, um.last_name, um.email, auls.ulevel
	FROM wf_ac_user_levels auls
	JOIN wf_users_master um ON um.id = auls.user_id
	WHERE agrs.ac_id = ?
	ORDER BY auls.user_id
	`
	rows, err = db.Query(q, id)
	if err != nil {
		return nil, err
	}

	elem.UserLevels = make(map[UserID]AcUserLevel)
	for rows.Next() {
		var level int64
		var u User
		err = rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &level)
		if err != nil {
			return nil, err
		}

		elem.UserLevels[UserID(u.ID)] = AcUserLevel{User: u, Level: level}
	}
	if rows.Err() != nil {
		return nil, err
	}

	return &elem, nil
}

// AddGroupRole assigns the specified role to the given group, if it
// is not already assigned.
func (ac *AccessContext) AddGroupRole(otx *sql.Tx, gid GroupID, rid RoleID) error {
	if gid <= 0 || rid <= 0 {
		return errors.New("group ID and role ID should be positive integers")
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

	_, err := tx.Exec(`INSERT INTO wf_ac_group_roles(ac_id, group_id, role_id) VALUES(?, ?, ?)`, ac.ID, gid, rid)
	if err != nil {
		return err
	}

	if otx == nil {
		err := tx.Commit()
		if err != nil {
			return err
		}
	}

	return nil
}

// RemoveGroupRole unassigns the specified role from the given group.
func (ac *AccessContext) RemoveGroupRole(otx *sql.Tx, gid GroupID, rid RoleID) error {
	if gid <= 0 || rid <= 0 {
		return errors.New("group ID and role ID should be positive integers")
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

	_, err := tx.Exec(`DELETE FROM wf_access_contexts WHERE ns_id = ? AND group_id = ? AND role_id = ?`, ac.ID, gid, rid)
	if err != nil {
		return err
	}

	if otx == nil {
		err := tx.Commit()
		if err != nil {
			return err
		}
	}

	return nil
}

// AddUser adds the given user to this access context, with the
// specified level in the hierarchy.
func (ac *AccessContext) AddUser(otx *sql.Tx, uid UserID, level int64) error {
	if uid <= 0 || level <= 0 {
		return errors.New("user ID and user level should be positive integers")
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

	q := `INSERT INTO wf_ac_user_levels(ac_id, user_id, ulevel) VALUES (?, ?, ?)`
	_, err := tx.Exec(q, ac.ID, uid, level)
	if err != nil {
		return err
	}

	if otx == nil {
		err := tx.Commit()
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteUser removes the given user from this access context.
func (ac *AccessContext) DeleteUser(otx *sql.Tx, uid UserID) error {
	if uid <= 0 {
		return errors.New("user ID should be positive integers")
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

	q := `DELETE FROM wf_ac_user_levels WHERE ac_id = ? AND user_id = ?`
	_, err := tx.Exec(q, ac.ID, uid)
	if err != nil {
		return err
	}

	if otx == nil {
		err := tx.Commit()
		if err != nil {
			return err
		}
	}

	return nil
}

// UserHasPermission answers `true` if the given user has the
// requested action enabled on the specified document type; `false`
// otherwise.
func (ac *AccessContext) UserHasPermission(uid UserID, dtype DocTypeID, action DocActionID) (bool, error) {
	if uid <= 0 || dtype <= 0 || action <= 0 {
		return false, errors.New("invalid user ID or document type or document action")
	}

	q := `
	SELECT role_id FROM wf_ac_perms_v
	WHERE ac_id = ?
	AND user_id = ?
	AND doctype_id = ?
	AND docaction_id = ?
	LIMIT 1
	`
	row := db.QueryRow(q, ac.ID, uid, dtype, action)
	var roleID int64
	err := row.Scan(&roleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GroupHasPermission answers `true` if the given group has the
// requested action enabled on the specified document type; `false`
// otherwise.
func (ac *AccessContext) GroupHasPermission(gid GroupID, dtype DocTypeID, action DocActionID) (bool, error) {
	if gid <= 0 || dtype <= 0 || action <= 0 {
		return false, errors.New("invalid group ID or document type or document action")
	}

	q := `
	SELECT role_id FROM wf_ac_perms_v
	WHERE ac_id = ?
	AND group_id = ?
	AND doctype_id = ?
	AND docaction_id = ?
	LIMIT 1
	`
	row := db.QueryRow(q, ac.ID, gid, dtype, action)
	var roleID int64
	err := row.Scan(&roleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
