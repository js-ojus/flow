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

// DocStateID is the type of unique identifiers of document states.
type DocStateID int64

// DocState is one of a set of enumerated states for a top-level
// document, as defined by the consuming application.
//
// `flow`, therefore, does not assume anything about the specifics of
// any state.  Applications can, for example, embed `DocState` in a
// struct that provides more context.  Document states should be
// loaded during application initialisation.
//
// N.B. A `DocState` once defined and used, should *NEVER* be removed.
// At best, it can be deprecated by defining a new one, and then
// altering the corresponding workflow definition to use the new one
// instead.
type DocState struct {
	ID      DocStateID `json:"ID,omitempty"`   // Unique identifier of this document state
	DocType DocType    `json:"DocType"`        // For namespace purposes
	Name    string     `json:"Name,omitempty"` // Unique identifier of this state in its workflow
}

// Unexported type, only for convenience methods.
type _DocStates struct{}

// DocStates provides a resource-like interface to document actions
// in the system.
var DocStates *_DocStates

// New creates an enumerated state as defined by the consuming
// application.
func (dss *_DocStates) New(otx *sql.Tx, dtype DocTypeID, name string) (DocStateID, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, errors.New("name cannot be empty")
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

	res, err := tx.Exec("INSERT INTO wf_docstates_master(doctype_id, name) VALUES(?, ?)", dtype, name)
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

	return DocStateID(id), nil
}

// List answers a subset of the document states, based on the input
// specification.
//
// Result set begins with ID >= `offset`, and has not more than
// `limit` elements.  A value of `0` for `offset` fetches from the
// beginning, while a value of `0` for `limit` fetches until the end.
func (dss *_DocStates) List(offset, limit int64) ([]*DocState, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.New("offset and limit must be non-negative integers")
	}
	if limit == 0 {
		limit = math.MaxInt64
	}

	q := `
	SELECT dsm.id, dtm.id, dtm.name, dsm.name
	FROM wf_docstates_master dsm
	JOIN wf_doctypes_master dtm ON dsm.doctype_id = dtm.id
	ORDER BY dsm.id
	LIMIT ? OFFSET ?
	`
	rows, err := db.Query(q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ary := make([]*DocState, 0, 10)
	for rows.Next() {
		var elem DocState
		err = rows.Scan(&elem.ID, &elem.DocType.ID, &elem.DocType.Name, &elem.Name)
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

// Get retrieves the document state for the given ID.
func (dss *_DocStates) Get(id DocStateID) (*DocState, error) {
	if id <= 0 {
		return nil, errors.New("ID should be a positive integer")
	}

	var elem DocState
	q := `
	SELECT dsm.id, dtm.id, dtm.name, dsm.name
	FROM wf_docstates_master dsm
	JOIN wf_doctypes_master dtm ON dsm.doctype_id = dtm.id
	WHERE dsm.id = ?
	`
	row := db.QueryRow(q, id)
	err := row.Scan(&elem.ID, &elem.DocType.ID, &elem.DocType.Name, &elem.Name)
	if err != nil {
		return nil, err
	}

	return &elem, nil
}

// GetByName answers the document state, if one with the given name is
// registered; `nil` and the error, otherwise.
func (dss *_DocStates) GetByName(dtype DocTypeID, name string) (*DocState, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("document state cannot be empty")
	}

	var elem DocState
	row := db.QueryRow("SELECT id, doctype_id, name FROM wf_docstates_master WHERE doctype_id = ? AND name = ?", dtype, name)
	err := row.Scan(&elem.ID, &elem.DocType.ID, &elem.Name)
	if err != nil {
		return nil, err
	}

	return &elem, nil
}

// Rename renames the given document state.
func (dss *_DocStates) Rename(otx *sql.Tx, id DocStateID, name string) error {
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

	_, err := tx.Exec("UPDATE wf_docstates_master SET name = ? WHERE id = ?", name, id)
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

// Transitions answers the possible document states into which a
// document currently in the given state can transition.
func (dss *_DocStates) Transitions(dtype DocTypeID, state DocStateID) (map[DocAction]DocState, error) {
	q := `
	SELECT dst.docaction_id, dam.name, dst.to_state_id, dsm.name, dtm.name
	FROM wf_docstate_transitions dst
	JOIN wf_docstates_master dsm ON dst.to_state_id = dsm.id
	JOIN wf_docactions_master dam ON dst.docaction_id = dam.id
	JOIN wf_doctypes_master dtm ON dst.doctype_id = dtm.id
	WHERE dst.doctype_id = ?
	AND dst.from_state_id = ?
	`
	rows, err := db.Query(q, dtype, state)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hash := make(map[DocAction]DocState)
	for rows.Next() {
		var da DocAction
		var ds DocState
		err := rows.Scan(&da.ID, &da.Name, &ds.ID, &ds.Name, &ds.DocType.Name)
		if err != nil {
			return nil, err
		}
		ds.DocType.ID = dtype
		hash[da] = ds
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return hash, nil
}

// _Transitions answers the possible document states into which a
// document currently in the given state can transition.  Only
// identifiers are answered in the map.
func (dss *_DocStates) _Transitions(dtype DocTypeID, state DocStateID) (map[DocActionID]DocStateID, error) {
	q := `
	SELECT docaction_id, to_state_id
	FROM wf_docstate_transitions
	WHERE doctype_id = ?
	AND from_state_id = ?
	`
	rows, err := db.Query(q, dtype, state)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hash := make(map[DocActionID]DocStateID)
	for rows.Next() {
		var da DocActionID
		var ds DocStateID
		err := rows.Scan(&da, &ds)
		if err != nil {
			return nil, err
		}
		hash[da] = ds
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return hash, nil
}

// AddTransition associates a target document state with a document
// action performed on documents in the given current state.
func (dss *_DocStates) AddTransition(otx *sql.Tx, dtype DocTypeID, state DocStateID,
	action DocActionID, toState DocStateID) error {
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
	INSERT INTO wf_docstate_transitions(doctype_id, from_state_id, docaction_id, to_state_id)
	VALUES(?, ?, ?, ?)
	`
	_, err := tx.Exec(q, dtype, state, action, toState)
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

// RemoveTransition disassociates a target document state with a
// document action performed on documents in the given current state.
func (dss *_DocStates) RemoveTransition(otx *sql.Tx, dtype DocTypeID, state DocStateID,
	action DocActionID, toState DocStateID) error {
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
	DELETE FROM wf_docstate_transitions
	WHERE doctype_id = ?
	AND from_state_id =?
	AND docaction_id = ?
	AND to_state_id = ?
	`
	_, err := tx.Exec(q, dtype, state, action, toState)
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
