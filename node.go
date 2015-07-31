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
// processors of documents in workflows.
//
// These functions generate document events that are applied to
// documents to transform them.  For purely informational messages,
// the event returned can be `nil` i.e., the document need not undergo
// a state transition necessarily.  However, for the workflow itself
// to move forward, the answered message must point to the next node
// in the flow.
//
// Error should be returned only when an impossible situation arises,
// and processing needs to abort.  Note that returning an error stops
// the workflow.  Manual intervention will be needed to move the
// document further.
//
// NodeFunc instances must be stateless and not capture their
// environment in any manner!
type NodeFunc func(*Document, ...interface{}) (*DocEvent, *Message, error)

// NodeDefinition represents a specific logical unit of processing in
// a workflow.  A single definition can be used to instantiate any
// number of node instances.
type NodeDefinition struct {
	wfdefn *WfDefinition // containing flow of this node
	name   string        // unique within its workflow
	ntype  NodeType
	nfunc  NodeFunc
}

// newNodeDefinition creates and initialises a node definition in the
// given workflow, using the given processing function.
func newNodeDefinition(wf *WfDefinition, name string, ntype NodeType, nfunc NodeFunc) *NodeDefinition {
	return &NodeDefinition{wfdefn: wf, name: name, ntype: ntype, nfunc: nfunc}
}

// WfDefinition answers this node definition's containing workflow
// definition.
func (nd *NodeDefinition) WfDefinition() *WfDefinition {
	return nd.wfdefn
}

// Name answers the descriptive title of this node.
func (nd *NodeDefinition) Name() string {
	return nd.name
}

// Type answers the type of this node.
func (nd *NodeDefinition) Type() NodeType {
	return nd.ntype
}

// Func answers the processing function registered in this node
// definition.
func (nd *NodeDefinition) Func() NodeFunc {
	return nd.nfunc
}

// Instance creates and initialises a node acting on the given
// document, using this node definition's processing function.
func (nd *NodeDefinition) Instance(wf *Workflow, doc *Document) (*Node, error) {
	if wf == nil || doc == nil {
		return nil, fmt.Errorf("invalid initialisation data -- workflow: %v, doc: %v", wf, doc)
	}

	// WARNING: In a truly busy application, this manner of generating
	// IDs could lead to clashes.
	t := time.Now().UTC().UnixNano()
	node := &Node{id: uint64(t), defn: nd, workflow: wf, doc: doc}
	return node, nil
}

// Node represents some processing or a response to an explicit user
// action.
type Node struct {
	id       uint64
	defn     *NodeDefinition
	workflow *Workflow
	doc      *Document
	resEv    *DocEvent
	resMsg   *Message
}

// Definition answers the node definition of this instance.
func (n *Node) Definition() *NodeDefinition {
	return n.defn
}

// Workflow answers the containing flow of this node.
func (n *Node) Workflow() *Workflow {
	return n.workflow
}

// Document answers the document in this workflow.
func (n *Node) Document() *Document {
	return n.doc
}

// Run processes the document in this node, using the registered
// function and the given parameters.
func (n *Node) Run(args ...interface{}) (*Message, error) {
	ev, msg, err := n.defn.nfunc(n.doc, args...)
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
