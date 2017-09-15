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
	"crypto/sha1"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"
)

// Blob is a simple data holder for information concerning the
// user-supplied name of the binary object, the path of the stored
// binary object, and its SHA1 checksum.
type Blob struct {
	Name    string // User-given name to the binary object
	Path    string // Path to the stored binary object
	Sha1Sum string // SHA1 checksum of the binary object
}

// DocumentID is the type of unique document identifiers.
type DocumentID int64

// Document represents a task in a workflow, whose life cycle it
// tracks.
//
// Documents are central to the workflow engine and its operations. In
// the process, it accumulates various details, and tracks the times
// of its modifications.  The life cycle typically involves several
// state transitions, whose details are also tracked.
//
// `Document` is a recursive structure: it can contain other
// documents.  Most applications should embed `Document` in their
// document structures rather than use this directly.
type Document struct {
	dtype DocType    // For namespacing
	id    DocumentID // Globally-unique identifier of this document

	user  UserID    // Creator of this document
	state DocState  // Current state
	ctime time.Time // Creation time of this revision

	title string // Human-readable title; applicable only for top-level documents
	data  []byte // Primary content of the document
}

// Type answers this document's type.
func (d *Document) Type() DocType {
	return d.dtype
}

// ID answers this document's globally-unique ID.
func (d *Document) ID() DocumentID {
	return d.id
}

// User answers the user who created this document.
func (d *Document) User() UserID {
	return d.user
}

// Ctime answers the creation time of this document.
func (d *Document) Ctime() time.Time {
	return d.ctime
}

// State answer this document's current state.
func (d *Document) State() DocState {
	return d.state
}

// Title answers this document's title.
func (d *Document) Title() string {
	return d.title
}

// Data answers this document's primary data content.
func (d *Document) Data() []byte {
	dt := make([]byte, len(d.data))
	copy(dt, d.data)
	return dt
}

// Unexported type, only for convenience methods.
type _Documents struct{}

var _documents *_Documents

func init() {
	_documents = &_Documents{}
}

