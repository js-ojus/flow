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
// It is an environment in which users are mapped into a hierarchy
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
	ID     AccessContextID `json:"ID"`               // Unique identifier of this access context
	Name   string          `json:"Name,omitempty"`   // Globally-unique namespace; can be a department, project, location, branch, etc.
	Active bool            `json:"Active,omitempty"` // Can a workflow be initiated in this context?
}

// AcGroupRoles holds the information of the various roles that each
// group has been assigned in this access context.
type AcGroupRoles struct {
	Group string `json:"Group"` // Group whose roles this represents
	Roles []Role `json:"Roles"` // Map holds the role assignments to groups
}

// AcGroup holds the information of a user together with the user's
// reporting authority.
type AcGroup struct {
	Group     `json:"Group"` // An assigned user
	ReportsTo GroupID        `json:"ReportsTo"` // Reporting authority of this user
}

// Unexported type, only for convenience methods.
type _AccessContexts struct{}

// AccessContexts provides a resource-like interface to access
// contexts in the system.
var AccessContexts _AccessContexts

// New creates a new access context with the globally-unique name
// given.
func (_AccessContexts) New(otx *sql.Tx, name string) (AccessContextID, error) {
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

	q := `INSERT INTO wf_access_contexts(name, active) VALUES(?, 1)`
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
func (_AccessContexts) List(prefix string, offset, limit int64) ([]*AccessContext, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.New("offset and limit should be non-negative integers")
	}
	if limit == 0 {
		limit = math.MaxInt64
	}

	var q string
	var rows *sql.Rows
	var err error

	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		q = `
		SELECT id, name, active
		FROM wf_access_contexts
		ORDER BY id
		LIMIT ? OFFSET ?
		`
		rows, err = db.Query(q, limit, offset)
	} else {
		q = `
		SELECT id, name, active
		FROM wf_access_contexts
		WHERE name LIKE ?
		ORDER BY id
		LIMIT ? OFFSET ?
		`
		rows, err = db.Query(q, prefix+"%", limit, offset)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ary := make([]*AccessContext, 0, 10)
	for rows.Next() {
		var elem AccessContext
		err = rows.Scan(&elem.ID, &elem.Name, &elem.Active)
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

// ListByGroup answers a list of access contexts in which the given
// group is included.
//
// Result set begins with ID >= `offset`, and has not more than
// `limit` elements.  A value of `0` for `offset` fetches from the
// beginning, while a value of `0` for `limit` fetches until the end.
func (_AccessContexts) ListByGroup(gid GroupID, offset, limit int64) ([]*AccessContext, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.New("offset and limit should be non-negative integers")
	}
	if limit == 0 {
		limit = math.MaxInt64
	}

	q := `
	SELECT ac.id, ac.name, ac.active
	FROM wf_access_contexts ac
	JOIN wf_ac_group_hierarchy agh ON agh.ac_id = ac.id
	WHERE agh.group_id = ?
	ORDER BY agh.ac_id
	LIMIT ? OFFSET ?
	`
	rows, err := db.Query(q, gid, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ary := make([]*AccessContext, 0, 10)
	for rows.Next() {
		var elem AccessContext
		err = rows.Scan(&elem.ID, &elem.Name, &elem.Active)
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

// ListByUser answers a list of access contexts in which the given
// group is included.
//
// Result set begins with ID >= `offset`, and has not more than
// `limit` elements.  A value of `0` for `offset` fetches from the
// beginning, while a value of `0` for `limit` fetches until the end.
func (_AccessContexts) ListByUser(uid UserID, offset, limit int64) ([]*AccessContext, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.New("offset and limit should be non-negative integers")
	}
	if limit == 0 {
		limit = math.MaxInt64
	}

	q := `
	SELECT ac.id, ac.name, ac.active
	FROM wf_access_contexts ac
	JOIN wf_ac_group_hierarchy agh ON agh.ac_id = ac.id
	WHERE agh.group_id = (
		SELECT gm.id
		FROM wf_groups_master gm
		JOIN wf_group_users gu ON gu.group_id = gm.id
		WHERE gu.user_id = ?
		AND gm.group_type = 'S'
	)
	ORDER BY agh.ac_id
	LIMIT ? OFFSET ?
	`
	rows, err := db.Query(q, uid, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ary := make([]*AccessContext, 0, 10)
	for rows.Next() {
		var elem AccessContext
		err = rows.Scan(&elem.ID, &elem.Name, &elem.Active)
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
func (_AccessContexts) Get(id AccessContextID) (*AccessContext, error) {
	q := `
	SELECT id, name, active
	FROM wf_access_contexts
	WHERE id = ?
	`
	res := db.QueryRow(q, id)
	var elem AccessContext
	err := res.Scan(&elem.ID, &elem.Name, &elem.Active)
	if err != nil {
		return nil, err
	}

	return &elem, nil
}

// Rename changes the name of the given access context to the
// specified new name.
func (_AccessContexts) Rename(otx *sql.Tx, id AccessContextID, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("access context name should be non-empty")
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

	q := `
	UPDATE wf_access_contexts
	SET name = ?
	WHERE id = ?
	`
	_, err := tx.Exec(q, name, id)
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

// SetActive updates the given access context with the new active
// status.
func (_AccessContexts) SetActive(otx *sql.Tx, id AccessContextID, active bool) error {
	act := 0
	if active {
		act = 1
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

	q := `
	UPDATE wf_access_contexts
	SET active = ?
	WHERE id = ?
	`
	_, err := tx.Exec(q, act, id)
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

// GroupRoles retrieves the groups --> roles mapping for this access
// context.
func (_AccessContexts) GroupRoles(id AccessContextID, gid GroupID, offset, limit int64) (map[GroupID]*AcGroupRoles, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.New("offset and limit should be non-negative integers")
	}
	if limit == 0 {
		limit = math.MaxInt64
	}

	q := `
	SELECT agrs.group_id, gm.name, agrs.role_id, rm.name
	FROM wf_ac_group_roles agrs
	JOIN wf_groups_master gm ON gm.id = agrs.group_id
	JOIN wf_roles_master rm ON rm.id = agrs.role_id
	WHERE agrs.ac_id = ?
	AND agrs.group_id = ?
	ORDER BY agrs.group_id
	LIMIT ? OFFSET ?
	`
	rows, err := db.Query(q, id, gid, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	grs := make(map[GroupID]*AcGroupRoles)
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
			gr = grs[GroupID(gid)]
		} else {
			gr = &AcGroupRoles{Group: gname, Roles: make([]Role, 0, 4)}
			curGid = gid
		}
		gr.Roles = append(gr.Roles, role)
		grs[GroupID(gid)] = gr
	}
	if rows.Err() != nil {
		return nil, err
	}

	return grs, nil
}

// AddGroupRole assigns the specified role to the given group, if it
// is not already assigned.
func (_AccessContexts) AddGroupRole(otx *sql.Tx, id AccessContextID, gid GroupID, rid RoleID) error {
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

	_, err := tx.Exec(`INSERT INTO wf_ac_group_roles(ac_id, group_id, role_id) VALUES(?, ?, ?)`, id, gid, rid)
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
func (_AccessContexts) RemoveGroupRole(otx *sql.Tx, id AccessContextID, gid GroupID, rid RoleID) error {
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

	_, err := tx.Exec(`DELETE FROM wf_ac_group_roles WHERE ac_id = ? AND group_id = ? AND role_id = ?`, id, gid, rid)
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

// Groups retrieves the users included in this access context.
func (_AccessContexts) Groups(id AccessContextID, offset, limit int64) (map[GroupID]*AcGroup, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.New("offset and limit should be non-negative integers")
	}
	if limit == 0 {
		limit = math.MaxInt64
	}

	q := `
	SELECT gm.id, gm.name, gm.group_type, auh.reports_to
	FROM wf_groups_master gm
	JOIN wf_ac_group_hierarchy auh ON auh.group_id = gm.id
	WHERE auh.ac_id = ?
	ORDER BY auh.group_id
	LIMIT ? OFFSET ?
	`
	rows, err := db.Query(q, id, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	gh := make(map[GroupID]*AcGroup)
	for rows.Next() {
		var g AcGroup
		err = rows.Scan(&g.ID, &g.Name, &g.GroupType, &g.ReportsTo)
		if err != nil {
			return nil, err
		}

		gh[GroupID(g.ID)] = &g
	}
	if rows.Err() != nil {
		return nil, err
	}

	return gh, nil
}

// AddGroup adds the given group to this access context, with the
// specified reporting authority within the hierarchy of this access
// context.
func (_AccessContexts) AddGroup(otx *sql.Tx, id AccessContextID, gid, reportsTo GroupID) error {
	if gid <= 0 || reportsTo < 0 {
		return errors.New("group ID should be a positive integer; reporting authority ID should be a non-negative integer")
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

	q := `INSERT INTO wf_ac_group_hierarchy(ac_id, group_id, reports_to) VALUES (?, ?, ?)`
	_, err := tx.Exec(q, id, gid, reportsTo)
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

// DeleteGroup removes the given group from this access context.
func (_AccessContexts) DeleteGroup(otx *sql.Tx, id AccessContextID, gid GroupID) error {
	if gid <= 0 {
		return errors.New("user ID should be positive integer")
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

	q := `DELETE FROM wf_ac_group_hierarchy WHERE ac_id = ? AND group_id = ?`
	_, err := tx.Exec(q, id, gid)
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

// GroupReportsTo answers the group to whom the given group reports to,
// within this access context.
func (_AccessContexts) GroupReportsTo(id AccessContextID, uid GroupID) (GroupID, error) {
	q := `
	SELECT reports_to
	FROM wf_ac_group_hierarchy
	WHERE ac_id = ?
	AND group_id = ?
	`
	row := db.QueryRow(q, id, uid)
	var repID int64
	err := row.Scan(&repID)
	if err != nil {
		return 0, err
	}

	return GroupID(repID), nil
}

// GroupReportees answers a list of the groups who report to the given
// group, within this access context.
func (_AccessContexts) GroupReportees(id AccessContextID, uid GroupID) ([]GroupID, error) {
	q := `
	SELECT group_id
	FROM wf_ac_group_hierarchy
	WHERE ac_id = ?
	AND reports_to = ?
	`
	rows, err := db.Query(q, id, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ary := make([]GroupID, 0, 4)
	for rows.Next() {
		var repID int64
		err = rows.Scan(&repID)
		if err != nil {
			return nil, err
		}
		ary = append(ary, GroupID(repID))
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ary, nil
}

// ChangeReporting reassigns the group to a different reporting
// authority.
func (_AccessContexts) ChangeReporting(otx *sql.Tx, id AccessContextID, gid, reportsTo GroupID) error {
	if gid <= 0 || reportsTo < 0 {
		return errors.New("group ID should be positive integer; reporting authority ID should be a non-negative integer")
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

	q := `
	UPDATE wf_ac_group_hierarchy
	SET reports_to = ?
	WHERE ac_id = ?
	AND group_id = ?
	`
	_, err := tx.Exec(q, reportsTo, id, gid)
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

// IncludesGroup answers `true` if the given group is included in this
// access context.
func (_AccessContexts) IncludesGroup(id AccessContextID, gid GroupID) (bool, error) {
	if gid <= 0 {
		return false, errors.New("group ID should be a positive integer")
	}

	q := `
	SELECT reports_to
	FROM wf_ac_group_hierarchy
	WHERE ac_id = ?
	AND group_id = ?
	`
	var repTo int64
	row := db.QueryRow(q, id, gid)
	err := row.Scan(&repTo)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// IncludesUser answers `true` if the given user is included in this
// access context.
func (_AccessContexts) IncludesUser(id AccessContextID, uid UserID) (bool, error) {
	if uid <= 0 {
		return false, errors.New("user ID should be a positive integer")
	}

	q := `
	SELECT COUNT(agh.reports_to)
	FROM wf_ac_group_hierarchy agh
	WHERE agh.ac_id = ?
	AND agh.group_id IN (
		SELECT gm.id
		FROM wf_groups_master gm
		JOIN wf_group_users gu ON gu.group_id = gm.id
		WHERE gu.user_id = ?
	)
	`
	var count int64
	row := db.QueryRow(q, id, uid)
	err := row.Scan(&count)
	if err != nil {
		return false, err
	}

	if count == 0 {
		return false, nil
	}

	return true, nil
}

// UserPermissions answers a list of the permissions available to the
// given user in this access context.
func (_AccessContexts) UserPermissions(id AccessContextID, uid UserID) (map[DocTypeID][]DocAction, error) {
	if uid <= 0 {
		return nil, errors.New("user ID should be a positive integer")
	}

	q := `
	SELECT acpv.doctype_id, acpv.docaction_id, dam.name
	FROM wf_ac_perms_v acpv
	JOIN wf_docactions_master dam ON dam.id = acpv.docaction_id
	WHERE acpv.ac_id = ?
	AND acpv.user_id = ?
	`
	rows, err := db.Query(q, id, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := map[DocTypeID][]DocAction{}
	for rows.Next() {
		var dtid int64
		var da DocAction
		err = rows.Scan(&dtid, &da.ID, &da.Name)
		if err != nil {
			return nil, err
		}

		ary, ok := res[DocTypeID(dtid)]
		if !ok {
			ary = []DocAction{}
		}
		ary = append(ary, da)
		res[DocTypeID(dtid)] = ary
	}
	if rows.Err() != nil {
		return nil, err
	}

	return res, nil
}

// UserPermissionsByDocType answers a list of the permissions
// available on the given document type, to the given user, in this
// access context.
func (_AccessContexts) UserPermissionsByDocType(id AccessContextID, dtype DocTypeID, uid UserID) ([]DocAction, error) {
	if id <= 0 || dtype <= 0 || uid <= 0 {
		return nil, errors.New("all identifiers should be positive integers")
	}

	q := `
	SELECT acpv.docaction_id, dam.name
	FROM wf_ac_perms_v acpv
	JOIN wf_docactions_master dam ON dam.id = acpv.docaction_id
	WHERE acpv.ac_id = ?
	AND acpv.doctype_id = ?
	AND acpv.user_id = ?
	`
	rows, err := db.Query(q, id, dtype, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := []DocAction{}
	for rows.Next() {
		var da DocAction
		err = rows.Scan(&da.ID, &da.Name)
		if err != nil {
			return nil, err
		}

		res = append(res, da)
	}
	if rows.Err() != nil {
		return nil, err
	}

	return res, nil
}

// GroupPermissions answers a list of the permissions available to the
// given user in this access context.
func (_AccessContexts) GroupPermissions(id AccessContextID, gid GroupID) (map[DocTypeID][]DocAction, error) {
	if gid <= 0 {
		return nil, errors.New("group ID should be a positive integer")
	}

	q := `
	SELECT acpv.doctype_id, acpv.docaction_id, dam.name
	FROM wf_ac_perms_v acpv
	JOIN wf_docactions_master dam ON dam.id = acpv.docaction_id
	WHERE acpv.ac_id = ?
	AND acpv.group_id = ?
	`
	rows, err := db.Query(q, id, gid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := map[DocTypeID][]DocAction{}
	for rows.Next() {
		var dtid int64
		var da DocAction
		err = rows.Scan(&dtid, &da.ID, &da.Name)
		if err != nil {
			return nil, err
		}

		ary, ok := res[DocTypeID(dtid)]
		if !ok {
			ary = []DocAction{}
		}
		ary = append(ary, da)
		res[DocTypeID(dtid)] = ary
	}
	if rows.Err() != nil {
		return nil, err
	}

	return res, nil
}

// GroupPermissionsByDocType answers a list of the permissions
// available on the given document type, to the given user, in this
// access context.
func (_AccessContexts) GroupPermissionsByDocType(id AccessContextID, dtype DocTypeID, gid GroupID) ([]DocAction, error) {
	if id <= 0 || dtype <= 0 || gid <= 0 {
		return nil, errors.New("all identifiers should be positive integers")
	}

	q := `
	SELECT acpv.docaction_id, dam.name
	FROM wf_ac_perms_v acpv
	JOIN wf_docactions_master dam ON dam.id = acpv.docaction_id
	WHERE acpv.ac_id = ?
	AND acpv.doctype_id = ?
	AND acpv.group_id = ?
	`
	rows, err := db.Query(q, id, dtype, gid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := []DocAction{}
	for rows.Next() {
		var da DocAction
		err = rows.Scan(&da.ID, &da.Name)
		if err != nil {
			return nil, err
		}

		res = append(res, da)
	}
	if rows.Err() != nil {
		return nil, err
	}

	return res, nil
}

// UserHasPermission answers `true` if the given user has the
// requested action enabled on the specified document type; `false`
// otherwise.
func (_AccessContexts) UserHasPermission(id AccessContextID, uid UserID, dtype DocTypeID, action DocActionID) (bool, error) {
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
	row := db.QueryRow(q, id, uid, dtype, action)
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
func (ac *AccessContext) GroupHasPermission(id AccessContextID, gid GroupID, dtype DocTypeID, action DocActionID) (bool, error) {
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
	row := db.QueryRow(q, id, gid, dtype, action)
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
