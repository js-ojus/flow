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
	"math"
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
	ID         WorkflowID `json:"ID,omitempty"`     // Globally-unique identifier of this workflow
	Name       string     `json:"Name,omitempty"`   // Globally-unique name of this workflow
	DocType    DocType    `json:"DocType"`          // Document type of which this workflow defines the life cycle
	BeginState DocState   `json:"BeginState"`       // Where this flow begins
	Active     bool       `json:"Active,omitempty"` // Is this workflow enabled?
}

// ApplyEvent takes an input user action or a system event, and
// applies its document action to the given document.  This results in
// a possibly new document state.  This method also prepares a message
// that is posted to applicable mailboxes.
func (w *Workflow) ApplyEvent(otx *sql.Tx, event *DocEvent, recipients []GroupID) (DocStateID, error) {
	if !w.Active {
		return 0, ErrWorkflowInactive
	}
	if event.Status == EventStatusApplied {
		return 0, ErrDocEventAlreadyApplied
	}
	if w.DocType.ID != event.DocType {
		return 0, ErrDocEventDocTypeMismatch
	}

	n, err := Nodes.GetByState(w.DocType.ID, event.State)
	if err != nil {
		return 0, err
	}

	var gt string
	tq := `SELECT group_type FROM wf_groups_master WHERE id = ?`
	row := db.QueryRow(tq, event.Group)
	err = row.Scan(&gt)
	if err != nil {
		return 0, err
	}
	if gt != "S" {
		return 0, errors.New("group must be singleton")
	}

	var tx *sql.Tx
	if otx == nil {
		tx, err = db.Begin()
		if err != nil {
			return 0, err
		}
		defer tx.Rollback()
	} else {
		tx = otx
	}

	nstate, err := n.applyEvent(tx, event, recipients)
	if err != nil {
		return 0, err
	}

	if otx == nil {
		err = tx.Commit()
		if err != nil {
			return 0, err
		}
	}

	return nstate, nil
}

// Unexported type, only for convenience methods.
type _Workflows struct{}

// Workflows provides a resource-like interface to the workflows
// defined in the system.
var Workflows _Workflows

