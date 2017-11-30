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
	"log"
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
		DocType: DocType{
			ID: d.DocType.ID,
		},
		DocID: d.ID,
		Event: event.ID,
		Title: d.Title,
		Data:  event.Text,
	}
}

// Node represents a specific logical unit of processing and routing
// in a workflow.
type Node struct {
	ID       NodeID          `json:"ID"`                      // Unique identifier of this node
	DocType  DocTypeID       `json:"DocType"`                 // Document type which this node's workflow manages
	State    DocStateID      `json:"DocState"`                // A document arriving at this node must be in this state
	AccCtx   AccessContextID `json:"AccessContext,omitempty"` // Specific access context associated with this state, if any
	Wflow    WorkflowID      `json:"Workflow"`                // Containing flow of this node
	Name     string          `json:"Name"`                    // Unique within its workflow
	NodeType NodeType        `json:"NodeType"`                // Topology type of this node
	nfunc    NodeFunc        // Processing function of this node
}

// Transitions answers the possible document states into which a
// document currently in the given state can transition.
func (n *Node) Transitions() (map[DocActionID]DocStateID, error) {
	return DocTypes._Transitions(n.DocType, n.State)
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
	if err != nil {
		return 0, err
	}
	tstate, ok := ts[event.Action]
	if !ok {
		return 0, ErrWorkflowInvalidAction
	}

	// Check document's current state.
	doc, err := Documents.Get(otx, event.DocType, event.DocID)
	if err != nil {
		return 0, err
	}
	if doc.State.ID != event.State {
		return 0, ErrDocEventStateMismatch
	}

	// Document has already transitioned.  So, we note that the event
	// is applied, and return.
	//
	// N.B. This has implications for `NodeTypeJoinAny` below.  Should
	// you alter this logic or its position, verify that the
	// corresponding logic in the switch below is in coherence.
	if doc.State.ID == tstate {
		err = n.recordEvent(otx, event, tstate, true)
		if err != nil {
			return 0, err
		}
		return tstate, ErrDocEventRedundant
	}

	// Transition document state according to the target node type.

	tnode, err := Nodes.GetByState(n.DocType, tstate)
	if err != nil {
		return 0, err
	}

	switch tnode.NodeType {
	case NodeTypeJoinAny:
		// Multiple 'in's, but any one suffices.

		// We have already checked to see if the document has
		// transitioned into the target state.  If we have come this
		// far, the event can be applied.
		fallthrough

	case NodeTypeBegin, NodeTypeEnd, NodeTypeLinear, NodeTypeBranch:
		// Any node type having a single 'in'.

		// Update the document to transition the state.
		err = Documents.setState(otx, event.DocType, event.DocID, tstate, tnode.AccCtx)
		if err != nil {
			return 0, err
		}

		// Record event application.
		err = n.recordEvent(otx, event, tstate, false)
		if err != nil {
			return 0, err
		}

		// Post messages.
		msg := n.nfunc(doc, event)
		if len(recipients) == 0 {
			recipients, err = tnode.determineRecipients(otx, event.Group)
			if err != nil {
				return 0, err
			}
		}
		// It is legal to not have any recipients, too.
		if len(recipients) > 0 {
			err = n.postMessage(otx, msg, recipients)
			if err != nil {
				return 0, err
			}
		}

	case NodeTypeJoinAll:
		// Multiple 'in's, and all are required.

		// TODO(js)

	default:
		log.Panicf("unknown node type encountered : %s\n", tnode.NodeType)
	}

	return tstate, nil
}

// recordEvent writes a record stating that the given event has
// successfully been applied to effect a document state transition.
func (n *Node) recordEvent(otx *sql.Tx, event *DocEvent, tstate DocStateID, statusOnly bool) error {
	if !statusOnly {
		q := `
		INSERT INTO wf_docevent_application(doctype_id, doc_id, from_state_id, docevent_id, to_state_id)
		VALUES(?, ?, ?, ?, ?)
		`
		_, err := otx.Exec(q, event.DocType, event.DocID, event.State, event.ID, tstate)
		if err != nil {
			return err
		}
	}

	q := `UPDATE wf_docevents SET status = 'A' WHERE id = ?`
	_, err := otx.Exec(q, event.ID)
	if err != nil {
		return err
	}

	return nil
}

// determineRecipients takes the document type and access context into
// account, and determines the list of groups to which the
// notification should be posted.
func (n *Node) determineRecipients(otx *sql.Tx, group GroupID) ([]GroupID, error) {
	q := `
	SELECT reports_to
	FROM wf_ac_group_hierarchy
	WHERE ac_id = ?
	AND group_id = ?
	ORDER BY group_id
	LIMIT 1
	`
	rows, err := otx.Query(q, n.AccCtx, group)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ary := make([]GroupID, 0, 4)
	for rows.Next() {
		var gid int64
		err = rows.Scan(&gid)
		if err != nil {
			return nil, err
		}
		ary = append(ary, GroupID(gid))
	}
	if rows.Err() != nil {
		return nil, err
	}
	return ary, nil
}

// postMessage posts the given message into the mailboxes of the
// specified recipients.
func (n *Node) postMessage(otx *sql.Tx, msg *Message, recipients []GroupID) error {
	// Record the message.

	q := `
	INSERT INTO wf_messages(doctype_id, doc_id, docevent_id, title, data)
	VALUES(?, ?, ?, ?, ?)
	`
	res, err := otx.Exec(q, msg.DocType.ID, msg.DocID, msg.Event, msg.Title, msg.Data)
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

// Nodes provides a resource-like interface to the nodes defined in
// this system.
var Nodes _Nodes

// List answers a list of the nodes comprising the given workflow.
func (_Nodes) List(id WorkflowID) ([]*Node, error) {
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
		err = rows.Scan(&elem.ID, &elem.DocType, &elem.State, &elem.Wflow, &elem.Name, &elem.NodeType)
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
func (_Nodes) Get(id NodeID) (*Node, error) {
	if id <= 0 {
		return nil, errors.New("node ID must be a positive integer")
	}

	var elem Node
	var acID sql.NullInt64
	q := `
	SELECT id, doctype_id, docstate_id, ac_id, workflow_id, name, type
	FROM wf_workflow_nodes
	WHERE id = ?
	`
	row := db.QueryRow(q, id)
	err := row.Scan(&elem.ID, &elem.DocType, &elem.State, &acID, &elem.Wflow, &elem.Name, &elem.NodeType)
	if err != nil {
		return nil, err
	}
	if acID.Valid {
		elem.AccCtx = AccessContextID(acID.Int64)
	}

	elem.nfunc = defNodeFunc
	return &elem, nil
}

// GetByState retrieves the requested node from the database, as per
// the document state specification.
func (_Nodes) GetByState(dtype DocTypeID, state DocStateID) (*Node, error) {
	var elem Node
	var acID sql.NullInt64
	q := `
	SELECT id, doctype_id, docstate_id, ac_id, workflow_id, name, type
	FROM wf_workflow_nodes
	WHERE doctype_id = ?
	AND docstate_id = ?
	`
	row := db.QueryRow(q, dtype, state)
	err := row.Scan(&elem.ID, &elem.DocType, &elem.State, &acID, &elem.Wflow, &elem.Name, &elem.NodeType)
	if err != nil {
		return nil, err
	}
	if acID.Valid {
		elem.AccCtx = AccessContextID(acID.Int64)
	}

	elem.nfunc = defNodeFunc
	return &elem, nil
}
