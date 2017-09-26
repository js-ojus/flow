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
	ID      MessageID  `json:"id"`       // Globally-unique identifier of this message
	DocType DocTypeID  `json:"docType"`  // Document type of the associated document
	DocID   DocumentID `json:"docID"`    // Document in the workflow
	Event   DocEventID `json:"docEvent"` // Event that triggered this message
	Title   string     `json:"title"`    // Subject of this message
	Data    string     `json:"data"`     // Body of this message
}
