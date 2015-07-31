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

// NodeFunc defines the type of functions that can be used as
// processors of documents in workflows.  These functions generate
// document events that are applied to documents to transform them.
type NodeFunc func(*Document, ...interface{}) (*DocEvent, *Message, error)

// Node represents some processing or a response to an explicit user
// action.
type Node struct {
	id       uint64
	workflow *Workflow // containing flow of this node
	name     string    // for display purposes only
	ntype    NodeType
	nfunc    NodeFunc
	doc      *Document
	resEv    *DocEvent
	resMsg   *Message
}

// NewNode creates and initialises a node acting on the given
// document, using the given processing function.
func NewNode(wf *Workflow, name string, ntype NodeType, nfunc NodeFunc) (*Node, error) {
	if wf == nil || nfunc == nil {
		return nil, fmt.Errorf("invalid initialisation data -- workflow: %v, nfunc: %v", wf, nfunc)
	}

	// WARNING: In a truly busy application, this manner of generating
	// IDs could lead to clashes.
	t := time.Now().UTC().UnixNano()
	n := &Node{id: uint64(t), workflow: wf, name: name, ntype: ntype, nfunc: nfunc}
	return n, nil
}

// Workflow answers the containing flow of this node.
func (n *Node) Workflow() *Workflow {
	return n.workflow
}

// Name answers the descriptive title of this node.
func (n *Node) Name() string {
	return n.name
}

// Type answers the type of this node.
func (n *Node) Type() NodeType {
	return n.ntype
}

// SetDocument registers the document to process.
func (n *Node) SetDocument(doc *Document) error {
	if n.doc != nil {
		return fmt.Errorf("node already has a document: %d", n.doc.id)
	}
	if doc == nil {
		return fmt.Errorf("nil document provided")
	}

	n.doc = doc
	return nil
}

// Document answers the document in this workflow.
func (n *Node) Document() *Document {
	return n.doc
}

// Run processes the document in this node, using the registered
// function and the given parameters.
func (n *Node) Run(args ...interface{}) (*Message, error) {
	ev, msg, err := n.nfunc(n.doc, args...)
	if err != nil {
		return nil, err
	}

	// Needed for recovery.
	n.resEv = ev
	n.resMsg = msg

	err = n.doc.applyEvent(ev)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

// ResultEvent answers the document event that is the outcome of
// processing in this node, if the processing has already occurred.
func (n *Node) ResultEvent() *DocEvent {
	return n.resEv
}

// ResultMessage answers the message that represents the next step in
// the processing of this workflow, and which is handed over to the
// workflow instance.  This is available only after the processing has
// already occurred.
func (n *Node) ResultMessage() *Message {
	return n.resMsg
}
