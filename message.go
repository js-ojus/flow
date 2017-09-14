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
	"fmt"
	"time"
)

// Message is the fundamental unit of workflow transition from a state
// to another.
//
// Messages can be informational or seek action.  Each message that
// seeks user intervention contains a reference to the document that
// began the current workflow, as well as the next node in the
// workflow.
type Message struct {
	id       uint64    // globally-unique
	workflow *Workflow // containing workflow
	node     *Node     // next step in the workflow
	user     *User     // intended recipient of this message
	group    *Group    // in case of a group 'to do' item or a broadcast message
	doc      *Document // can be `nil` for informational messages
	text     string    // must be non-empty if doc == nil
	mtime    time.Time // time of last modification of the message
}

// NewMessage creates and initialises a message to participate in the
// given workflow.
func NewMessage(wf *Workflow, node *Node) (*Message, error) {
	if wf == nil || node == nil {
		return nil, fmt.Errorf("invalid initialisation data -- workflow: %v, node: %v", wf, node)
	}

	// WARNING: In a truly busy application, this manner of generating
	// IDs could lead to clashes.
	t := time.Now().UTC()
	msg := &Message{id: uint64(t.UnixNano()), workflow: wf, node: node}
	msg.mtime = t
	return msg, nil
}
