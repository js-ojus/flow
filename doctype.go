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
type DocType string

// Unexported type, only for convenience methods.
type _DocTypes struct{}

var _doctypes *_DocTypes

func init() {
	_doctypes = &_DocTypes{}
}

// DocTypes provides a resource-like interface to document types in
// the system.
func DocTypes() *_DocTypes {
	return _doctypes
}

// List answers a subset of the document types, based on the input
// specification.
//
// Result set begins with ID >= `offset`, and has not more than
// `limit` elements.  A value of `0` for `offset` fetches from the
// beginning, while a value of `0` for `limit` fetches until the end.
func (dts *_DocTypes) List(offset, limit int64) ([]DocType, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.New("offset and limit must be non-negative integers")
	}
	if limit == 0 {
		limit = math.MaxInt64
	}

	q := `
	SELECT name
	FROM wf_doctypes_master
	ORDER BY id
	LIMIT ? OFFSET ?
	`
	rows, err := db.Query(q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var name string
	dtary := make([]DocType, 0, 10)
	for rows.Next() {
		err = rows.Scan(&name)
		if err != nil {
			return nil, err
		}
		dtary = append(dtary, DocType(name))
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return dtary, nil
}

// New creates and registers a new document type in the system.
func (dts *_DocTypes) New(otx *sql.Tx, dt DocType) error {
	name := strings.TrimSpace(string(dt))
	if name == "" {
		return errors.New("document type cannot be empty")
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

	_, err := tx.Exec("INSERT INTO wf_doctypes_master(name) VALUES(?)", name)
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

// Exists answers `true` if a document type with the given name is
// registered; `false` otherwise.
func (dts *_DocTypes) Exists(dt DocType) (bool, error) {
	name := strings.TrimSpace(string(dt))
	if name == "" {
		return false, errors.New("document type cannot be empty")
	}

	row := db.QueryRow("SELECT COUNT(*) from wf_doctypes_master WHERE name = ?", name)
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
