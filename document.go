package flow

import (
	"fmt"
	"time"
)

// Document represents a task in a workflow, whose life cycle it
// tracks.
//
// Documents are central to the workflow engine and its operations.
// Each document represents a task whose life cycle it tracks.  In the
// process, it accumulates various details, and tracks the times of
// its modifications.  The life cycle typically involves several state
// transitions, whose details are also tracked.
//
// Applications are expected to embed `Document` in their document
// structures.
type Document struct {
	id       uint64 // globally-unique
	dtype    DocType
	title    string
	text     string // primary content
	author   *User  // creator of the document
	ctime    time.Time
	events   []*DocEvent // state transitions so far, tracked in order
	revision uint16
}

// NewDocument creates and initialises a document.
//
// The document created through this method has a life cycle that is
// associated with it through a particular workflow.
func NewDocument(id uint64, dtype DocType, title string, author *User) (*Document, error) {
	if id == 0 || string(dtype) == "" || title == "" || author == nil {
		return nil, fmt.Errorf("invalid initialisation data -- id: %d, dtype: %s, title: %s, author: %s", id, dtype, title, author.Name())
	}

	d := &Document{id: id, dtype: dtype, title: title, author: author}
	d.ctime = time.Now().UTC()
	d.events = make([]*DocEvent, 1)
	return d, nil
}

// ID answers this document's globally-unique ID.
func (d *Document) ID() uint64 {
	return d.id
}

// Type answers this document's type.
func (d *Document) Type() DocType {
	return d.dtype
}

// Title answers this document's title.
func (d *Document) Title() string {
	return d.title
}

// SetText sets the primary content of this document.
func (d *Document) SetText(t string) error {
	if t == "" {
		return fmt.Errorf("empty content given")
	}

	d.text = t
	return nil
}

// Text answers this document's primary content.
func (d *Document) Text() string {
	return d.text
}

// Revision answers the current revision number of this document.
func (d *Document) Revision() uint16 {
	return d.revision
}
