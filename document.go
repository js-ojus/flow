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
	"errors"
	"fmt"
	"log"
	"strings"
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
// `Document` is a recursive structure: it can contain other
// documents.  Most applications should embed `Document` in their
// document structures rather than use this directly.
type Document struct {
	dtype     DocType   // for namespacing
	id        uint64    // globally-unique identifier of this document
	revision  uint16    // running revision number
	outerType DocType   // type of the containing document, if any
	outerID   uint64    // ID of the containing document, if any
	outer     *Document // containing document, if any (loaded)

	author uint64    // creator of this document
	state  *DocState // current state
	ctime  time.Time // creation time of this revision

	title    string               // human-readable title; applicable only for top-level documents
	data     []byte               // primary content of the document
	blobs    []string             // paths to enclosures
	tags     []string             // user-defined tags associated with this document
	children map[uint64]*Document // children documents of this document

	mutex sync.RWMutex
}

// NewDocument creates and initialises a document.
//
// The document created through this method has a life cycle that is
// associated with it through a particular workflow.
func NewDocument(dtype DocType, author uint64, instate *DocState) (*Document, error) {
	dt := strings.TrimSpace(string(dtype))
	if dt == "" || author == 0 {
		return nil, fmt.Errorf("invalid initialisation data -- dtype: %s, author: %d", dtype, author)
	}

	d := &Document{dtype: dtype, author: author, state: instate}
	d.children = make(map[uint64]*Document)
	return d, nil
}

// Type answers this document's type.
func (d *Document) Type() DocType {
	return d.dtype
}

// ID answers this document's globally-unique ID.
func (d *Document) ID() uint64 {
	return d.id
}

// Revision answers the current revision number of this document.
func (d *Document) Revision() uint16 {
	return d.revision
}

// OuterType answers this document's containing type, if any.
func (d *Document) OuterType() DocType {
	return d.outerType
}

// OuterID answers this document's containing document's globally-unique ID, if any.
func (d *Document) OuterID() uint64 {
	return d.outerID
}

// Outer answers this document's containing document, if any.
func (d *Document) Outer() *Document {
	return d.outer
}

// Author answers the user who created this document.
func (d *Document) Author() uint64 {
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

// Title answers this document's title.
func (d *Document) Title() string {
	return d.title
}

// SetTitle sets the title of the document, if one is not already set.
func (d *Document) SetTitle(title string) error {
	if d.title != "" {
		return errors.New("document already has a title")
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.title = strings.TrimSpace(title)
	return nil
}

// Data answers this document's primary data content.
func (d *Document) Data() []byte {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	dt := make([]byte, len(d.data))
	copy(dt, d.data)
	return dt
}

// SetData replaces this document's primary content with the given
// data, only if the user is the same as the author of this document.
func (d *Document) SetData(data []byte, author uint64) error {
	if author != d.author {
		return errors.New("author of modification not the same as the original author of this document")
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.data = make([]byte, len(data))
	copy(d.data, data)
	return nil
}

// Blobs answers the copy of this document's enclosures (as paths, not
// the actual blobs).
func (d *Document) Blobs() []string {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	bs := make([]string, len(d.blobs))
	copy(bs, d.blobs)
	return bs
}

// AddBlob adds the path to an enclosure to this document.
func (d *Document) AddBlob(path string, author uint64) error {
	if author != d.author {
		return errors.New("author of modification not the same as the original author of this document")
	}
	path = strings.TrimSpace(path)
	if path == "" {
		return errors.New("blob path should not be empty")
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.blobs = append(d.blobs, path)
	return nil
}

// RemoveBlob remove the specified path from the list of this
// document's blobs.
func (d *Document) RemoveBlob(path string, author uint64) error {
	if author != d.author {
		return errors.New("author of modification not the same as the original author of this document")
	}
	path = strings.TrimSpace(path)
	if path == "" {
		return errors.New("blob path should not be empty")
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	idx := -1
	for i, p := range d.blobs {
		if p == path {
			idx = i
			break
		}
	}
	if idx == -1 {
		return errors.New("given blob path not found")
	}

	d.blobs = append(d.blobs[:idx], d.blobs[idx+1:]...)
	return nil
}

// Tags answers a copy of the tags associated with this document.
func (d *Document) Tags() []string {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	ts := make([]string, len(d.tags))
	copy(ts, d.tags)
	return ts
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

// ChildrenIDs answers a copy of the list of this document's children
// IDs.
func (d *Document) ChildrenIDs() []uint64 {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	ids := make([]uint64, len(d.children))
	for k := range d.children {
		ids = append(ids, k)
	}
	return ids
}

// Children answers a copy of the list of this document's children.
func (d *Document) Children() []*Document {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	docs := make([]*Document, len(d.children))
	for _, v := range d.children {
		docs = append(docs, v)
	}
	return docs
}

// AddChild adds the given document as a child of this document, if it
// is not already a child.
func (d *Document) AddChild(cid uint64, ch *Document) error {
	if cid == 0 {
		return errors.New("document ID cannot be 0")
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	if _, ok := d.children[cid]; ok {
		return errors.New("given document is already a child of this document")
	}

	d.children[cid] = ch
	return nil
}

// RemoveChild removes the given child from this document.
func (d *Document) RemoveChild(cid uint64) error {
	if cid == 0 {
		return errors.New("document ID cannot be 0")
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	if _, ok := d.children[cid]; !ok {
		return errors.New("given document is not a child of this document")
	}

	delete(d.children, cid)
	return nil
}

// TODO(js): LoadDocument()

// TODO(js): Save()

// TODO(js): LoadChild()

// applyEvent transitions this document into a new state as per the
// applied event.
//
// We perform as many validations as possible when constructing the
// event, so that we spend a minimum amount of time in this
// synchronised method.
// func (d *Document) applyEvent(e *DocEvent) error {
// 	if e.doc.id != d.id {
// 		return fmt.Errorf("mismatched document IDs -- current: %d, event's: %d", d.id, e.doc.id)
// 	}

// 	d.mutex.Lock()
// 	defer d.mutex.Unlock()

// 	if !e.mtime.IsZero() {
// 		return fmt.Errorf("event already applied")
// 	}
// 	if e.oldRev != d.revision {
// 		return fmt.Errorf("revision mismatch -- document rev: %d, event rev: %d", d.revision, e.oldRev)
// 	}

// 	e.mtime = time.Now().UTC()
// 	d.revision++
// 	e.newRev = d.revision
// 	d.state = e.state
// 	d.events = append(d.events, e)
// 	return nil
// }
