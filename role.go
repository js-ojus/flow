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

// NewRole creates a role with the given name.
func NewRole(otx *sql.Tx, name string) (RoleID, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, errors.New("role name cannot not be empty")
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
	rid, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	if otx == nil {
		err = tx.Commit()
		if err != nil {
			return 0, err
		}
	}

	return RoleID(rid), nil
}

// GetRole loads the role object corresponding to the given role ID
// from the database, and answers that.
func GetRole(rid RoleID) (*Role, error) {
	if rid <= 0 {
		return nil, errors.New("role ID must be a positive integer")
	}

	row := db.QueryRow("SELECT name FROM wf_roles_master WHERE id = ?", rid)
	var name string
	err := row.Scan(&name)
	if err != nil {
		return nil, err
	}

	r := &Role{id: rid, name: name}
	return r, nil
}

// ID answers this role's identifier.
func (r *Role) ID() RoleID {
	return r.id
}

// Name answers this role's name.
func (r *Role) Name() string {
	return r.name
}

// AddPermission adds the given action to this role, for the given
// document type.
func (r *Role) AddPermission(otx *sql.Tx, dt DocType, da DocAction) error {
	tdt := strings.TrimSpace(string(dt))
	tda := strings.TrimSpace(string(da))
	if tdt == "" || tda == "" {
		return errors.New("document type and document action cannot be empty")
	}

	var dtid, daid int64
	row := db.QueryRow("SELECT id FROM wf_doctypes_master WHERE name = ?", tdt)
	err := row.Scan(&dtid)
	if err != nil {
		return err
	}
	row = db.QueryRow("SELECT id FROM wf_docactions_master WHERE name = ?", tda)
	err = row.Scan(&daid)
	if err != nil {
		return err
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
	INSERT INTO wf_role_docactions(role_id, doctype_id, docaction_id)
	VALUES(?, ?, ?)
	`
	_, err = tx.Exec(q, r.id, dtid, daid)
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
func (r *Role) RemovePermission(otx *sql.Tx, dt DocType, da DocAction) error {
	tdt := strings.TrimSpace(string(dt))
	tda := strings.TrimSpace(string(da))
	if tdt == "" || tda == "" {
		return errors.New("document type and document action cannot be empty")
	}

	var dtid, daid int64
	row := db.QueryRow("SELECT id FROM wf_doctypes_master WHERE name = ?", tdt)
	err := row.Scan(&dtid)
	if err != nil {
		return err
	}
	row = db.QueryRow("SELECT id FROM wf_docactions_master WHERE name = ?", tda)
	err = row.Scan(&daid)
	if err != nil {
		return err
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
	DELETE FROM wf_role_docactions
	WHERE role_id = ?
	AND doctype_id = ?
	AND docaction_id = ?
	`
	_, err = tx.Exec(q, r.id, dtid, daid)
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
func (r *Role) Permissions() (map[DocType][]DocAction, error) {
	q := `
	SELECT dtm.name, dam.name
	FROM wf_doctypes_master dtm, wf_docactions_master dam
	JOIN wf_role_docactions rdas ON dtm.id = rdas.doctype_id
	JOIN ON dam.id = rdas.docaction_id
	WHERE rdas.role_id = ?
	`
	rows, err := db.Query(q, r.id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	das := make(map[DocType][]DocAction)
	var dt DocType
	var da DocAction
	for rows.Next() {
		err = rows.Scan(&dt, &da)
		if err != nil {
			return nil, err
		}
		sl, ok := das[dt]
		if !ok {
			sl = make([]DocAction, 0, 1)
		}
		sl = append(sl, da)
		das[dt] = sl
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return das, nil
}

// HasPermission answers `true` if this role has the queried
// permission for the given document type.
func (r *Role) HasPermission(dt DocType, da DocAction) (bool, error) {
	q := `
	SELECT COUNT(*) FROM wf_role_docactions rdas
	JOIN wf_doctypes_master dtm ON rdas.doctype_id = dtm.id
	JOIN wf_docactions_master dam ON rdas.docaction_id = dam.id
	WHERE rdas.role_id = ?
	AND dtm.name = ?
	AND dam.name = ?
	`
	row := db.QueryRow(q, r.id, dt, da)
	var n int64
	err := row.Scan(&n)
	if err != nil {
		return false, err
	}
	if n == 0 {
		return false, nil
	}

	return true, nil
}
