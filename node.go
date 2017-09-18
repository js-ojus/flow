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
	id      NodeID                  // Unique identifier of this node
	wflow   WorkflowID              // Containing flow of this node
	name    string                  // Unique within its workflow
	ntype   NodeType                // Topology type of this node
	state   DocStateID              // A document arriving at this node must be in this state
	nstates map[DocStateID]struct{} // List of possible states into which a document - currently at this node - can transition
	nfunc   NodeFunc                // Processing function of this node
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

// NextStates answers a list of possible document states into which a
// document - currently at this node - can transition.
func (n *Node) NextStates() ([]DocStateID, error) {
	l := len(n.nstates)

	// If we already have them handy, answer straight away.
	if l > 0 {
		sts := make([]DocStateID, 0, l)
		for st := range n.nstates {
			sts = append(sts, st)
		}
		return sts, nil
	}

	// Else, fetch from the database ...
	nstates, err := _nodes.NextStates(n.id)
	if err != nil {
		return nil, err
	}

	// ... and cache for future (i.e. for as long as this instance
	// lives).
	for _, st := range nstates {
		n.nstates[st] = struct{}{}
	}
	return nstates, nil
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

// NextStates answers the possible document states into which a
// document currently at this node can transition.
func (ns *_Nodes) NextStates(id NodeID) ([]DocStateID, error) {
	q := `
	SELECT docstate_id
	FROM wf_node_next_states
	WHERE node_id = ?
	`
	rows, err := db.Query(q, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ary := make([]DocStateID, 0, 5)
	for rows.Next() {
		var elem DocStateID
		err := rows.Scan(&elem)
		if err != nil {
			return nil, err
		}
		ary = append(ary, elem)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ary, nil
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
