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
	"fmt"
	"sync"
	"time"
)

// Mailbox is the message delivery destination for both action and
// informational messages.
//
// Both users and groups have mailboxes.  In all normal cases, a
// message is 'consumed' by the recipient.  Messages can be moved into
// and out of mailboxes to facilitate reassignments under specific or
// extraordinary conditions.
type Mailbox struct {
	id    uint64 // globally-unique
	user  *User  // in case of an individual user
	group *Group // in case of a group; does not receive broadcast messages

	mutex sync.RWMutex
	msgs  []*Message // messages awaiting user consumption
}

// newMailbox creates and sets up a mailbox for the given user or
// group.
func newMailbox(u *User, grp *Group) (*Mailbox, error) {
	if u == nil && grp == nil {
		return nil, fmt.Errorf("neither user nor group given")
	}

	t := time.Now().UTC().UnixNano()
	if u != nil {
		return &Mailbox{id: uint64(t), user: u, msgs: make([]*Message, 0, 1)}, nil
	}
	return &Mailbox{id: uint64(t), group: grp, msgs: make([]*Message, 0, 1)}, nil
}

// TODO(js): `OpenMailbox()`

// ID answers this mailbox's ID.
func (mb *Mailbox) ID() uint64 {
	return mb.id
}

// User answers the recipient user of this mailbox.
func (mb *Mailbox) User() *User {
	return mb.user
}

// Group answers the recipient group of this mailbox.
func (mb *Mailbox) Group() *Group {
	return mb.group
}

// MsgCount answers the number of messages in this mailbox.
func (mb *Mailbox) MsgCount() int {
	mb.mutex.RLock()
	defer mb.mutex.RUnlock()

	return len(mb.msgs)
}

// recvMsg delivers the given message into this mailbox.
func (mb *Mailbox) recvMsg(msg *Message) error {
	if msg == nil {
		return fmt.Errorf("`nil` message given")
	}

	mb.mutex.Lock()
	defer mb.mutex.Unlock()

	for _, el := range mb.msgs {
		if el.id == msg.id {
			return nil
		}
	}

	mb.msgs = append(mb.msgs, msg)
	return nil
}

// reassignMsg removes the message with the given ID from this
// mailbox, and delivers it to the given other mailbox.
func (mb *Mailbox) reassignMsg(msgID uint64, other *Mailbox) error {
	if msgID == 0 || other == nil {
		return fmt.Errorf("invalid inputs -- message ID: %d, mailbox: %v", msgID, other)
	}

	// TODO(js): this can deadlock currently; resolution needed

	mb.mutex.Lock()
	defer mb.mutex.Unlock()

	idx := -1
	var msg *Message
	for i, el := range mb.msgs {
		if el.id == msgID {
			idx = i
			msg = el
			break
		}
	}
	if msg == nil {
		return fmt.Errorf("message not found -- ID: %d", msgID)
	}

	other.mutex.Lock()
	defer other.mutex.Unlock()

	mb.msgs = append(mb.msgs[:idx], mb.msgs[idx+1:]...)
	other.msgs = append(other.msgs, msg)
	return nil
}
