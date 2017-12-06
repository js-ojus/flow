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
	"time"
)

// MessageID is the type of unique identifiers of messages.
type MessageID int64

// Message is the content part of a notification sent by the workflow
// engine to possibly multiple mailboxes.
//
// Messages can be informational or seek action.  Each message
// contains a reference to the document that began the current
// workflow, as well as the event that triggered this message.
type Message struct {
	ID      MessageID        `json:"ID"` // Globally-unique identifier of this message
	DocType `json:"DocType"` // Document type of the associated document
	DocID   DocumentID       `json:"DocID"`    // Document in the workflow
	Event   DocEventID       `json:"DocEvent"` // Event that triggered this message
	Title   string           `json:"Title"`    // Subject of this message
	Data    string           `json:"Data"`     // Body of this message
}

// Notification tracks the 'unread' status of a message in a mailbox.
//
// Since a single message can be delivered to multiple mailboxes, the
// 'unread' status cannot be associated with a message.  Instead,
// `Notification` is the entity that tracks it per mailbox.
type Notification struct {
	GroupID `json:"Group"`   // The group whose mailbox this notification is in
	Message `json:"Message"` // The underlying message
	Unread  bool             `json:"Unread"` // Status flag reflecting if the message is still not read
	Ctime   time.Time        `json:"Ctime"`  // Time when this notification was posted
}
