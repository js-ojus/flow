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

// Error defines `flow`-specific errors, and satisfies the `error`
// interface.
type Error string

// Error implements the `error` interface.
func (e Error) Error() string {
	return string(e)
}

//

const (
	// ErrUnknown : unknown internal error
	ErrUnknown = Error("ErrUnknown : unknown internal error")

	// ErrDocEventRedundant : another equivalent event has already effected this action
	ErrDocEventRedundant = Error("ErrDocEventRedundant : another equivalent event has already applied this action")
	// ErrDocEventDocTypeMismatch : document's type does not match event's type
	ErrDocEventDocTypeMismatch = Error("ErrDocEventDocTypeMismatch : document's type does not match event's type")
	// ErrDocEventStateMismatch : document's state does not match event's state
	ErrDocEventStateMismatch = Error("ErrDocEventStateMismatch : document's state does not match event's state")
	// ErrDocEventAlreadyApplied : event already applied; nothing to do
	ErrDocEventAlreadyApplied = Error("ErrDocEventAlreadyApplied : event already applied; nothing to do")

	// ErrDocumentNoParent : document is a root document
	ErrDocumentNoParent = Error("ErrDocumentNoParent : document is a root document")
	// ErrDocumentIsChild : cannot have its own state, title or tags
	ErrDocumentIsChild = Error("ErrDocumentIsChild : cannot have its own state, title or tags")

	// ErrWorkflowInactive : this workflow is currently inactive
	ErrWorkflowInactive = Error("ErrWorkflowInactive : this workflow is currently inactive")
	// ErrWorkflowInvalidAction : given action cannot be performed on this document's current state
	ErrWorkflowInvalidAction = Error("ErrWorkflowInvalidAction : given action cannot be performed on this document's current state")

	// ErrMessageNoRecipients : list of recipients is empty
	ErrMessageNoRecipients = Error("ErrMessageNoRecipients : list of recipients is empty")
)
