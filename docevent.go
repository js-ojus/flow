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
	"strings"
)

// DocEvent represents a user action performed on a document in the
// system.
//
// Together with documents and nodes, events are central to the
// workflow engine in `flow`.  Events cause documents to transition
// from one state to another, usually in response to user actions.  It
// is possible for system events to cause state transitions, as well.
type DocEvent struct {
	doc    *Document // document to which this event is to be applied
	user   uint64    // user who caused this action
	action DocAction // action performed by the user
	text   string    // comment or other content
}

// NewDocEvent creates and initialises an event that transforms the
// document that it refers to.
func NewDocEvent(doc *Document, user uint64, action DocAction, text string) (*DocEvent, error) {
	taction := strings.TrimSpace(string(action))
	if doc == nil || user == 0 || taction == "" {
		return nil, fmt.Errorf("nil initialisation data")
	}

	e := &DocEvent{doc: doc, user: user, action: DocAction(taction), text: text}
	return e, nil
}

// Document answers the document which this event transformed.
func (e *DocEvent) Document() *Document {
	return e.doc
}

// User answers the user who caused this event to occur.
func (e *DocEvent) User() uint64 {
	return e.user
}

// Action answers the document action that this event represents.
func (e *DocEvent) Action() DocAction {
	return e.action
}

// Text answers the comment or other content enclosed in this event.
func (e *DocEvent) Text() string {
	return e.text
}
