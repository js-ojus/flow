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

// NodeID is the type of unique identifiers of nodes.
type NodeID int64

// NodeFunc defines the type of functions that can be used as
// post-processors of documents in workflows.
//
// These functions are triggered by document events that are applied
// to documents to transform them.  Invocation of a `NodeFunc` results
// in a transformed document, with the state of the system adjusted
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
	id          NodeID                     // Unique identifier of this node
	wflow       WorkflowID                 // Containing flow of this node
	name        string                     // Unique within its workflow
	ntype       NodeType                   // Topology type of this node
	state       DocStateID                 // A document arriving at this node must be in this state
	transitions map[DocActionID]DocStateID // Possible actions leading to states into which a document - currently at this node - can transition
	nfunc       NodeFunc                   // Processing function of this node
}

// ID answers the unique identifier of this workflow node.
func (n *Node) ID() NodeID {
	return n.id
}

// Workflow answers this node definition's containing workflow.
func (n *Node) Workflow() WorkflowID {
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

// State answers the state that any document arriving at this node
// must be in.
func (n *Node) State() DocStateID {
	return n.state
}

// Func answers the processing function registered in this node
// definition.
func (n *Node) Func() NodeFunc {
	return n.nfunc
}

// Unexported type, only for convenience methods.
type _Nodes struct{}

var _nodes *_Nodes

func init() {
	_nodes = &_Nodes{}
}

// Nodes provides a resource-like interface to the nodes defined in
// this system.
func Nodes() *_Nodes {
	return _nodes
}
