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
	"fmt"
	"time"
)

// DocEvent represents a user action performed on a document in the
// system.
//
// Together with documents, events are central to the workflow engine
// in `flow`.  Events cause documents to switch from one state to
// another, usually in response to user actions.  They also carry
// information of the modification to the document.
type DocEvent struct {
	doc      *Document
	user     *User // user causing this modification
	mtime    time.Time
	text     string    // comment or other content
	state    *DocState // result of the modification
	revision uint16    // serves as a cross-check
}

// NewDocEvent creates and initialises an event that transforms the
// document that it refers to.
func NewDocEvent(doc *Document, user *User, text string, ns *DocState) (*DocEvent, error) {
	if doc == nil || user == nil || ns == nil {
		return nil, fmt.Errorf("nil initialisation data")
	}
	if doc.dtype != ns.dtype {
		return nil, fmt.Errorf("mismatched document types -- document's: %s, new state's: %s", doc.dtype, ns.dtype)
	}

	found := false
	ds := doc.state
	for _, el := range ds.successors {
		if el.name == ns.name {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("impossible transition requested -- current: %s, new: %s", ds.name, ns.name)
	}

	e := &DocEvent{doc: doc, user: user, text: text, state: ns}
	return e, nil
}

// Document answers the document which this event transformed.
func (e *DocEvent) Document() *Document {
	return e.doc
}

// User answers the user who caused this event to occur.
func (e *DocEvent) User() *User {
	return e.user
}

// Mtime answers the time when this event affected the referred
// document.  If this event has not been applied to the document yet,
// the answered time satisfies `time.IsZero()`.
func (e *DocEvent) Mtime() time.Time {
	return e.mtime
}

// State answers the document state into which the referred document
// transitioned.
func (e *DocEvent) State() *DocState {
	return e.state
}

// Revision answers the revision number of the referred document after
// this event affected it.  If this event has not been applied to the
// document yet, the answered revision number is zero.
func (e *DocEvent) Revision() uint16 {
	return e.revision
}
