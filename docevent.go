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
	id     DocEventID  // Unique ID of this event
	dtype  DocTypeID   // Document type of the document to which this event is to be applied
	docID  DocumentID  // Document to which this event is to be applied
	state  DocStateID  // Current state of the document must equal this
	action DocActionID // Action performed by the user
	user   UserID      // User who caused this action
	text   string      // Comment or other content
	ctime  time.Time   // Time at which the event occurred
	status EventStatus // Status of this event
}

// ID answers the unique identifier of this document event.
func (e *DocEvent) ID() DocEventID {
	return e.id
}

// DocumentType answers the document type of the document to which
// this event is to be applied.
func (e *DocEvent) DocumentType() DocTypeID {
	return e.dtype
}

// Document answers the document to which this event is to be applied.
func (e *DocEvent) Document() DocumentID {
	return e.docID
}

// State answers the state of the document as of the time this event
// occurred.
func (e *DocEvent) State() DocStateID {
	return e.state
}

// Action answers the document action that this event represents.
func (e *DocEvent) Action() DocActionID {
	return e.action
}

// User answers the user who caused this event to occur.
func (e *DocEvent) User() UserID {
	return e.user
}

// Text answers the comment or other content enclosed in this event.
func (e *DocEvent) Text() string {
	return e.text
}

// Time answers the time at which this event occurred.
func (e *DocEvent) Time() time.Time {
	return e.ctime
}

// Status answers the status of this event.
func (e *DocEvent) Status() (EventStatus, error) {
	var dstatus string
	row := db.QueryRow("SELECT status FROM wf_docevents WHERE id = ?", e.id)
	err := row.Scan(&dstatus)
	if err != nil {
		return 0, err
	}
	switch dstatus {
	case "A":
		e.status = EventStatusApplied

	case "P":
		e.status = EventStatusPending

	default:
		return 0, fmt.Errorf("unknown event status : %s", dstatus)
	}

	return e.status, nil
}

// Unexported type, only for convenience methods.
type _DocEvents struct{}

var _docevents *_DocEvents

func init() {
	_docevents = &_DocEvents{}
}

// New creates and initialises an event that transforms the document
// that it refers to.
func (des *_DocEvents) New(otx *sql.Tx, user UserID, dtype DocTypeID, did DocumentID,
	state DocStateID, action DocActionID, text string) (DocEventID, error) {
	if user <= 0 {
		return 0, errors.New("user ID should be a positive integer")
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
	INSERT INTO wf_docevents(doctype_id, doc_id, docstate_id, docaction_id, user_id, data, ctime, status)
	VALUES(?, ?, ?, ?, ?, NOW(), 'P')
	`
	res, err := tx.Exec(q, dtype, did, state, action, text)
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

// List answers a subset of document events, based on the input
// specification.
//
// `status` should be one of `all`, `applied` and `pending`.
//
// Result set begins with ID >= `offset`, and has not more than
// `limit` elements.  A value of `0` for `offset` fetches from the
// beginning, while a value of `0` for `limit` fetches until the end.
func (des *_DocEvents) List(status EventStatus, offset, limit int64) ([]*DocEvent, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.New("offset and limit must be non-negative integers")
	}
	if limit == 0 {
		limit = math.MaxInt64
	}

	q := `
	SELECT *
	FROM wf_docevents
	`
	switch status {
	case EventStatusAll:
		// Intentionally left blank

	case EventStatusApplied:
		q = q + `
		WHERE status = 'A'
		`

	case EventStatusPending:
		q = q + `
		WHERE status = 'P'
		`

	default:
		return nil, fmt.Errorf("unknown event status specified in filter : %d", status)
	}
	q = q + `
	ORDER BY id
	LIMIT ? OFFSET ?
	`
	rows, err := db.Query(q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var text sql.NullString
	var dstatus string
	ary := make([]*DocEvent, 0, 10)
	for rows.Next() {
		var elem DocEvent
		err = rows.Scan(&elem.id, &elem.dtype, &elem.docID, &elem.state, &elem.action, &elem.user, &text, &elem.ctime, &dstatus)
		if err != nil {
			return nil, err
		}
		if text.Valid {
			elem.text = text.String
		}
		switch dstatus {
		case "A":
			elem.status = EventStatusApplied

		case "P":
			elem.status = EventStatusPending

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
func (des *_DocEvents) Get(eid DocEventID) (*DocEvent, error) {
	if eid <= 0 {
		return nil, errors.New("event ID should be a positive integer")
	}

	var text sql.NullString
	var dstatus string
	var elem DocEvent
	row := db.QueryRow("SELECT * FROM wf_docevents WHERE id = ?", eid)
	err := row.Scan(&elem.id, &elem.dtype, &elem.docID, &elem.state, &elem.action, &elem.user, &text, &elem.ctime, &dstatus)
	if err != nil {
		return nil, err
	}
	if text.Valid {
		elem.text = text.String
	}
	switch dstatus {
	case "A":
		elem.status = EventStatusApplied

	case "P":
		elem.status = EventStatusPending

	default:
		return nil, fmt.Errorf("unknown event status : %s", dstatus)
	}

	return &elem, nil
}
