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
type NodeFunc func(*Document, *DocEvent, ...interface{}) (DocState, *Message, error)

// Node represents a specific logical unit of processing and routing
// in a workflow.
type Node struct {
	wflow  *Workflow              // containing flow of this node
	name   string                 // unique within its workflow
	ntype  NodeType               // topology type of this node
	nfunc  NodeFunc               // processing function of this node
	xforms map[DocAction]DocState // map of which action at this node transforms a document into which state
}

// NewNode creates and initialises a node definition in the
// given workflow, using the given processing function.
func NewNode(wf *Workflow, name string, ntype NodeType, nfunc NodeFunc) *Node {
	node := &Node{wflow: wf, name: name, ntype: ntype, nfunc: nfunc}
	node.xforms = make(map[DocAction]DocState)
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

// AddXform adds a transform at this node, for the given document
// action to the specified target document state.
//
// N.B. This method is not protected by a mutex since this is expected
// to be exercised only during start-up.  Do not violate that!
func (n *Node) AddXform(da DocAction, ds DocState) error {
	if tds, ok := n.xforms[da]; ok {
		return fmt.Errorf("target state '%s' already registered for the given action", tds.name)
	}

	n.xforms[da] = ds
	return nil
}
