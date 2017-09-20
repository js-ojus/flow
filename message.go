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

// MessageID is the type of unique identifiers of messages.
type MessageID int64

// Message is the notification sent by the workflow engine to possibly
// multiple mailboxes.
//
// Messages can be informational or seek action.  Each message
// contains a reference to the document that began the current
// workflow, as well as the event that triggered this message.
type Message struct {
	id    MessageID  // Globally-unique identifier of this message
	dtype DocTypeID  // Document type of the associated document
	docID DocumentID // Document in the workflow
	event DocEventID // Event that triggered this message
	title string     // Subject of this message
	data  string     // Body of this message
}

// ID answers the unique identifier of this message.
func (m *Message) ID() MessageID {
	return m.id
}

// Document answers the document type and identifier of the document
// in whose context this message was generated.
func (m *Message) Document() (DocTypeID, DocumentID) {
	return m.dtype, m.docID
}

// Event answers the event that triggered the generation of this
// message.
func (m *Message) Event() DocEventID {
	return m.event
}

// Title answers the subject of this message.
func (m *Message) Title() string {
	return m.title
}

// Data answers the body of this message.
func (m *Message) Data() string {
	return m.data
}
