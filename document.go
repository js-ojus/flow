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
	"log"
	"sync"
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
// Most applications should embed `Document` in their document
// structures rather than use this directly.
type Document struct {
	id     uint64 // globally-unique
	dtype  DocType
	title  string
	text   string // primary content
	author *User  // creator of the document
	ctime  time.Time

	mutex    sync.RWMutex
	state    *DocState   // current state
	events   []*DocEvent // state transitions so far, tracked in order
	revision uint16
	tags     []string // user-defined tags associated with this document
}

// NewDocument creates and initialises a document.
//
// The document created through this method has a life cycle that is
// associated with it through a particular workflow.
func NewDocument(id uint64, dtype DocType, title string, author *User, instate *DocState) (*Document, error) {
	if id == 0 || string(dtype) == "" || title == "" || author == nil {
		return nil, fmt.Errorf("invalid initialisation data -- id: %d, dtype: %s, title: %s, author: %s", id, dtype, title, author.Name())
	}

	d := &Document{id: id, dtype: dtype, title: title, author: author}
	d.ctime = time.Now().UTC()
	d.state = instate
	d.events = make([]*DocEvent, 1)
	d.revision = 1
	d.tags = make([]string, 1)
	return d, nil
}

// TODO(js): OpenDocument()

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
	if d.revision > 1 {
		return fmt.Errorf("document revision > 1 : %d", d.revision)
	}
	if d.text != "" {
		return fmt.Errorf("document text already set")
	}
	if t == "" {
		return fmt.Errorf("empty content given")
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.text = t
	return nil
}

// Text answers this document's primary content.
func (d *Document) Text() string {
	return d.text
}

// Author answers the user who created this document.
func (d *Document) Author() *User {
	return d.author
}

// Ctime answers the creation time of this document.
func (d *Document) Ctime() time.Time {
	return d.ctime
}

// State answer this document's current state.
func (d *Document) State() *DocState {
	return d.state
}

// Mtime answers the time when this document was most-recently
// modified.  It answers the creation time, if it has not been updated
// since.
func (d *Document) Mtime() time.Time {
	l := len(d.events)
	if l == 0 {
		return d.ctime
	}

	return d.events[l-1].mtime
}

// Revision answers the current revision number of this document.
func (d *Document) Revision() uint16 {
	return d.revision
}

// Events answers a copy of the sequence of events that has
// transformed this document so far.
func (d *Document) Events() []*DocEvent {
	// Synchronised because events are dynamic.
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	es := make([]*DocEvent, len(d.events))
	copy(es, d.events)
	return es
}

// applyEvent transitions this document into a new state as per the
// applied event.
//
// We perform as many validations as possible when constructing the
// event, so that we spend a minimum amount of time in this
// synchronised method.
func (d *Document) applyEvent(e *DocEvent) error {
	if e.doc.id != d.id {
		return fmt.Errorf("mismatched document IDs -- current: %d, event's: %d", d.id, e.doc.id)
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	if !e.mtime.IsZero() {
		return fmt.Errorf("event already applied")
	}
	if e.oldRev != d.revision {
		return fmt.Errorf("revision mismatch -- document rev: %d, event rev: %d", d.revision, e.oldRev)
	}

	e.mtime = time.Now().UTC()
	d.revision++
	e.newRev = d.revision
	d.state = e.state
	d.events = append(d.events, e)
	return nil
}

// AddTag associates the given tag with this document.
func (d *Document) AddTag(tag string) bool {
	if tag == "" {
		log.Printf("empty tag given")
		return false
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	for _, el := range d.tags {
		if el == tag {
			return false
		}
	}

	d.tags = append(d.tags, tag)
	return true
}

// RemoveTag disassociates the given tag from this document.
func (d *Document) RemoveTag(tag string) bool {
	if tag == "" {
		return false
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	idx := -1
	for i, el := range d.tags {
		if el == tag {
			idx = i
			break
		}
	}

	if idx > -1 {
		d.tags = append(d.tags[:idx], d.tags[idx+1:]...)
		return true
	}

	return false
}

// Tags answers a copy of the tags associated with this document.
func (d *Document) Tags() []string {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	ts := make([]string, len(d.tags))
	copy(ts, d.tags)
	return ts
}
