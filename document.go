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

// DocSection represents one user-edited section of text and,
// possibly, enclosures.
//
// N.B. All updates to data in this structure must be made in the
// mutex context of its containing document.
type DocSection struct {
	id     uint16    // serial number of this section in the document
	text   string    // textual context of this section
	blobs  []string  // paths to enclosures
	author uint64    // editor of this section
	mtime  time.Time // modification time of this section
}

// NewDocSection creates a new section with the given data.
func NewDocSection(id uint16, text string, blobs []string, author uint64) *DocSection {
	return &DocSection{id: id, text: text, blobs: blobs, author: author, mtime: time.Now()}
}

// ID answers this section's unique identifier in its document.
func (s *DocSection) ID() uint16 {
	return s.id
}

// Text answers this section's text content.
func (s *DocSection) Text() string {
	return s.text
}

// setText replaces this section's text with the given text, only if
// the user is the same as the author of this section.
func (s *DocSection) setText(text string, author uint64) error {
	if author != s.author {
		return errors.New("author of modification not the same as the original author of this section")
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return errors.New("section text should not be empty")
	}

	s.text = text
	s.mtime = time.Now()
	return nil
}

// Blobs answers the copy of this section's enclosures (as paths, not
// the actual blobs).
func (s *DocSection) Blobs() []string {
	bs := make([]string, len(s.blobs))
	copy(bs, s.blobs)
	return bs
}

// addBlob adds the path to an enclosure to this section.
func (s *DocSection) addBlob(path string, author uint64) error {
	if author != s.author {
		return errors.New("author of modification not the same as the original author of this section")
	}
	path = strings.TrimSpace(path)
	if path == "" {
		return errors.New("blob path should not be empty")
	}

	s.blobs = append(s.blobs, path)
	s.mtime = time.Now()
	return nil
}

// removeBlob remove the specified path from the list of this
// section's blobs.
func (s *DocSection) removeBlob(path string, author uint64) error {
	if author != s.author {
		return errors.New("author of modification not the same as the original author of this section")
	}
	path = strings.TrimSpace(path)
	if path == "" {
		return errors.New("blob path should not be empty")
	}

	idx := -1
	for i, p := range s.blobs {
		if p == path {
			idx = i
			break
		}
	}
	if idx == -1 {
		return errors.New("given blob path not found")
	}

	s.blobs = append(s.blobs[:idx], s.blobs[idx+1:]...)
	s.mtime = time.Now()
	return nil
}

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
	id       uint64        // globally-unique
	dtype    DocType       // for namespacing
	title    string        // human-readable title
	sections []*DocSection // sections in this document
	owner    uint64        // creator of this document
	ctime    time.Time     // creation time of this revision

	mutex    sync.RWMutex
	state    *DocState   // current state
	events   []*DocEvent // state transitions so far, tracked in order
	revision uint16      // running revision number
	tags     []string    // user-defined tags associated with this document
}

// NewDocument creates and initialises a document.
//
// The document created through this method has a life cycle that is
// associated with it through a particular workflow.
func NewDocument(id uint64, dtype DocType, title string, owner uint64, instate *DocState) (*Document, error) {
	dt := strings.TrimSpace(string(dtype))
	title = strings.TrimSpace(title)
	if id == 0 || dt == "" || title == "" || owner == 0 {
		return nil, fmt.Errorf("invalid initialisation data -- id: %d, dtype: %s, title: %s, author: %d", id, dtype, title, owner)
	}

	d := &Document{id: id, dtype: dtype, title: title, owner: owner}
	d.sections = make([]*DocSection, 0, 1)
	d.ctime = time.Now().UTC()
	d.state = instate
	d.events = make([]*DocEvent, 0, 1)
	d.revision = 1
	d.tags = make([]string, 0, 1)
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

// Sections answers the list of this document's sections.
func (d *Document) Sections() []*DocSection {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	dss := make([]*DocSection, len(d.sections))
	copy(dss, d.sections)
	return dss
}

// SetSectionText sets the text of the given section to the given
// text.
func (d *Document) SetSectionText(id uint16, text string, author uint64) error {
	if id == 0 || int(id) > len(d.sections) {
		return errors.New("section ID out of bounds")
	}

	d.mutex.RLock()
	defer d.mutex.RUnlock()
	sec := d.sections[id-1]
	return sec.setText(text, author)
}

// AddSectionBlob adds the given path to a blob to the specified
// section.
func (d *Document) AddSectionBlob(id uint16, path string, author uint64) error {
	if id == 0 || int(id) > len(d.sections) {
		return errors.New("section ID out of bounds")
	}

	d.mutex.RLock()
	defer d.mutex.RUnlock()
	sec := d.sections[id-1]
	return sec.addBlob(path, author)
}

// RemoveSectionBlob removes the given path to a blob to the specified
// section.
func (d *Document) RemoveSectionBlob(id uint16, path string, author uint64) error {
	if id == 0 || int(id) > len(d.sections) {
		return errors.New("section ID out of bounds")
	}

	d.mutex.RLock()
	defer d.mutex.RUnlock()
	sec := d.sections[id-1]
	return sec.removeBlob(path, author)
}

// Owner answers the user who created this document.
func (d *Document) Owner() uint64 {
	return d.owner
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
