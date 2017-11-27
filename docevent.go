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
	"math"
	"strings"
	"time"
)

// EventStatus enumerates the query parameter values for filtering by
// event state.
type EventStatus uint8

const (
	// EventStatusAll does not filter events.
	EventStatusAll EventStatus = iota
	// EventStatusApplied selects only those events that have been successfully applied.
	EventStatusApplied
	// EventStatusPending selects only those events that are pending application.
	EventStatusPending
)

// DocEventID is the type of unique document event identifiers.
type DocEventID int64

// DocEvent represents a user action performed on a document in the
// system.
//
// Together with documents and nodes, events are central to the
// workflow engine in `flow`.  Events cause documents to transition
// from one state to another, usually in response to user actions.  It
// is possible for system events to cause state transitions, as well.
type DocEvent struct {
	ID      DocEventID  `json:"ID"`        // Unique ID of this event
	DocType DocTypeID   `json:"DocType"`   // Document type of the document to which this event is to be applied
	DocID   DocumentID  `json:"DocID"`     // Document to which this event is to be applied
	State   DocStateID  `json:"DocState"`  // Current state of the document must equal this
	Action  DocActionID `json:"DocAction"` // Action performed by the user
	Group   GroupID     `json:"Group"`     // Group (singleton) who caused this action
	Text    string      `json:"Text"`      // Comment or other content
	Ctime   time.Time   `json:"Ctime"`     // Time at which the event occurred
	Status  EventStatus `json:"Status"`    // Status of this event
}

// StatusInDB answers the status of this event.
func (e *DocEvent) StatusInDB() (EventStatus, error) {
	var dstatus string
	row := db.QueryRow("SELECT status FROM wf_docevents WHERE id = ?", e.ID)
	err := row.Scan(&dstatus)
	if err != nil {
		return 0, err
	}
	switch dstatus {
	case "A":
		e.Status = EventStatusApplied

	case "P":
		e.Status = EventStatusPending

	default:
		return 0, fmt.Errorf("unknown event status : %s", dstatus)
	}

	return e.Status, nil
}

// Unexported type, only for convenience methods.
type _DocEvents struct{}

// DocEvents exposes a resource-like interface to document events.
var DocEvents _DocEvents

// DocEventsNewInput holds information needed to create a new document
// event in the system.
type DocEventsNewInput struct {
	DocTypeID          // Type of the document; required
	DocumentID         // Unique identifier of the document; required
	DocStateID         // Document must be in this state for this event to be applied; required
	DocActionID        // Action performed by `Group`; required
	GroupID            // Group (user) who performed the action that raised this event; required
	Text        string // Any comments or notes
}

