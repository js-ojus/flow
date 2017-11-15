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
	"math"
	"os"
	"path"
	"strings"
	"time"
)

// Blob is a simple data holder for information concerning the
// user-supplied name of the binary object, the path of the stored
// binary object, and its SHA1 checksum.
type Blob struct {
	Name    string `json:"Name"`           // User-given name to the binary object
	Path    string `json:"Path,omitempty"` // Path to the stored binary object
	Sha1Sum string `json:"Sha1sum"`        // SHA1 checksum of the binary object
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
	ID      DocumentID `json:"ID"`      // Globally-unique identifier of this document
	DocType DocType    `json:"DocType"` // For namespacing

	OrigAccCtx AccessContextID `json:"OrigAccessContext"` // The originating access context of this document; applicable only to a top-level document
	AccCtx     AccessContext   `json:"AccessContext"`     // Current access context of this document; applicable only to a top-level document
	State      DocState        `json:"DocState"`          // Current state of this document; applicable only to a top-level document

	User  UserID    `json:"User"`  // Creator of this document
	Ctime time.Time `json:"Ctime"` // Creation time of this (possibly child) document

	Title string `json:"Title"` // Human-readable title; applicable only for top-level documents
	Data  []byte `json:"Data"`  // Primary content of the document
}

// Unexported type, only for convenience methods.
type _Documents struct{}

var _documents *_Documents

func init() {
	_documents = &_Documents{}
}

// Documents provides a resource-like interface to the documents in
// this system.
func Documents() *_Documents {
	return _documents
}

// New creates and initialises a document.
//
// The document created through this method has a life cycle that is
// associated with it through a particular workflow.  In addition, the
// operations that different users can perform on this document, are
// determined in the scope of the access context applicable to the
// current state of the document.
//
// N.B. Blobs, tags and children documents have to be associated with
// this document, if needed, through appropriate separate calls.
func (ds *_Documents) New(otx *sql.Tx, acID AccessContextID,
	user UserID, dtype DocTypeID, otype DocTypeID, oid DocumentID,
	title string, data []byte) (DocumentID, error) {
	if user <= 0 || acID <= 0 || dtype <= 0 || otype < 0 || oid < 0 {
		return 0, errors.New("all identifiers should be positive integers; parent document references should be zero or positive integers")
	}
	if len(data) == 0 {
		return 0, errors.New("document's body should be non-empty")
	}

	if oid > 0 {
		// A child document does not have its own title.
		outer, err := _documents.Get(otx, otype, oid)
		if err != nil {
			return 0, err
		}
		title = outer.Title
	}

	q := `
	SELECT docstate_id
	FROM wf_workflows
	WHERE doctype_id = ?
	AND active = 1
	`
	var dsid int64
	row := db.QueryRow(q, dtype)
	err := row.Scan(&dsid)
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return 0, errors.New("no active workflow is defined for the given document type")

		default:
			return 0, err
		}
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
	q2 := `INSERT INTO ` + tbl + `(orig_ac_id, ac_id, docstate_id, user_id, ctime, title, data)
	VALUES (?, ?, ?, ?, NOW(), ?, ?)
	`
	res, err := tx.Exec(q2, acID, acID, dsid, user, title, data)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	q2 = `
	INSERT INTO wf_document_children(parent_doctype_id, parent_id, child_doctype_id, child_id)
	VALUES (?, ?, ?, ?)
	`
	if oid > 0 {
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

// List answers a subset of the documents of the given document type,
// based on the input specification.
//
// Result set begins with ID >= `offset`, and has not more than
// `limit` elements.  A value of `0` for `offset` fetches from the
// beginning, while a value of `0` for `limit` fetches until the end.
func (ds *_Documents) List(dtype DocTypeID, offset, limit int64) ([]*Document, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.New("offset and limit must be non-negative integers")
	}
	if limit == 0 {
		limit = math.MaxInt64
	}

	tbl := _doctypes.docStorName(dtype)
	q := `
	SELECT id, user_id, docstate_id, ctime, title, data
	FROM ` + tbl + `
	ORDER BY id
	LIMIT ? OFFSET ?
	`
	rows, err := db.Query(q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ary := make([]*Document, 0, 10)
	for rows.Next() {
		var elem Document
		err = rows.Scan(&elem.ID, &elem.User, &elem.State, &elem.Ctime, &elem.Title, &elem.Data)
		if err != nil {
			return nil, err
		}
		elem.DocType.ID = dtype
		ary = append(ary, &elem)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ary, nil
}

// Get initialises a document by reading from the database.
//
// N.B. This retrieves the primary data of the document.  Other
// information viz. blobs, tags and children documents have to be
// fetched separately.
func (ds *_Documents) Get(otx *sql.Tx, dtype DocTypeID, id DocumentID) (*Document, error) {
	tbl := _doctypes.docStorName(dtype)
	var d Document
	q := `
	SELECT docs.user_id, docs.docstate_id, docs.ctime, docs.title, docs.data, states.name
	FROM ` + tbl + ` AS docs
	JOIN wf_docstates_master states ON docs.docstate_id = states.id
	WHERE docs.id = ?
	`

	var row *sql.Row
	if otx == nil {
		row = db.QueryRow(q, id, dtype)
	} else {
		row = otx.QueryRow(q, id, dtype)
	}
	err := row.Scan(&d.User, &d.State.ID, &d.Ctime, &d.Title, &d.Data, &d.State.Name)
	if err != nil {
		return nil, err
	}
	q = `SELECT name FROM wf_doctypes_master WHERE id = ?`
	row = db.QueryRow(q, dtype)
	err = row.Scan(&d.DocType.Name)
	if err != nil {
		return nil, err
	}

	d.ID = id
	d.DocType.ID = dtype
	return &d, nil
}

// GetOuter answers the identifiers of the parent document of the
// specified document.
func (ds *_Documents) GetOuter(dtype DocTypeID, id DocumentID) (DocTypeID, DocumentID, error) {
	q := `
	SELECT parent_doctype_id, parent_id
	FROM wf_document_children
	WHERE child_doctype_id = ?
	AND child_id = ?
	LIMIT 1
	`
	row := db.QueryRow(q, dtype, id)
	var dtid, did int64
	err := row.Scan(&dtid, &did)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, 0, errors.New("this is a top-level document")
		}
		return 0, 0, err
	}

	return DocTypeID(dtid), DocumentID(did), nil
}

