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
)

// NodeFunc defines the type of functions that can be used as
// processors of documents in workflows.
//
// These functions consume document events that are applied to
// documents to transform them.  Invocation of a `NodeFunc` results in
// a transformed document, with the state of the system adjusted
// accordingly.  It is, of course, possible for an event to not result
// in a document transformation or a state transition.
//
// Error should be returned only when an impossible situation arises,
// and processing needs to abort.  Note that returning an error stops
// the workflow.  Manual intervention will be needed to move the
// document further.
//
// N. B. NodeFunc instances must be stateless and not capture their
// environment in any manner!
type NodeFunc func(*Document, DocAction, ...interface{}) (DocState, *Message, error)

// Node represents a specific logical unit of processing and routing
// in a workflow.
type Node struct {
	wflow  *Workflow             // containing flow of this node
	name   string                // unique within its workflow
	state  DocState              // a document reaching this node must be in this state
	ntype  NodeType              // topology type of this node
	nfunc  NodeFunc              // processing function of this node
	xforms map[DocState]struct{} // list of possible states into which a document can transition
}

// NewNode creates and initialises a node definition in the
// given workflow, using the given processing function.
func NewNode(wf *Workflow, name string, state DocState, ntype NodeType, nfunc NodeFunc) *Node {
	node := &Node{wflow: wf, name: name, state: state, ntype: ntype, nfunc: nfunc}
	node.xforms = make(map[DocState]struct{})
	return node
}

// Workflow answers this node definition's containing workflow
// definition.
func (n *Node) Workflow() *Workflow {
	return n.wflow
}

// Name answers the descriptive title of this node.
func (n *Node) Name() string {
	return n.name
}

// Type answers the type of this node.
func (n *Node) Type() NodeType {
	return n.ntype
}

// Func answers the processing function registered in this node
// definition.
func (n *Node) Func() NodeFunc {
	return n.nfunc
}

// AddXform adds a possible transition at this node, for a document
// arriving at this node.
//
// N.B. This method is not protected by a mutex since this is expected
// to be exercised only during start-up.  Do not violate that!
func (n *Node) AddXform(ds DocState) error {
	if _, ok := n.xforms[ds]; ok {
		return fmt.Errorf("target state '%s' already registered", ds.name)
	}

	n.xforms[ds] = struct{}{}
	return nil
}

// applyEvent takes an input user action or a system event, and
// applies its document action to the given document.  This results in
// a possibly new document state.  In addition, a registered
// processing function is invoked on the document to prepare a message
// that can be posted to applicable mailboxes.
func (n *Node) applyEvent(event *DocEvent, args ...interface{}) error {
	doc, err := _documents.Get(event.dtype, event.docID)
	if err != nil {
		return err
	}

	// nds, msg, err := n.nfunc(event.doc, event.action, args...)
	_, _, err = n.nfunc(doc, event.action, args...)
	if err != nil {
		return err
	}

	// TODO(js): routing of message

	return nil
}
