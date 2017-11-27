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
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	// reDocPath defines the regular expression for each component of
	// a document's path.
	reDocPath = regexp.MustCompile("[0-9]+?:[0-9]+?/")
)

// DocPath helps in managing document hierarchies.  It provides a set
// of utility methods that ease path management.
type DocPath string

// Root answers the root document information.
func (p *DocPath) Root() (DocTypeID, DocumentID, error) {
	root := reDocPath.FindString(string(*p))
	if root == "" {
		return 0, 0, nil
	}

	parts := strings.Split(root, ":")
	dtid, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	did, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, err
	}

	return DocTypeID(dtid), DocumentID(did), nil
}

// Components answers a sequence of this path's components, in order.
func (p *DocPath) Components() ([]struct {
	DocTypeID
	DocumentID
}, error) {
	comps := reDocPath.FindAllString(string(*p), -1)
	if len(comps) == 0 {
		return nil, nil
	}

	ary := []struct {
		DocTypeID
		DocumentID
	}{}
	for _, comp := range comps {
		parts := strings.Split(comp, ":")
		dtid, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return nil, err
		}
		did, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return nil, err
		}

		ary = append(ary, struct {
			DocTypeID
			DocumentID
		}{DocTypeID(dtid), DocumentID(did)})
	}

	return ary, nil
}

// Append adds the given document type-document ID pair to this path,
// updating it as a result.
func (p *DocPath) Append(dtid DocTypeID, did DocumentID) error {
	if dtid <= 0 || did <= 0 {
		return errors.New("document type ID and document ID should be positive integers")
	}

	*p = *p + DocPath(fmt.Sprintf("%d:%d/", dtid, did))
	return nil
}

