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
	"math"
	"strings"
)

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
type DocAction string

// Unexported type, only for convenience methods.
type _DocActions struct{}

var _docactions *_DocActions

func init() {
	_docactions = &_DocActions{}
}

// DocActions provides a resource-like interface to document actions
// in the system.
func DocActions() *_DocActions {
	return _docactions
}

// List answers a subset of the document actions, based on the input
// specification.
//
// Result set begins with ID >= `offset`, and has not more than
// `limit` elements.  A value of `0` for `offset` fetches from the
// beginning, while a value of `0` for `limit` fetches until the end.
func (das *_DocActions) List(offset, limit int64) ([]DocAction, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.New("offset and limit must be non-negative integers")
	}
	if limit == 0 {
		limit = math.MaxInt64
	}

	q := `
	SELECT name
	FROM wf_docactions_master
	ORDER BY id
	LIMIT ? OFFSET ?
	`
	rows, err := db.Query(q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var name string
	daary := make([]DocAction, 0, 10)
	for rows.Next() {
		err = rows.Scan(&name)
		if err != nil {
			return nil, err
		}
		daary = append(daary, DocAction(name))
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return daary, nil
}

// New creates and registers a new document action in the system.
func (das *_DocActions) New(otx *sql.Tx, da DocAction) error {
	name := strings.TrimSpace(string(da))
	if name == "" {
		return errors.New("document action cannot be empty")
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

	_, err := tx.Exec("INSERT INTO wf_docactions_master(name) VALUES(?)", name)
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

// Exists answers `true` if a document action with the given name is
// registered; `false` otherwise.
func (das *_DocActions) Exists(da DocAction) (bool, error) {
	name := strings.TrimSpace(string(da))
	if name == "" {
		return false, errors.New("document action cannot be empty")
	}

	row := db.QueryRow("SELECT COUNT(*) from wf_docactions_master WHERE name = ?", name)
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