// New creates and initialises a workflow definition using the given
// name, the document type whose life cycle this workflow should
// manage, and the initial document state in which this workflow
// begins.
//
// N.B.  Workflow names must be globally-unique.
func (_Workflows) New(otx *sql.Tx, name string, dtype DocTypeID, state DocStateID) (WorkflowID, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, errors.New("name should not be empty")
	}
	if dtype <= 0 {
		return 0, errors.New("document type should be a positive integer")
	}
	if state <= 1 {
		return 0, errors.New("initial document state should be an integer > 1")
	}

	var tx *sql.Tx
	var err error
	if otx == nil {
		tx, err = db.Begin()
		if err != nil {
			return 0, err
		}
		defer tx.Rollback()
	} else {
		tx = otx
	}

	q := `
	INSERT INTO wf_workflows(name, doctype_id, docstate_id, active)
	VALUES(?, ?, ?, 1)
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

// List answers a subset of the workflows defined in the system,
// according to the given specification.
//
// Result set begins with ID >= `offset`, and has not more than
// `limit` elements.  A value of `0` for `offset` fetches from the
// beginning, while a value of `0` for `limit` fetches until the end.
func (_Workflows) List(offset, limit int64) ([]*Workflow, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.New("offset and limit must be non-negative integers")
	}
	if limit == 0 {
		limit = math.MaxInt64
	}

	q := `
	SELECT wf.id, wf.name, dtm.id, dtm.name, dsm.id, dsm.name, wf.active
	FROM wf_workflows wf
	JOIN wf_doctypes_master dtm ON wf.doctype_id = dtm.id
	JOIN wf_docstates_master dsm ON wf.docstate_id = dsm.id
	ORDER BY wf.id
	LIMIT ? OFFSET ?
	`
	rows, err := db.Query(q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ary := make([]*Workflow, 0, 10)
	for rows.Next() {
		var elem Workflow
		err = rows.Scan(&elem.ID, &elem.Name, &elem.DocType.ID, &elem.DocType.Name,
			&elem.BeginState.ID, &elem.BeginState.Name, &elem.Active)
		if err != nil {
			return nil, err
		}
		ary = append(ary, &elem)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ary, nil
}

// Get retrieves the details of the requested workflow from the
// database.
//
// N.B.  This method retrieves the primary information of the
// workflow.  Information of the nodes comprising this workflow have
// to be fetched separately.
func (_Workflows) Get(id WorkflowID) (*Workflow, error) {
	q := `
	SELECT wf.id, wf.name, dtm.id, dtm.name, dsm.id, dsm.name, wf.active
	FROM wf_workflows wf
	JOIN wf_doctypes_master dtm ON dtm.id = wf.doctype_id
	JOIN wf_docstates_master dsm ON dsm.id = wf.docstate_id
	WHERE wf.id = ?
	`
	row := db.QueryRow(q, id)
	var elem Workflow
	err := row.Scan(&elem.ID, &elem.Name, &elem.DocType.ID, &elem.DocType.Name,
		&elem.BeginState.ID, &elem.BeginState.Name, &elem.Active)
	if err != nil {
		return nil, err
	}

	return &elem, nil
}

// GetByDocType retrieves the details of the requested workflow from
// the database.
//
// N.B.  This method retrieves the primary information of the
// workflow.  Information of the nodes comprising this workflow have
// to be fetched separately.
func (_Workflows) GetByDocType(dtid DocTypeID) (*Workflow, error) {
	q := `
	SELECT wf.id, wf.name, dtm.id, dtm.name, dsm.id, dsm.name, wf.active
	FROM wf_workflows wf
	JOIN wf_doctypes_master dtm ON dtm.id = wf.doctype_id
	JOIN wf_docstates_master dsm ON dsm.id = wf.docstate_id
	WHERE wf.doctype_id = ?
	`
	row := db.QueryRow(q, dtid)
	var elem Workflow
	err := row.Scan(&elem.ID, &elem.Name, &elem.DocType.ID, &elem.DocType.Name,
		&elem.BeginState.ID, &elem.BeginState.Name, &elem.Active)
	if err != nil {
		return nil, err
	}

	return &elem, nil
}

// GetByName retrieves the details of the requested workflow from the
// database.
//
// N.B.  This method retrieves the primary information of the
// workflow.  Information of the nodes comprising this workflow have
// to be fetched separately.
func (_Workflows) GetByName(name string) (*Workflow, error) {
	q := `
	SELECT wf.id, wf.name, dtm.id, dtm.name, dsm.id, dsm.name, wf.active
	FROM wf_workflows wf
	JOIN wf_doctypes_master dtm ON wf.doctype_id = dtm.id
	JOIN wf_docstates_master dsm ON wf.docstate_id = dsm.id
	WHERE wf.name = ?
	`
	row := db.QueryRow(q, name)
	var elem Workflow
	err := row.Scan(&elem.ID, &elem.Name, &elem.DocType.ID, &elem.DocType.Name,
		&elem.BeginState.ID, &elem.BeginState.Name, &elem.Active)
	if err != nil {
		return nil, err
	}

	return &elem, nil
}

// Rename assigns a new name to the given workflow.
func (_Workflows) Rename(otx *sql.Tx, id WorkflowID, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("name should be non-empty")
	}

	var tx *sql.Tx
	var err error
	if otx == nil {
		tx, err = db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()
	} else {
		tx = otx
	}

	q := `
	UPDATE wf_workflows SET name = ?
	WHERE id = ?
	`
	_, err = tx.Exec(q, name, id)
	if err != nil {
		return err
	}

	if otx == nil {
		err = tx.Commit()
		if err != nil {
			return err
		}
	}

	return nil
}

// SetActive sets the status of the workflow as either active or
// inactive, helping in workflow management and deprecation.
func (_Workflows) SetActive(otx *sql.Tx, id WorkflowID, active bool) error {
	var tx *sql.Tx
	var err error
	if otx == nil {
		tx, err = db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()
	} else {
		tx = otx
	}

	var flag int
	if active {
		flag = 1
	}
	q := `
	UPDATE wf_workflows SET active = ?
	WHERE id = ?
	`
	_, err = tx.Exec(q, flag, id)
	if err != nil {
		return err
	}

	if otx == nil {
		err = tx.Commit()
		if err != nil {
			return err
		}
	}

	return nil
}

// AddNode maps the given document state to the specified node.  This
// map is consulted by the workflow when performing a state transition
// of the system.
func (_Workflows) AddNode(otx *sql.Tx, dtype DocTypeID, state DocStateID,
	ac AccessContextID, wid WorkflowID, name string, ntype NodeType) (NodeID, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, errors.New("name should not be empty")
	}

	var tx *sql.Tx
	var err error
	if otx == nil {
		tx, err = db.Begin()
		if err != nil {
			return 0, err
		}
		defer tx.Rollback()
	} else {
		tx = otx
	}

	q := `
	INSERT INTO wf_workflow_nodes(doctype_id, docstate_id, ac_id, workflow_id, name, type)
	VALUES(?, ?, ?, ?, ?, ?)
	`
	res, err := tx.Exec(q, dtype, state, ac, wid, name, string(ntype))
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

	return NodeID(id), nil
}

// RemoveNode unmaps the given document state to the specified node.
// This map is consulted by the workflow when performing a state
// transition of the system.
func (_Workflows) RemoveNode(otx *sql.Tx, wid WorkflowID, nid NodeID) error {
	var tx *sql.Tx
	var err error
	if otx == nil {
		tx, err = db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()
	} else {
		tx = otx
	}

	q := `
	DELETE FROM wf_workflow_nodes
	WHERE workflow_id = ?
	AND id = ?
	`
	_, err = tx.Exec(q, wid, nid)
	if err != nil {
		return err
	}

	if otx == nil {
		err = tx.Commit()
		if err != nil {
			return err
		}
	}

	return nil
}