// New creates and initialises an event that transforms the document
// that it refers to.
func (_DocEvents) New(otx *sql.Tx, input *DocEventsNewInput) (DocEventID, error) {
	if input.DocumentID <= 0 {
		return 0, errors.New("document ID should be a positive integer")
	}
	if input.Text == "" {
		return 0, errors.New("please add comments or notes")
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
	INSERT INTO wf_docevents(doctype_id, doc_id, docstate_id, docaction_id, group_id, data, ctime, status)
	VALUES(?, ?, ?, ?, ?, ?, NOW(), 'P')
	`
	res, err := tx.Exec(q, input.DocTypeID, input.DocumentID, input.DocStateID, input.DocActionID, input.GroupID, input.Text)
	if err != nil {
		return 0, err
	}
	var id int64
	id, err = res.LastInsertId()
	if err != nil {
		return 0, err
	}

	if otx == nil {
		err = tx.Commit()
		if err != nil {
			return 0, err
		}
	}

	return DocEventID(id), nil
}

// DocEventsListInput specifies a set of filter conditions to narrow
// down document listings.
type DocEventsListInput struct {
	DocTypeID                   // Events on documents of this type are listed
	AccessContextID             // Access context from within which to list
	GroupID                     // List events created by this (singleton) group
	DocStateID                  // List events acting on this state
	CtimeStarting   time.Time   // List events created after this time
	CtimeBefore     time.Time   // List events created before this time
	Status          EventStatus // List events that are in this state of application
}

// List answers a subset of document events, based on the input
// specification.
//
// `status` should be one of `all`, `applied` and `pending`.
//
// Result set begins with ID >= `offset`, and has not more than
// `limit` elements.  A value of `0` for `offset` fetches from the
// beginning, while a value of `0` for `limit` fetches until the end.
func (_DocEvents) List(input *DocEventsListInput, offset, limit int64) ([]*DocEvent, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.New("offset and limit must be non-negative integers")
	}
	if limit == 0 {
		limit = math.MaxInt64
	}

	// Base query.

	q := `
	SELECT de.id, de.doctype_id, de.doc_id, de.docstate_id, de.docaction_id, de.group_id, de.data, de.ctime, de.status
	FROM wf_docevents de
	`

	// Process input specification.

	where := []string{}
	args := []interface{}{}

	if input.AccessContextID > 0 {
		tbl := DocTypes.docStorName(input.DocTypeID)
		q += `JOIN ` + tbl + ` docs ON docs.id = de.doc_id
		`
		where = append(where, `docs.ac_id = ?`)
		args = append(args, input.AccessContextID)
	}

	switch input.Status {
	case EventStatusAll:
		// Intentionally left blank

	case EventStatusApplied:
		where = append(where, `status = 'A'`)

	case EventStatusPending:
		where = append(where, `status = 'P'`)

	default:
		return nil, fmt.Errorf("unknown event status specified in filter : %d", input.Status)
	}

	if input.GroupID > 0 {
		where = append(where, `de.group_id = ?`)
		args = append(args, input.GroupID)
	}

	if input.DocStateID > 0 {
		where = append(where, `de.docstate_id = ?`)
		args = append(args, input.DocStateID)
	}

	if !input.CtimeStarting.IsZero() {
		where = append(where, `de.ctime >= ?`)
		args = append(args, input.CtimeStarting)
	}

	if !input.CtimeBefore.IsZero() {
		where = append(where, `de.ctime < ?`)
		args = append(args, input.CtimeBefore)
	}

	if len(where) > 0 {
		q += ` WHERE ` + strings.Join(where, ` AND `)
	}

	q += `
	ORDER BY de.id
	LIMIT ? OFFSET ?
	`
	args = append(args, limit, offset)
	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var text sql.NullString
	var dstatus string
	ary := make([]*DocEvent, 0, 10)
	for rows.Next() {
		var elem DocEvent
		err = rows.Scan(&elem.ID, &elem.DocType, &elem.DocID, &elem.State, &elem.Action, &elem.Group, &text, &elem.Ctime, &dstatus)
		if err != nil {
			return nil, err
		}
		if text.Valid {
			elem.Text = text.String
		}
		switch dstatus {
		case "A":
			elem.Status = EventStatusApplied

		case "P":
			elem.Status = EventStatusPending

		default:
			return nil, fmt.Errorf("unknown event status : %s", dstatus)
		}
		ary = append(ary, &elem)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ary, nil
}

// Get retrieves a document event from the database, using the given
// event ID.
func (_DocEvents) Get(eid DocEventID) (*DocEvent, error) {
	if eid <= 0 {
		return nil, errors.New("event ID should be a positive integer")
	}

	var text sql.NullString
	var dstatus string
	var elem DocEvent
	q := `
	SELECT id, doctype_id, doc_id, docstate_id, docaction_id, group_id, data, ctime, status
	FROM wf_docevents
	WHERE id = ?
	`
	row := db.QueryRow(q, eid)
	err := row.Scan(&elem.ID, &elem.DocType, &elem.DocID, &elem.State, &elem.Action, &elem.Group, &text, &elem.Ctime, &dstatus)
	if err != nil {
		return nil, err
	}
	if text.Valid {
		elem.Text = text.String
	}
	switch dstatus {
	case "A":
		elem.Status = EventStatusApplied

	case "P":
		elem.Status = EventStatusPending

	default:
		return nil, fmt.Errorf("unknown event status : %s", dstatus)
	}

	return &elem, nil
}