// setState sets the new state of the document.
//
// This method is not exported.  It is used internally by `Workflow`
// to move the document along the workflow, into a new document state.
func (ds *_Documents) setState(otx *sql.Tx, dtype DocTypeID, id DocumentID, state DocStateID) error {
	tbl := _doctypes.docStorName(dtype)

	q := `UPDATE ` + tbl + ` SET state = ? WHERE doc_id = ?`
	_, err := otx.Exec(q, state, id)
	return err
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

// Blobs answers a list of this document's enclosures (as names, not
// the actual blobs).
func (ds *_Documents) Blobs(dtype DocTypeID, id DocumentID) ([]*Blob, error) {
	bs := make([]*Blob, 0, 1)
	q := `
	SELECT name, sha1sum
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
		err = rows.Scan(&b.Name, &b.Sha1Sum)
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

// GetBlob retrieves the requested blob from the specified document,
// if one such exists.  Lookup happens based on the given blob name.
// The retrieved blob is copied into the specified path.
func (ds *_Documents) GetBlob(dtype DocTypeID, id Document, blob *Blob) error {
	if blob == nil {
		return errors.New("blob should be non-nil")
	}

	q := `
	SELECT name, path, sha1sum
	FROM wf_document_blobs
	WHERE doctype_id = ?
	AND doc_id = ?
	AND name = ?
	`
	row := db.QueryRow(q, dtype, id, blob.Name)
	var b Blob
	err := row.Scan(&b.Name, &b.Path, &b.Sha1Sum)
	if err != nil {
		return err
	}

	// Copy the blob into the destination path given.

	inf, err := os.Open(b.Path)
	if err != nil {
		return err
	}
	defer inf.Close()
	outf, err := os.Create(blob.Path)
	if err != nil {
		return err
	}
	defer outf.Close()
	_, err = io.Copy(outf, inf)
	if err != nil {
		return err
	}

	return nil
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
