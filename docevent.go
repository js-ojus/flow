// (c) Copyright 2015 JONNALAGADDA Srinivas
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

import "time"

// DocEvent represents a user action performed on a document in the
// system.
//
// Together with documents, events are central to the workflow engine
// in `flow`.  Events cause documents to switch from one state to
// another, usually in response to user actions.  They also carry
// information of the modification to the document.
type DocEvent struct {
	doc         *Document
	user        *User // user causing this modification
	mtime       time.Time
	text        string    // comment or other content
	newState    *DocState // result of the modification
	newRevision uint16    // serves as a cross-check
}
