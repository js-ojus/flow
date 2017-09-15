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
	id    DocStateID // Unique identifier of this document state
	dtype DocType    // For namespace purposes
	name  string     // Unique identifier of this state in its workflow
}

// ID answers the unique identifier of this document state.
func (s *DocState) ID() DocStateID {
	return s.id
}

// Type answers the document type for which this defines a state.
func (s *DocState) Type() DocType {
	return s.dtype
}

// Name answers this state's name.
func (s *DocState) Name() string {
	return s.name
}

// Unexported type, only for convenience methods.
type _DocStates struct{}

var _docstates *_DocStates

func init() {
	_docstates = &_DocStates{}
}

// DocStates provides a resource-like interface to document actions
// in the system.
func DocStates() *_DocStates {
	return _docstates
}

// New creates an enumerated state as defined by the consuming
// application.
func (dss *_DocStates) New(otx *sql.Tx, dtype *DocType, name string) (DocStateID, error) {
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

	res, err := tx.Exec("INSERT INTO wf_docstates_master(doctype_id, name) VALUES(?, ?)", dtype.id, name)
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
	SELECT *
	FROM wf_docstates_master
	ORDER BY id
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
		err = rows.Scan(&elem.id, &elem.dtype, &elem.name)
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
	row := db.QueryRow("SELECT id, doctype_id, name FROM wf_docstates_master WHERE id = ?", id)
	err := row.Scan(&elem.id, &elem.dtype, &elem.name)
	if err != nil {
		return nil, err
	}

	return &elem, nil
}

// Rename renames the given document state.
func (dss *_DocStates) Rename(otx *sql.Tx, elem *DocState, name string) error {
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

	_, err := tx.Exec("UPDATE wf_docstates_master SET name = ? WHERE id = ?", name, elem.id)
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

// Exists answers its unique ID, if a document state with the given
// name is registered; `0` and the error, otherwise.
func (dss *_DocStates) Exists(dtype *DocType, name string) (DocStateID, error) {
	if dtype == nil {
		return 0, errors.New("document type should not be `nil`")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, errors.New("document state cannot be empty")
	}

	row := db.QueryRow("SELECT id FROM wf_docstates_master WHERE doctype_id = ? AND name = ?", dtype.id, name)
	var n int64
	err := row.Scan(&n)
	if err != nil {
		return 0, err
	}

	return DocStateID(n), nil
}
