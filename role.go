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

// RoleID is the type of unique role identifiers.
type RoleID int64

// Role represents a collection of privileges.
//
// Each group in the system can have one or more roles assigned.
type Role struct {
	id   RoleID // globally-unique ID of this role
	name string // name of this role
}

// ID answers this role's identifier.
func (r *Role) ID() RoleID {
	return r.id
}

// Name answers this role's name.
func (r *Role) Name() string {
	return r.name
}

// Unexported type, only for convenience methods.
type _Roles struct{}

var _roles *_Roles

func init() {
	_roles = &_Roles{}
}

// Roles provides a resource-like interface to roles in the system.
func Roles() *_Roles {
	return _roles
}

// New creates a role with the given name.
func (rs *_Roles) New(otx *sql.Tx, name string) (RoleID, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, errors.New("name cannot not be empty")
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

	res, err := tx.Exec("INSERT INTO wf_roles_master(name) VALUES(?)", name)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	if otx == nil {
		err = tx.Commit()
		if err != nil {
			return 0, err
		}
	}

	return RoleID(id), nil
}

// List answers a subset of the roles, based on the input
// specification.
//
// Result set begins with ID >= `offset`, and has not more than
// `limit` elements.  A value of `0` for `offset` fetches from the
// beginning, while a value of `0` for `limit` fetches until the end.
func (rs *_Roles) List(offset, limit int64) ([]*Role, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.New("offset and limit must be non-negative integers")
	}
	if limit == 0 {
		limit = math.MaxInt64
	}

	q := `
	SELECT id, name
	FROM wf_roles_master
	ORDER BY id
	LIMIT ? OFFSET ?
	`
	rows, err := db.Query(q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ary := make([]*Role, 0, 10)
	for rows.Next() {
		var elem Role
		err = rows.Scan(&elem.id, &elem.name)
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

// Get loads the role object corresponding to the given role ID from
// the database, and answers that.
func (rs *_Roles) Get(id RoleID) (*Role, error) {
	if id <= 0 {
		return nil, errors.New("ID must be a positive integer")
	}

	var elem Role
	row := db.QueryRow("SELECT id, name FROM wf_roles_master WHERE id = ?", id)
	err := row.Scan(&elem.id, &elem.name)
	if err != nil {
		return nil, err
	}

	return &elem, nil
}

// GetByName answers the role, if one with the given name is
// registered; `nil` and the error, otherwise.
func (dts *_Roles) GetByName(name string) (*Role, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("role cannot be empty")
	}

	var elem Role
	row := db.QueryRow("SELECT id, name FROM wf_roles_master WHERE name = ?", name)
	err := row.Scan(&elem.id, &elem.name)
	if err != nil {
		return nil, err
	}

	return &elem, nil
}

// Rename renames the given role.
func (rs *_Roles) Rename(otx *sql.Tx, id RoleID, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("name cannot be empty")
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

	_, err := tx.Exec("UPDATE wf_roles_master SET name = ? WHERE id = ?", name, id)
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

// AddPermission adds the given action to this role, for the given
// document type.
func (rs *_Roles) AddPermission(otx *sql.Tx, rid RoleID, dtype DocTypeID, action DocActionID) error {
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
	INSERT INTO wf_role_docactions(role_id, doctype_id, docaction_id)
	VALUES(?, ?, ?)
	`
	_, err := tx.Exec(q, rid, dtype, action)
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

// RemovePermission removes all permissions from this role, for the
// given document type.
func (rs *_Roles) RemovePermission(otx *sql.Tx, rid RoleID, dtype DocTypeID, action DocActionID) error {
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
	DELETE FROM wf_role_docactions
	WHERE role_id = ?
	AND doctype_id = ?
	AND docaction_id = ?
	`
	_, err := tx.Exec(q, rid, dtype, action)
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

// Permissions answers the current set of permissions this role has.
// It answers `nil` in case the given document type does not have any
// permissions set in this role.
func (rs *_Roles) Permissions(rid RoleID) (map[DocType][]*DocAction, error) {
	q := `
	SELECT dtm.id, dtm.name, dam.id, dam.name
	FROM wf_doctypes_master dtm
	JOIN wf_role_docactions rdas ON dtm.id = rdas.doctype_id
	JOIN wf_docactions_master dam ON dam.id = rdas.docaction_id
	WHERE rdas.role_id = ?
	`
	rows, err := db.Query(q, rid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	das := make(map[DocType][]*DocAction)
	for rows.Next() {
		var dt DocType
		var da DocAction
		err = rows.Scan(&dt.id, &dt.name, &da.id, &da.name)
		if err != nil {
			return nil, err
		}
		ary, ok := das[dt]
		if !ok {
			ary = make([]*DocAction, 0, 1)
		}
		ary = append(ary, &da)
		das[dt] = ary
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return das, nil
}

// HasPermission answers `true` if this role has the queried
// permission for the given document type.
func (rs *_Roles) HasPermission(rid RoleID, dtype DocTypeID, action DocActionID) (bool, error) {
	q := `
	SELECT rdas.id FROM wf_role_docactions rdas
	JOIN wf_doctypes_master dtm ON rdas.doctype_id = dtm.id
	JOIN wf_docactions_master dam ON rdas.docaction_id = dam.id
	WHERE rdas.role_id = ?
	AND dtm.id = ?
	AND dam.id = ?
	ORDER BY rdas.id
	LIMIT 1
	`
	row := db.QueryRow(q, rid, dtype, action)
	var n int64
	err := row.Scan(&n)
	if err != nil {
		return false, err
	}

	return true, nil
}
