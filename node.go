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

import "fmt"

// NodeFunc defines the type of functions that can be used as
// processors of documents in workflows.  These functions generate
// document events that are applied to documents to transform them.
type NodeFunc func(*Document, ...interface{}) (*DocEvent, *Message, error)

// Node represents some processing or a response to an explicit user
// action.
type Node struct {
	workflow *Workflow // containing flow of this node
	name     string    // unique name within its workflow
	ntype    NodeType
	doc      *Document
	nfunc    NodeFunc
	event    *DocEvent
}

// NewNode creates and initialises a node acting on the given
// document, using the given processing function.
func NewNode(doc *Document, nfunc NodeFunc) (*Node, error) {
	if doc == nil || nfunc == nil {
		return nil, fmt.Errorf("invalid initialisation data -- doc: %v, nfunc: %v", doc, nfunc)
	}

	n := &Node{doc: doc, nfunc: nfunc}
	return n, nil
}

// Document answers the document in this workflow.
func (n *Node) Document() *Document {
	return n.doc
}

// Type answers the type of this node.
func (n *Node) Type() NodeType {
	return n.ntype
}

// Run processes the document in this node, using the registered
// function and the given parameters.
func (n *Node) Run(args ...interface{}) (*Message, error) {
	ev, msg, err := n.nfunc(n.doc, args)
	if err != nil {
		return nil, err
	}

	err = n.doc.applyEvent(ev)
	if err != nil {
		return nil, err
	}

	return msg, nil
}
