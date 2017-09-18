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
	"strings"
)

// WorkflowID is the type of unique workflow identifiers.
type WorkflowID int64

// Workflow represents the entire life cycle of a single document.
//
// A workflow begins with the creation of a document, and drives its
// life cycle through a sequence of responses to user actions or other
// system events.
//
// The engine in `flow` is visible primarily through workflows,
// documents and their behaviour.
//
// Currently, the topology of workflows is a graph, and is determined
// by the node definitions herein.
//
// N.B. It is highly recommended, but not necessary, that workflow
// names be defined in a system of hierarchical namespaces.
type Workflow struct {
	id     WorkflowID // Globally-unique identifier of this workflow
	name   string     // Globally-unique name of this workflow
	dtype  DocTypeID  // Document type of which this workflow defines the life cycle
	bstate DocStateID // Where this flow begins
}

// ID answers the unique identifier of this workflow.
func (w *Workflow) ID() WorkflowID {
	return w.id
}

// Name answers the globally-unique name of this workflow.
func (w *Workflow) Name() string {
	return w.name
}

// DocType answers the document type for which this defines the
// workflow.
func (w *Workflow) DocType() DocTypeID {
	return w.dtype
}

// BeginState answers the document state in which the execution of
// this workflow begins.
func (w *Workflow) BeginState() DocStateID {
	return w.bstate
}

// Unexported type, only for convenience methods.
type _Workflows struct{}

var _workflows *_Workflows

func init() {
	_workflows = &_Workflows{}
}

// Workflows provides a resource-like interface to the workflows
// defined in this system.
func Workflows() *_Workflows {
	return _workflows
}

// New creates and initialises a workflow definition using the given
// name, the document type whose life cycle this workflow should
// manage, and the initial document state in which this workflow
// begins.
//
// N.B.  Workflow names must be globally-unique.
func (ws *_Workflows) New(otx *sql.Tx, name string, dtype DocTypeID, state DocStateID) (WorkflowID, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, errors.New("name should not be empty")
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

	q := `
	INSERT INTO wf_workflows(name, doctype_id, docstate_id)
	VALUES(?, ?, ?)
	`
	res, err := tx.Exec(q, name, dtype, state)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	if otx == nil {
		err = tx.Commit()
		if err != nil {
			return 0, err
		}
	}

	return WorkflowID(id), nil
}

// AddNode maps the given document state to the specified node.  This
// map is consulted by the workflow when performing a state transition
// of the system.
func (ws *_Workflows) AddNode(otx *sql.Tx, wid WorkflowID, name string,
	ntype NodeType, state DocStateID, nstates []DocStateID) (NodeID, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, errors.New("name should not be empty")
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

	q := `
	INSERT INTO wf_workflow_nodes(workflow_id, name, type, docstate_id)
	VALUES(?, ?, ?, ?)
	`
	res, err := tx.Exec(q, wid, name, ntype, state)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	q = `
	INSERT INTO wf_node_next_states(node_id, docstate_id)
	VALUES(?, ?)
	`
	for _, ns := range nstates {
		_, err := tx.Exec(q, id, ns)
		if err != nil {
			return 0, err
		}
	}

	if otx == nil {
		err = tx.Commit()
		if err != nil {
			return 0, err
		}
	}

	return NodeID(id), nil
}

// ApplyEvent takes an input user action or a system event, and
// applies its document action to the given document.  This results in
// a possibly new document state.  In addition, a registered
// processing function is invoked on the document to prepare a message
// that can be posted to applicable mailboxes.
//
// Internally, the workflow delegates this method to the appropriate
// node, if one such is registered.
func (w *Workflow) ApplyEvent(event *DocEvent, args ...interface{}) (DocStateID, error) {
	doc, err := _documents.Get(event.dtype, event.docID)
	if err != nil {
		return 0, err
	}
	if doc.state != event.state {
		return 0, fmt.Errorf("document state is : %s, but event is targeting state : %s", doc.state.name, event.state.name)
	}

	// TODO(js): implement this

	return 0, nil
}