// New creates and initialises a document.
//
// The document created through this method has a life cycle that is
// associated with it through a particular workflow.
//
// N.B. Blobs, tags and children documents have to be associated with
// this document, if needed, through appropriate separate calls.
func (ds *_Documents) New(otx *sql.Tx, user UserID, dtype DocTypeID, otype DocTypeID, oid DocumentID,
	state DocStateID, title string, data []byte) (DocumentID, error) {
	if user <= 0 {
		return 0, errors.New("user ID must be a positive integer")
	}

	// A child document does not have its own title.
	if oid > 0 {
		title = ChildDocTitle
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

	tbl := _doctypes.docStorName(dtype)
	q := `INSERT INTO ` + tbl + `(user_id, docstate_id, ctime, title, data)
	VALUES (?, ?, NOW(), ?, ?)
	`
	res, err := tx.Exec(q, user, state, title, data)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	if oid > 0 {
		q2 := `
		INSERT INTO wf_document_children(parent_doctype_id, parent_id, child_doctype_id, child_id)
		VALUES (?, ?, ?, ?)
		`
		res, err = tx.Exec(q2, otype, oid, dtype, id)
		if err != nil {
			return 0, err
		}
	}

	if otx == nil {
		err = tx.Commit()
		if err != nil {
			return 0, err
		}
	}

	return DocumentID(id), nil
}

// Get initialises a document by reading from the database.
//
// N.B. This retrieves the primary data of the document.  Other
// information viz. blobs, tags and children documents have to be
// fetched separately.
func (ds *_Documents) Get(dtype DocTypeID, id DocumentID) (*Document, error) {
	tbl := _doctypes.docStorName(dtype)
	var d Document
	q := `
	SELECT docs.user_id, docs.docstate_id, docs.ctime, docs.title, docs.data, states.name
	FROM ` + tbl + ` AS docs
	JOIN wf_docstates_master AS states ON docs.docstate_id = states.id
	WHERE docs.id = ?
	`
	row := db.QueryRow(q, id, dtype)
	err := row.Scan(&d.user, &d.state.id, &d.ctime, &d.title, &d.data, &d.state.name)
	if err != nil {
		return nil, err
	}
	q = `SELECT name FROM wf_doctypes_master WHERE id = ?`
	row = db.QueryRow(q, dtype)
	err = row.Scan(&d.dtype.name)
	if err != nil {
		return nil, err
	}

	d.id = id
	d.dtype.id = dtype
	return &d, nil
}

// SetTitle sets the title of the document.
func (ds *_Documents) SetTitle(otx *sql.Tx, user UserID, dtype DocTypeID, id DocumentID, title string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return errors.New("document title should not be empty")
	}

	// A child document does not have its own title.

	q := `
	SELECT child_id
	FROM wf_document_children
	WHERE child_doctype_id = ?
	AND child_id = ?
	ORDER BY child_id
	LIMIT 1
	`
	var tid int64
	row := db.QueryRow(q, dtype, id)
	err := row.Scan(&tid)
	if err == nil {
		return errors.New("a child document cannot have its own title")
	}

	tbl := _doctypes.docStorName(dtype)
	var duser UserID
	q = `SELECT user FROM ` + tbl + ` WHERE id = ?`
	row = db.QueryRow(q, id)
	err = row.Scan(&duser)
	if err != nil {
		return err
	}
	if duser != user {
		return fmt.Errorf("author mismatch -- original user : %d, current user : %d", duser, user)
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

	q = `UPDATE ` + tbl + ` SET title = ? WHERE doc_id = ?`
	_, err = tx.Exec(q, title, id)
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

// SetData sets the data of the document.
func (ds *_Documents) SetData(otx *sql.Tx, user UserID, dtype DocTypeID, id DocumentID, data []byte) error {
	if data == nil {
		return errors.New("document data should not be empty")
	}

	tbl := _doctypes.docStorName(dtype)
	var duser UserID
	q := `SELECT user FROM ` + tbl + ` WHERE id = ?`
	row := db.QueryRow(q, id)
	err := row.Scan(&duser)
	if err != nil {
		return err
	}
	if duser != user {
		return fmt.Errorf("author mismatch -- original user : %d, current user : %d", duser, user)
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

	q = `UPDATE ` + tbl + ` SET data = ? WHERE doc_id = ?`
	_, err = tx.Exec(q, data, id)
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

// Blobs answers a list of this document's enclosures (as paths, not
// the actual blobs).
func (ds *_Documents) Blobs(dtype DocTypeID, id DocumentID) ([]*Blob, error) {
	bs := make([]*Blob, 0, 1)
	q := `
	SELECT name, path, sha1sum
	FROM wf_document_blobs
	WHERE doctype_id = ?
	AND doc_id = ?
	`
	rows, err := db.Query(q, dtype, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var b Blob
		err = rows.Scan(&b.Name, &b.Path, &b.Sha1Sum)
		if err != nil {
			return nil, err
		}
		bs = append(bs, &b)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return bs, nil
}

// AddBlob adds the path to an enclosure to this document.
func (ds *_Documents) AddBlob(otx *sql.Tx, user UserID, dtype DocTypeID, id DocumentID, blob *Blob) error {
	if blob == nil {
		return errors.New("blob should be non-nil")
	}

	tbl := _doctypes.docStorName(dtype)
	var duser UserID
	q := `SELECT user FROM ` + tbl + ` WHERE id = ?`
	row := db.QueryRow(q, id)
	err := row.Scan(&duser)
	if err != nil {
		return err
	}
	if duser != user {
		return fmt.Errorf("author mismatch -- original user : %d, current user : %d", duser, user)
	}

	// Verify the given checksum.

	f, err := os.Open(blob.Path)
	if err != nil {
		return err
	}
	defer f.Close()
	h := sha1.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return err
	}
	csum := fmt.Sprintf("%x", h.Sum(nil))
	if blob.Sha1Sum != csum {
		return fmt.Errorf("checksum mismatch -- given SHA1 sum : %s, computed SHA1 sum : %s", blob.Sha1Sum, csum)
	}

	// Store the blob in the appropriate path.

	success := false
	bpath := path.Join(blobsDir, csum[0:2], csum)
	err = os.Rename(blob.Path, bpath)
	if err != nil {
		return err
	}
	// Clean-up in case of any error.  However, this mechanism is not
	// adequate if this method runs in the scope of an outer
	// transaction.  The moved file will be orphaned, should the outer
	// transaction abort later.
	defer func() {
		if !success {
			os.Remove(bpath)
		}
	}()

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

	// Now write the database entry.

	q = `
	INSERT INTO wf_document_blobs(doctype_id, doc_id, name, path, sha1sum)
	VALUES(?, ?, ?, ?, ?)
	`
	_, err = tx.Exec(q, dtype, id, blob.Name, bpath, csum)
	if err != nil {
		return err
	}

	if otx == nil {
		err = tx.Commit()
		if err != nil {
			return err
		}
	}

	success = true
	return nil
}

// AddTag associates the given tag with this document.
//
// Tags are converted to lower case (as per normal Unicode casing)
// before getting associated with documents.  Also, embedded spaces,
// if any, are retained.
func (ds *_Documents) AddTag(otx *sql.Tx, user UserID, dtype DocTypeID, id DocumentID, tag string) error {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return errors.New("tag should not be empty")
	}
	tag = strings.ToLower(tag)

	tbl := _doctypes.docStorName(dtype)
	var duser UserID
	q := `SELECT user FROM ` + tbl + ` WHERE id = ?`
	row := db.QueryRow(q, id)
	err := row.Scan(&duser)
	if err != nil {
		return err
	}
	if duser != user {
		return fmt.Errorf("author mismatch -- original user : %d, current user : %d", duser, user)
	}

	// A child document does not have its own tags.

	q = `
	SELECT child_id
	FROM wf_document_children
	WHERE child_doctype_id = ?
	AND child_id = ?
	ORDER BY child_id
	LIMIT 1
	`
	var tid int64
	row = db.QueryRow(q, dtype, id)
	err = row.Scan(&tid)
	if err == nil {
		return errors.New("a child document cannot have its own tags")
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

	// Now write the database entry.

	q = `
	INSERT INTO wf_document_tags(doctype_id, doc_id, tag)
	VALUES(?, ?, ?)
	`
	_, err = tx.Exec(q, dtype, id, tag)
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

// RemoveTag disassociates the given tag from this document.
func (ds *_Documents) RemoveTag(otx *sql.Tx, user UserID, dtype DocTypeID, id DocumentID, tag string) error {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return errors.New("tag should not be empty")
	}
	tag = strings.ToLower(tag)

	tbl := _doctypes.docStorName(dtype)
	var duser UserID
	q := `SELECT user FROM ` + tbl + ` WHERE id = ?`
	row := db.QueryRow(q, id)
	err := row.Scan(&duser)
	if err != nil {
		return err
	}
	if duser != user {
		return fmt.Errorf("author mismatch -- original user : %d, current user : %d", duser, user)
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

	// Now write the database entry.

	q = `
	DELETE FROM wf_document_tags
	WHERE doctype_id = ?
	AND doc_id = ?
	AND tag = ?
	`
	_, err = tx.Exec(q, dtype, id, tag)
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

// Tags answers a list of the tags associated with this document.
func (ds *_Documents) Tags(dtype DocTypeID, id DocumentID) ([]string, error) {
	ts := make([]string, 0, 1)
	q := `
	SELECT tag
	FROM wf_document_tags
	WHERE doctype_id = ?
	AND doc_id = ?
	`
	rows, err := db.Query(q, dtype, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var t string
		err = rows.Scan(&t)
		if err != nil {
			return nil, err
		}
		ts = append(ts, t)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return ts, nil
}

// ChildrenIDs answers a list of this document's children IDs.
func (ds *_Documents) ChildrenIDs(dtype DocTypeID, id DocumentID) ([]struct {
	DocTypeID
	DocumentID
}, error) {
	cids := make([]struct {
		DocTypeID
		DocumentID
	}, 0, 1)

	q := `
	SELECT child_doctype_id, child_id
	FROM wf_document_children
	WHERE parent_doctype_id = ?
	AND parent_id = ?
	`
	rows, err := db.Query(q, dtype, id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var s struct {
			DocTypeID
			DocumentID
		}
		err = rows.Scan(&s.DocTypeID, &s.DocumentID)
		if err != nil {
			return nil, err
		}
		cids = append(cids, s)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return cids, nil
}
