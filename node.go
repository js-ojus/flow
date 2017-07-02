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

import "time"

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

// newNode creates and initialises a node in the given workflow, as
// per the given node definition.
func newNode(defn *NodeDefinition, wf *Workflow, doc *Document) *Node {
	// WARNING: In a truly busy application, this manner of generating
	// IDs could lead to clashes.
	t := time.Now().UTC().UnixNano()
	return &Node{id: uint64(t), defn: defn, workflow: wf, doc: doc}
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

	// TODO(js): Resolve this.
	// err = n.doc.applyEvent(ev)
	if err != nil {
		return nil, err
	}

	n.workflow.moveToNode(n, msg.node)
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
