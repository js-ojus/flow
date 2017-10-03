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
	ID   DocTypeID `json:"ID"`   // Unique identifier of this document type
	Name string    `json:"Name"` // Unique name of this document type
}

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

// docStorName answers the appropriate table name for the given
// document type.
func (dts *_DocTypes) docStorName(dtid DocTypeID) string {
	return fmt.Sprintf("wf_documents_%03d", dtid)
}

// New creates and registers a new document type in the system.
func (dts *_DocTypes) New(otx *sql.Tx, name string) (DocTypeID, error) {
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

	tbl := _doctypes.docStorName(DocTypeID(id))
	q := `DROP TABLE IF EXISTS ` + tbl
	res, err = tx.Exec(q)
	if err != nil {
		return 0, err
	}
	q = `
	CREATE TABLE ` + tbl + ` (
		id INT NOT NULL AUTO_INCREMENT,
		user_id INT NOT NULL,
		docstate_id INT NOT NULL,
		ctime TIMESTAMP NOT NULL,
		title VARCHAR(250) NOT NULL,
		data BLOB NOT NULL,
		PRIMARY KEY (id),
		FOREIGN KEY (docstate_id) REFERENCES wf_docstates_master(id)
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
func (dts *_DocTypes) List(offset, limit int64) ([]*DocType, error) {
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
func (dts *_DocTypes) Get(id DocTypeID) (*DocType, error) {
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
func (dts *_DocTypes) GetByName(name string) (*DocType, error) {
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
func (dts *_DocTypes) Rename(otx *sql.Tx, id DocTypeID, name string) error {
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
