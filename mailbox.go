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
)

// Mailbox is the message delivery destination for both action and
// informational messages.
//
// Both users and groups have mailboxes.  In all normal cases, a
// message is 'consumed' by the recipient.  Messages can be moved into
// and out of mailboxes to facilitate reassignments under specific or
// extraordinary conditions.
type Mailbox struct {
	GroupID `json:"GroupID"` // Group (or singleton user) owner of this mailbox
}

// _Mailboxes provides a resource-like interface to group virtual
// mailboxes.
type _Mailboxes struct {
	// Intentionally left blank.
}

// Mailboxes is the singleton instance of `_Mailboxes`.
var Mailboxes _Mailboxes

// Count answers the number of messages in the given group's virtual
// mailbox. Specifying `true` for `unread` fetches a count of unread
// messages.
func (_Mailboxes) Count(gid GroupID, unread bool) (int64, error) {
	if gid <= 0 {
		return 0, errors.New("group ID should be a positive integer")
	}

	q := `
	SELECT COUNT(id)
	FROM wf_mailboxes
	WHERE group_id = ?
	`
	if unread {
		q += `AND unread = 1`
	}

	row := db.QueryRow(q, gid)
	var n int64
	err := row.Scan(&n)
	if err != nil {
		return 0, err
	}

	return n, nil
}

// List answers a list of the messages in the given group's virtual
// mailbox, as per the given specification.
//
// Result set begins with ID >= `offset`, and has not more than
// `limit` elements.  A value of `0` for `offset` fetches from the
// beginning, while a value of `0` for `limit` fetches until the end.
func (_Mailboxes) List(gid GroupID, offset, limit int64, unread bool) ([]*Message, error) {
	if gid <= 0 {
		return nil, errors.New("group ID should be a positive integer")
	}
	if offset < 0 || limit < 0 {
		return nil, errors.New("offset and limit must be non-negative integers")
	}
	if limit == 0 {
		limit = math.MaxInt64
	}

	q := `
	SELECT msgs.id, msgs.doctype_id, msgs.doc_id, msgs.docevent_id, msgs.title, msgs.data
	FROM wf_messages msgs
	JOIN wf_mailboxes mbs ON mbs.message_id = msgs.id
	WHERE mbs.group_id = ?
	`
	if unread {
		q = q + "AND mbs.unread = 1"
	}
	q = q + `
	ORDER BY msgs.id
	LIMIT ? OFFSET ?
	`

	rows, err := db.Query(q, gid, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ary := make([]*Message, 0, 10)
	for rows.Next() {
		var elem Message
		err = rows.Scan(&elem.ID, &elem.DocType, &elem.DocID, &elem.Event, &elem.Title, &elem.Data)
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

// ReassignMessage removes the message with the given ID from its
// current mailbox, and delivers it to the given other group's
// mailbox.
func (_Mailboxes) ReassignMessage(otx *sql.Tx, cgid, ngid GroupID, msgID MessageID) error {
	if cgid <= 0 || ngid <= 0 || msgID <= 0 {
		return errors.New("all identifiers should be positive integers")
	}

	var tx *sql.Tx
	if otx == nil {
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()
	} else {
		tx = otx
	}

	q := `
	UPDATE wf_mailboxes SET group_id = ?, unread = 1
	WHERE group_id = ?
	AND message_id = ?
	`
	_, err := tx.Exec(q, ngid, cgid, msgID)
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
