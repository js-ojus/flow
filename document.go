package flow

import "time"

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
