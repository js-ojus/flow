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
	"database/sql"
	"errors"
	"fmt"
)

// NodeID is the type of unique identifiers of nodes.
type NodeID int64

// NodeFunc defines the type of functions that generate notification
// messages in workflows.
//
// These functions are triggered by appropriate nodes, when document
// events are applied to documents to possibly transform them.
// Invocation of a `NodeFunc` should result in a message that can then
// be dispatched to applicable mailboxes.
//
// Error should be returned only when an impossible situation arises,
// and processing needs to abort.  Note that returning an error stops
// the workflow.  Manual intervention will be needed to move the
// document further.
//
// N. B. NodeFunc instances must be referentially transparent --
// stateless and not capture their environment in any manner.
// Unexpected bad things could happen otherwise!
type NodeFunc func(*Document, *DocEvent) *Message

// defNodeFunc prepares a simple message that can be posted to
// applicable mailboces.
func defNodeFunc(d *Document, event *DocEvent) *Message {
	return &Message{
		dtype: d.dtype.ID,
		docID: d.id,
		event: event.ID,
		title: d.title,
		data:  event.Text,
	}
}

// Node represents a specific logical unit of processing and routing
// in a workflow.
type Node struct {
	id    NodeID     // Unique identifier of this node
	dtype DocTypeID  // Document type which this node's workflow manages
	state DocStateID // A document arriving at this node must be in this state
	wflow WorkflowID // Containing flow of this node
	name  string     // Unique within its workflow
	ntype NodeType   // Topology type of this node
	nfunc NodeFunc   // Processing function of this node
}

// ID answers the unique identifier of this workflow node.
func (n *Node) ID() NodeID {
	return n.id
}

// Workflow answers this node definition's containing workflow.
func (n *Node) Workflow() WorkflowID {
	return n.wflow
}

// State answers the state that any document arriving at this node
// must be in.
func (n *Node) State() DocStateID {
	return n.state
}

// Name answers the descriptive title of this node.
func (n *Node) Name() string {
	return n.name
}

// Type answers the type of this node.
func (n *Node) Type() NodeType {
	return n.ntype
}