// Blob is a simple data holder for information concerning the
// user-supplied name of the binary object, the path of the stored
// binary object, and its SHA1 checksum.
type Blob struct {
	Name    string `json:"Name"`           // User-given name to the binary object
	Path    string `json:"Path,omitempty"` // Path to the stored binary object
	SHA1Sum string `json:"SHA1sum"`        // SHA1 checksum of the binary object
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
// documents.  Therefore, when a document is created, it is
// initialised with the path that leads from its root document to its
// immediate parent.  For root documents, this path is empty.
//
// Most applications should embed `Document` in their document
// structures rather than use this directly.  That enables them to
// control their data persistence mechanisms, while delegating
// workflow management to `flow`.
type Document struct {
	ID      DocumentID `json:"ID"`      // Globally-unique identifier of this document
	DocType DocType    `json:"DocType"` // For namespacing
	Path    DocPath    `json:"Path"`    // Path leading to, but not including, this document

	AccCtx AccessContext `json:"AccessContext"` // Originating access context of this document; applicable only to a root document
	State  DocState      `json:"DocState"`      // Current state of this document; applicable only to a root document

	Group GroupID   `json:"Group"` // Creator of this document
	Ctime time.Time `json:"Ctime"` // Creation time of this (possibly child) document

	Title string `json:"Title"` // Human-readable title; applicable only for root documents
	Data  []byte `json:"Data"`  // Primary content of the document
}

// Unexported type, only for convenience methods.
type _Documents struct{}

// Documents provides a resource-like interface to the documents in
// this system.
var Documents _Documents

// DocumentsNewInput specifies the initial data with which a new
// document has to be created in the system.
type DocumentsNewInput struct {
	DocTypeID                  // Type of the new document; required
	AccessContextID            // Access context in which the document should be created; required
	GroupID                    // (Singleton) group of the creator; required
	ParentType      DocTypeID  // Document type of the parent document, if any
	ParentID        DocumentID // Unique identifier of the parent document, if any
	Title           string     // Title of the new document; applicable to only root (top-level) documents
	Data            []byte     // Body of the new document; required
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
func (_Documents) New(otx *sql.Tx, input *DocumentsNewInput) (DocumentID, error) {
	if input.DocTypeID <= 0 || input.AccessContextID <= 0 || input.GroupID <= 0 {
		return 0, errors.New("all identifiers should be positive integers")
	}
	if len(input.Data) == 0 {
		return 0, errors.New("document's body should be non-empty")
	}

	var dsid int64
	var err error
	q := `
	SELECT docstate_id
	FROM wf_workflows
	WHERE doctype_id = ?
	AND active = 1
	`
	row := db.QueryRow(q, input.DocTypeID)
	err = row.Scan(&dsid)
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return 0, errors.New("no active workflow is defined for the given document type")

		default:
			return 0, err
		}
	}

	var path DocPath
	if input.ParentID > 0 {
		pdoc, err := Documents.Get(nil, input.ParentType, input.ParentID)
		if err != nil {
			return 0, err
		}
		path = pdoc.Path
		path.Append(input.ParentType, input.ParentID)

		// Child document does not have its own state.
		dsid = 0
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

	tbl := DocTypes.docStorName(input.DocTypeID)
	q2 := `INSERT INTO ` + tbl + `(path, ac_id, docstate_id, group_id, ctime, title, data)
	VALUES (?, ?, ?, NOW(), ?, ?)
	`
	res, err := tx.Exec(q2, path, input.AccessContextID, dsid, input.GroupID, input.Title, input.Data)
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
	res, err = tx.Exec(q2, input.ParentType, input.ParentID, input.DocTypeID, id)
	if err != nil {
		return 0, err
	}

	if otx == nil {
		err = tx.Commit()
		if err != nil {
			return 0, err
		}
	}

	return DocumentID(id), nil
}

// DocumentsListInput specifies a set of filter conditions to narrow
// down document listings.
type DocumentsListInput struct {
	DocTypeID                 // Documents of this type are listed; required
	AccessContextID           // Access context from within which to list; required
	GroupID                   // List documents created by this (singleton) group
	DocStateID                // List documents currently in this state
	CtimeStarting   time.Time // List documents created after this time
	CtimeBefore     time.Time // List documents created before this time
	TitleContains   string    // List documents whose title contains the given text; expensive operation
	RootOnly        bool      // List only root (top-level) documents
}

// List answers a subset of the documents based on the input
// specification.
//
// Result set begins with ID >= `offset`, and has not more than
// `limit` elements.  A value of `0` for `offset` fetches from the
// beginning, while a value of `0` for `limit` fetches until the end.
func (_Documents) List(input *DocumentsListInput, offset, limit int64) ([]*Document, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.New("offset and limit must be non-negative integers")
	}
	if limit == 0 {
		limit = math.MaxInt64
	}

	// Base query.

	tbl := DocTypes.docStorName(input.DocTypeID)
	q := `
	SELECT docs.id, docs.path, docs.group_id, docs.docstate_id, dsm.name, docs.ctime, docs.title
	FROM ` + tbl + ` docs
	JOIN wf_docstates_master dsm ON dsm.id = docs.docstate_id
	`

	// Process input specification.

	args := []interface{}{input.AccessContextID}
	q += `WHERE docs.ac_id = ?
	`

	if input.GroupID > 0 {
		q += `AND docs.group_id = ?
		`
		args = append(args, input.GroupID)
	}

	if input.DocStateID > 0 {
		q += `AND docs.docstate_id = ?
		`
		args = append(args, input.DocStateID)
	}

	if !input.CtimeStarting.IsZero() {
		q += `AND docs.ctime >= ?
		`
		args = append(args, input.CtimeStarting)
	}

	if !input.CtimeBefore.IsZero() {
		q += `AND docs.ctime < ?
		`
		args = append(args, input.CtimeBefore)
	}

	if input.TitleContains != "" {
		q += `AND docs.title LIKE ?
		`
		args = append(args, "%"+input.TitleContains+"%")
	}

	if input.RootOnly {
		q += `AND docs.path = ''
		`
	}

	q += `ORDER BY docs.id
	LIMIT ? OFFSET ?
	`
	args = append(args, limit, offset)

	// Fetch document data.

	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ary := make([]*Document, 0, 10)
	for rows.Next() {
		var elem Document
		var title sql.NullString
		err = rows.Scan(&elem.ID, &elem.Path, &elem.Group, &elem.State.ID, &elem.State.Name, &elem.Ctime, &title)
		if err != nil {
			return nil, err
		}
		elem.DocType.ID = input.DocTypeID
		if title.Valid {
			elem.Title = title.String
		}
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
func (_Documents) Get(otx *sql.Tx, dtype DocTypeID, id DocumentID) (*Document, error) {
	tbl := DocTypes.docStorName(dtype)
	var elem Document
	q := `
	SELECT docs.path, docs.group_id, docs.ctime, docs.title, docs.data, docs.docstate_id, dsm.name
	FROM ` + tbl + ` AS docs
	JOIN wf_docstates_master dsm ON docs.docstate_id = dsm.id
	WHERE docs.id = ?
	`

	var row *sql.Row
	if otx == nil {
		row = db.QueryRow(q, id)
	} else {
		row = otx.QueryRow(q, id)
	}
	err := row.Scan(&elem.Path, &elem.Group, &elem.Ctime, &elem.Title, &elem.Data, &elem.State.ID, &elem.State.Name)
	if err != nil {
		return nil, err
	}
	q = `SELECT name FROM wf_doctypes_master WHERE id = ?`
	row = db.QueryRow(q, dtype)
	err = row.Scan(&elem.DocType.Name)
	if err != nil {
		return nil, err
	}

	elem.ID = id
	elem.DocType.ID = dtype
	return &elem, nil
}

// GetParent answers the identifiers of the parent document of the
// specified document.
func (_Documents) GetParent(otx *sql.Tx, dtype DocTypeID, id DocumentID) (*Document, error) {
	q := `
	SELECT parent_doctype_id, parent_id
	FROM wf_document_children
	WHERE child_doctype_id = ?
	AND child_id = ?
	LIMIT 1
	`
	var row *sql.Row
	if otx == nil {
		row = db.QueryRow(q, dtype, id)
	} else {
		row = otx.QueryRow(q, dtype, id)
	}
	var ptid, pid int64
	err := row.Scan(&ptid, &pid)
	if err != nil {
		return nil, err
	}

	return Documents.Get(otx, DocTypeID(ptid), DocumentID(pid))
}

// setState sets the new state of the document.
//
// This method is not exported.  It is used internally by `Workflow`
// to move the document along the workflow, into a new document state.
func (_Documents) setState(otx *sql.Tx, dtype DocTypeID, id DocumentID, state DocStateID, ac AccessContextID) error {
	tbl := DocTypes.docStorName(dtype)

	var q string
	var err error
	if ac > 0 {
		q = `UPDATE ` + tbl + ` SET state = ?, ac_id = ? WHERE doc_id = ?`
		_, err = otx.Exec(q, state, ac, id)
	} else {
		q = `UPDATE ` + tbl + ` SET state = ? WHERE doc_id = ?`
		_, err = otx.Exec(q, state, id)
	}
	return err
}

// SetTitle sets the title of the document.
func (_Documents) SetTitle(otx *sql.Tx, dtype DocTypeID, id DocumentID, title string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return errors.New("document title should not be empty")
	}

	// A child document does not have its own title.
	tbl := DocTypes.docStorName(dtype)
	var path DocPath
	var dgroup GroupID
	q := `SELECT path, group_id FROM ` + tbl + ` WHERE id = ?`
	row := db.QueryRow(q, id)
	err := row.Scan(&path, &dgroup)
	if err != nil {
		return err
	}
	if path != "" {
		return errors.New("a child document cannot have its own title")
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
func (_Documents) SetData(otx *sql.Tx, dtype DocTypeID, id DocumentID, data []byte) error {
	if data == nil {
		return errors.New("document data should not be empty")
	}

	tbl := DocTypes.docStorName(dtype)

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

	q := `UPDATE ` + tbl + ` SET data = ? WHERE doc_id = ?`
	_, err := tx.Exec(q, data, id)
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
func (_Documents) Blobs(dtype DocTypeID, id DocumentID) ([]*Blob, error) {
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
		err = rows.Scan(&b.Name, &b.SHA1Sum)
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
func (_Documents) GetBlob(dtype DocTypeID, id Document, blob *Blob) error {
	if blob == nil {
		return errors.New("blob should be non-nil")
	}

	q := `
	SELECT name, path
	FROM wf_document_blobs
	WHERE doctype_id = ?
	AND doc_id = ?
	AND sha1sum = ?
	`
	row := db.QueryRow(q, dtype, id, blob.SHA1Sum)
	var b Blob
	err := row.Scan(&b.Name, &b.Path)
	if err != nil {
		return err
	}
	b.SHA1Sum = blob.SHA1Sum

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
func (_Documents) AddBlob(otx *sql.Tx, dtype DocTypeID, id DocumentID, blob *Blob) error {
	if blob == nil {
		return errors.New("blob should be non-nil")
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
	if blob.SHA1Sum != csum {
		return fmt.Errorf("checksum mismatch -- given SHA1 sum : %s, computed SHA1 sum : %s", blob.SHA1Sum, csum)
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
	//
	// TODO(js): Implement a better solution.
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

	q := `
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
func (_Documents) Tags(dtype DocTypeID, id DocumentID) ([]string, error) {
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

// AddTags associates the given tag with this document.
//
// Tags are converted to lower case (as per normal Unicode casing)
// before getting associated with documents.  Also, embedded spaces,
// if any, are retained.
func (_Documents) AddTags(otx *sql.Tx, dtype DocTypeID, id DocumentID, tags ...string) error {
	// A child document does not have its own tags.
	q := `
	SELECT parent_id
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
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		tag = strings.ToLower(tag)
		_, err = tx.Exec(q, dtype, id, tag)
		if err != nil {
			return err
		}
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
func (_Documents) RemoveTag(otx *sql.Tx, dtype DocTypeID, id DocumentID, tag string) error {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return errors.New("tag should not be empty")
	}
	tag = strings.ToLower(tag)

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
	q := `
	DELETE FROM wf_document_tags
	WHERE doctype_id = ?
	AND doc_id = ?
	AND tag = ?
	`
	_, err := tx.Exec(q, dtype, id, tag)
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
func (_Documents) ChildrenIDs(dtype DocTypeID, id DocumentID) ([]struct {
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
	defer rows.Close()

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
