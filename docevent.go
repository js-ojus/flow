package flow

import "time"

// DocEvent represents a user action performed on a document in the
// system.
//
// Together with documents, events are central to the workflow engine
// in `flow`.  Events cause documents to switch from one state to
// another, usually in response to user actions.  They also carry
// information of the modification to the document.
type DocEvent struct {
	doc         *Document
	user        *User // user causing this modification
	mtime       time.Time
	text        string    // comment or other content
	newState    *DocState // result of the modification
	newRevision uint16    // serves as a cross-check
}
