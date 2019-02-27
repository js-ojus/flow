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

// DocActionID is the type of unique identifiers of document actions.
type DocActionID int64

// DocAction enumerates the types of actions in the system, as defined
// by the consuming application.  Each document action has an
// associated workflow transition that it causes.
//
// Accordingly, `flow` does not assume anything about the specifics of
// the any document action.  Instead, it treats document actions as plain,
// but controlled, vocabulary.  Good examples include:
//
//     CREATE,
//     REVIEW,
//     APPROVE,
//     REJECT,
//     COMMENT,
//     CLOSE, and
//     REOPEN.
//
// N.B. All document actions must be defined as constant strings.
type DocAction struct {
	ID        DocActionID `json:"ID"`        // Unique identifier of this action
	Name      string      `json:"Name"`      // Globally-unique name of this action
	Reconfirm bool        `json:"Reconfirm"` // Should the user be prompted for a reconfirmation of this action?
}

// Unexported type, only for convenience methods.
type _DocActions struct{}

// DocActions provides a resource-like interface to document actions
// in the system.
var DocActions _DocActions

// New creates and registers a new document action in the system.
func (_DocActions) New(otx *sql.Tx, name string, reconfirm bool) (DocActionID, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, errors.New("document action cannot be empty")
	}

	var tx *sql.Tx
	var err error
	if otx == nil {
		tx, err = db.Begin()
		if err != nil {
			return 0, err
		}
		defer tx.Rollback()
	} else {
		tx = otx
	}

	var res sql.Result
	if reconfirm {
		res, err = tx.Exec("INSERT INTO wf_docactions_master(name, reconfirm) VALUES(?, ?)", name, 1)
	} else {
		res, err = tx.Exec("INSERT INTO wf_docactions_master(name, reconfirm) VALUES(?, ?)", name, 0)
	}
	if err != nil {
		return 0, err
	}
	var aid int64
	aid, err = res.LastInsertId()
	if err != nil {
		return 0, err
	}

	if otx == nil {
		err = tx.Commit()
		if err != nil {
			return 0, err
		}
	}

	return DocActionID(aid), nil
}

// List answers a subset of the document actions, based on the input
// specification.
//
// Result set begins with ID >= `offset`, and has not more than
// `limit` elements.  A value of `0` for `offset` fetches from the
// beginning, while a value of `0` for `limit` fetches until the end.
func (_DocActions) List(offset, limit int64) ([]*DocAction, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.New("offset and limit must be non-negative integers")
	}
	if limit == 0 {
		limit = math.MaxInt64
	}

	q := `
	SELECT id, name, reconfirm
	FROM wf_docactions_master
	ORDER BY id
	LIMIT ? OFFSET ?
	`
	rows, err := db.Query(q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ary := make([]*DocAction, 0, 10)
	for rows.Next() {
		var elem DocAction
		err = rows.Scan(&elem.ID, &elem.Name, &elem.Reconfirm)
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

// Get retrieves the document action for the given ID.
func (_DocActions) Get(id DocActionID) (*DocAction, error) {
	if id <= 0 {
		return nil, errors.New("ID should be a positive integer")
	}

	var elem DocAction
	row := db.QueryRow("SELECT id, name, reconfirm FROM wf_docactions_master WHERE id = ?", id)
	err := row.Scan(&elem.ID, &elem.Name, &elem.Reconfirm)
	if err != nil {
		return nil, err
	}

	return &elem, nil
}

// GetByName answers the document action, if one such with the given
// name is registered; `nil` and the error, otherwise.
func (_DocActions) GetByName(name string) (*DocAction, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("document action cannot be empty")
	}

	var elem DocAction
	row := db.QueryRow("SELECT id, name, reconfirm FROM wf_docactions_master WHERE name = ?", name)
	err := row.Scan(&elem.ID, &elem.Name, &elem.Reconfirm)
	if err != nil {
		return nil, err
	}

	return &elem, nil
}

// Rename renames the given document action.
func (_DocActions) Rename(otx *sql.Tx, id DocActionID, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("name cannot be empty")
	}

	var tx *sql.Tx
	var err error
	if otx == nil {
		tx, err = db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()
	} else {
		tx = otx
	}

	_, err = tx.Exec("UPDATE wf_docactions_master SET name = ? WHERE id = ?", name, id)
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
