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

// DocTypeID is the type of unique identifiers of document types.
type DocTypeID int64

// DocType enumerates the types of documents in the system, as defined
// by the consuming application.  Each document type has an associated
// workflow definition that drives its life cycle.
//
// Accordingly, `flow` does not assume anything about the specifics of
// the any document type.  Instead, it treats document types as plain,
// but controlled, vocabulary.  Nonetheless, it is highly recommended,
// but not necessary, that document types be defined in a system of
// hierarchical namespaces. For example:
//
//     PUR:RFQ
//
// could mean that the department is 'Purchasing', while the document
// type is 'Request For Quotation'.  As a variant,
//
//     PUR:ORD
//
// could mean that the document type is 'Purchase Order'.
//
// N.B. All document types must be defined as constant strings.
type DocType struct {
	ID   DocTypeID `json:"ID,omitempty"`   // Unique identifier of this document type
	Name string    `json:"Name,omitempty"` // Unique name of this document type
}

// Unexported type, only for convenience methods.
type _DocTypes struct{}

// DocTypes provides a resource-like interface to document types in
// the system.
var DocTypes _DocTypes

// docStorName answers the appropriate table name for the given
// document type.
func (_DocTypes) docStorName(dtid DocTypeID) string {
	return fmt.Sprintf("wf_documents_%03d", dtid)
}

// New creates and registers a new document type in the system.
func (_DocTypes) New(otx *sql.Tx, name string) (DocTypeID, error) {
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

	res, err := tx.Exec("INSERT INTO wf_doctypes_master(name) VALUES(?)", name)
	if err != nil {
		return 0, err
	}
	var id int64
	id, err = res.LastInsertId()
	if err != nil {
		return 0, err
	}

	tbl := DocTypes.docStorName(DocTypeID(id))
	q := `DROP TABLE IF EXISTS ` + tbl
	res, err = tx.Exec(q)
	if err != nil {
		return 0, err
	}
	q = `
	CREATE TABLE ` + tbl + ` (
		id INT NOT NULL AUTO_INCREMENT,
		path VARCHAR(1000) NOT NULL,
		ac_id INT NOT NULL,
		docstate_id INT NOT NULL,
		group_id INT NOT NULL,
		ctime TIMESTAMP NOT NULL,
		title VARCHAR(250) NULL,
		data BLOB NOT NULL,
		PRIMARY KEY (id),
		FOREIGN KEY (ac_id) REFERENCES wf_access_contexts(id),
		FOREIGN KEY (docstate_id) REFERENCES wf_docstates_master(id),
		FOREIGN KEY (group_id) REFERENCES wf_groups_master(id)
	)
	`
	res, err = tx.Exec(q)
	if err != nil {
		return 0, err
	}

	if otx == nil {
		err = tx.Commit()
		if err != nil {
			return 0, err
		}
	}
	return DocTypeID(id), nil
}

// List answers a subset of the document types, based on the input
// specification.
//
// Result set begins with ID >= `offset`, and has not more than
// `limit` elements.  A value of `0` for `offset` fetches from the
// beginning, while a value of `0` for `limit` fetches until the end.
func (_DocTypes) List(offset, limit int64) ([]*DocType, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.New("offset and limit must be non-negative integers")
	}
	if limit == 0 {
		limit = math.MaxInt64
	}

	q := `
	SELECT id, name
	FROM wf_doctypes_master
	ORDER BY id
	LIMIT ? OFFSET ?
	`
	rows, err := db.Query(q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ary := make([]*DocType, 0, 10)
	for rows.Next() {
		var elem DocType
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

// Get retrieves the document type for the given ID.
func (_DocTypes) Get(id DocTypeID) (*DocType, error) {
	if id <= 0 {
		return nil, errors.New("ID should be a positive integer")
	}

	var elem DocType
	row := db.QueryRow("SELECT id, name FROM wf_doctypes_master WHERE id = ?", id)
	err := row.Scan(&elem.ID, &elem.Name)
	if err != nil {
		return nil, err
	}

	return &elem, nil
}

// GetByName answers the document type, if one with the given name is
// registered; `nil` and the error, otherwise.
func (_DocTypes) GetByName(name string) (*DocType, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("document type cannot be empty")
	}

	var elem DocType
	row := db.QueryRow("SELECT id, name FROM wf_doctypes_master WHERE name = ?", name)
	err := row.Scan(&elem.ID, &elem.Name)
	if err != nil {
		return nil, err
	}

	return &elem, nil
}

// Rename renames the given document type.
func (_DocTypes) Rename(otx *sql.Tx, id DocTypeID, name string) error {
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

	_, err := tx.Exec("UPDATE wf_doctypes_master SET name = ? WHERE id = ?", name, id)
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

// Transition holds the information of which action results in which
// state.
type Transition struct {
	Upon DocAction // If user/system has performed this action
	To   DocState  // Document transitions into this state
}

// TransitionMap holds the state transitions defined for this document
// type.  It lays out which actions result in which target states,
// given current states.
type TransitionMap struct {
	From        DocState // When document is in this state
	Transitions map[DocActionID]Transition
}

// Transitions answers the possible document states into which a
// document currently in the given state can transition.
func (_DocTypes) Transitions(dtype DocTypeID) (map[DocStateID]*TransitionMap, error) {
	q := `
	SELECT dst.from_state_id, dsm1.name, dst.docaction_id, dam.name, dst.to_state_id, dsm2.name
	FROM wf_docstate_transitions dst
	JOIN wf_docstates_master dsm1 ON dsm1.id = dst.from_state_id
	JOIN wf_docstates_master dsm2 ON dsm2.id = dst.to_state_id
	JOIN wf_docactions_master dam ON dam.id = dst.docaction_id
	WHERE dst.doctype_id = ?
	`
	rows, err := db.Query(q, dtype)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := map[DocStateID]*TransitionMap{}
	for rows.Next() {
		var dsfrom DocState
		var t Transition
		err := rows.Scan(&dsfrom.ID, &dsfrom.Name, &t.Upon.ID, &t.Upon.Name, &t.To.ID, &t.To.Name)
		if err != nil {
			return nil, err
		}

		var elem *TransitionMap
		ok := false
		if elem, ok = res[dsfrom.ID]; !ok {
			elem = &TransitionMap{}
			elem.From = dsfrom
			elem.Transitions = map[DocActionID]Transition{}
		}

		elem.Transitions[t.Upon.ID] = t
		res[dsfrom.ID] = elem
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

// _Transitions answers the possible document states into which a
// document currently in the given state can transition.  Only
// identifiers are answered in the map.
func (_DocTypes) _Transitions(dtype DocTypeID, state DocStateID) (map[DocActionID]DocStateID, error) {
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
func (_DocTypes) AddTransition(otx *sql.Tx, dtype DocTypeID, state DocStateID,
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
func (_DocTypes) RemoveTransition(otx *sql.Tx, dtype DocTypeID, state DocStateID, action DocActionID) error {
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
	`
	_, err := tx.Exec(q, dtype, state, action)
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