// Transitions answers the possible document states into which a
// document currently in the given state can transition.
func (n *Node) Transitions() (map[DocActionID]DocStateID, error) {
	q := `
	SELECT docaction_id, to_state_id
	FROM wf_docstate_transitions
	WHERE doctype_id = ?
	AND from_state_id = ?
	`
	rows, err := db.Query(q, n.dtype, n.state)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hash := make(map[DocActionID]DocStateID)
	for rows.Next() {
		var da DocActionID
		var ds DocStateID
		err := rows.Scan(&da, &ds)
		if err != nil {
			return nil, err
		}
		hash[da] = ds
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return hash, nil
}

// SetFunc registers the given node function with this node.
//
// If `nil` is given, a default node function is registered instead.
// This default function sets the document title as the message
// subject, and the event's data as the message body.
func (n *Node) SetFunc(fn NodeFunc) error {
	if fn == nil {
		n.nfunc = defNodeFunc
		return nil
	}

	n.nfunc = fn
	return nil
}

// Func answers the processing function registered in this node
// definition.
func (n *Node) Func() NodeFunc {
	return n.nfunc
}

// applyEvent checks to see if the given event can be applied
// successfully.  Accordingly, it prepares a message by utilising the
// registered node function, and posts it to applicable mailboxes.
func (n *Node) applyEvent(otx *sql.Tx, event *DocEvent, recipients []GroupID) (DocStateID, error) {
	ts, err := n.Transitions()
	nstate, ok := ts[event.Action]
	if !ok {
		return 0, fmt.Errorf("given event's action '%d' cannot be performed on a document in the state '%d'", event.Action, n.state)
	}

	doc, err := _documents.Get(event.DocType, event.DocID)
	if err != nil {
		return 0, err
	}
	if doc.state.ID != event.State && doc.state.ID != nstate {
		return 0, fmt.Errorf("document state is : %d, but event is targeting state : %d", doc.state.ID, event.State)
	}

	var tx *sql.Tx
	if otx == nil {
		tx, err := db.Begin()
		if err != nil {
			return 0, err
		}
		defer tx.Rollback()
	} else {
		tx = otx
	}

	// Process state transition according to the target node type.

	tnode, err := _nodes.GetByState(n.dtype, nstate)
	if err != nil {
		return 0, err
	}

	switch tnode.ntype {
	case NodeTypeJoinAny:
		// Multiple 'in's, but any one suffices.

		// Document has already transitioned; nothing to do.
		if doc.state.ID == nstate {
			return nstate, nil
		}

		// This is the first event; process it.
		fallthrough

	case NodeTypeBegin, NodeTypeEnd, NodeTypeLinear, NodeTypeBranch:
		// Any node type having a single 'in'.

		// Update the document to transition the state.
		tbl := _doctypes.docStorName(event.DocType)
		q := `UPDATE ` + tbl + ` SET state = ? WHERE doc_id = ?`
		_, err = tx.Exec(q, nstate, event.DocID)
		if err != nil {
			return 0, err
		}

		// Record event application.
		err = n.recordEvent(tx, event, nstate)
		if err != nil {
			return 0, err
		}

		// Post messages.
		msg := n.nfunc(doc, event)
		err = n.postMessage(tx, msg, recipients)
		if err != nil {
			return 0, err
		}

	case NodeTypeJoinAll:
		// Multiple 'in's, and all are required.

		// TODO(js)

	default:
		return 0, fmt.Errorf("unknown node type encountered : %s", tnode.ntype)
	}

	if otx == nil {
		err = tx.Commit()
		if err != nil {
			return 0, err
		}
	}

	return nstate, nil
}

// recordEvent writes a record stating that the given event has
// successfully been applied to effect a document state transition.
func (n *Node) recordEvent(otx *sql.Tx, event *DocEvent, nstate DocStateID) error {
	q := `
	INSERT INTO wf_docevent_application(doctype_id, doc_id, from_state_id, docevent_id, to_state_id)
	VALUES(?, ?, ?, ?, ?)
	`
	_, err := otx.Exec(q, event.DocType, event.DocID, event.State, event.ID, nstate)
	if err != nil {
		return err
	}

	q = `UPDATE wf_docevents SET status = 'A' WHERE id = ?`
	_, err = otx.Exec(q, event.ID)
	if err != nil {
		return err
	}

	return nil
}

// postMessage posts the given message into the mailboxes of the
// specified recipients.
func (n *Node) postMessage(otx *sql.Tx, msg *Message, recipients []GroupID) error {
	// Record the message.

	q := `
	INSERT INTO wf_messages(doctype_id, doc_id, docevent_id, title, data)
	VALUES(?, ?, ?, ?, ?)
	`
	res, err := otx.Exec(q, msg.dtype, msg.docID, msg.event, msg.title, msg.data)
	if err != nil {
		return err
	}
	var msgid int64
	if msgid, err = res.LastInsertId(); err != nil {
		return err
	}

	// Post it into applicable mailboxes.

	q = `
	INSERT INTO wf_mailboxes(group_id, message_id, unread)
	VALUES(?, ?, 1)
	`
	for _, gid := range recipients {
		res, err = otx.Exec(q, gid, msgid)
		if err != nil {
			return err
		}
	}

	return nil
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

// Nodes answers a list of the nodes comprising the given workflow.
func (ns *_Nodes) Nodes(id WorkflowID) ([]*Node, error) {
	q := `
	SELECT id, doctype_id, docstate_id, workflow_id, name, type
	FROM wf_workflow_nodes
	WHERE workflow_id = ?
	`
	rows, err := db.Query(q, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ary := make([]*Node, 0, 5)
	for rows.Next() {
		var elem Node
		err = rows.Scan(&elem.id, &elem.dtype, &elem.state, &elem.wflow, &elem.name, &elem.ntype)
		if err != nil {
			return nil, err
		}
		elem.nfunc = defNodeFunc
		ary = append(ary, &elem)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ary, nil
}

// Get retrieves the requested node from the database.
func (ns *_Nodes) Get(id NodeID) (*Node, error) {
	if id <= 0 {
		return nil, errors.New("node ID must be a positive integer")
	}

	var elem Node
	q := `
	SELECT id, doctype_id, docstate_id, workflow_id, name, type
	FROM wf_workflow_nodes
	WHERE id = ?
	`
	row := db.QueryRow(q, id)
	err := row.Scan(&elem.id, &elem.dtype, &elem.state, &elem.wflow, &elem.name, &elem.ntype)
	if err != nil {
		return nil, err
	}

	elem.nfunc = defNodeFunc
	return &elem, nil
}

// GetByState retrieves the requested node from the database, as per
// the document state specification.
func (ns *_Nodes) GetByState(dtype DocTypeID, state DocStateID) (*Node, error) {
	var elem Node
	q := `
	SELECT id, doctype_id, docstate_id, workflow_id, name, type
	FROM wf_workflow_nodes
	WHERE doctype_id = ?
	AND docstate_id = ?
	`
	row := db.QueryRow(q, dtype, state)
	err := row.Scan(&elem.id, &elem.dtype, &elem.state, &elem.wflow, &elem.name, &elem.ntype)
	if err != nil {
		return nil, err
	}

	elem.nfunc = defNodeFunc
	return &elem, nil
}
