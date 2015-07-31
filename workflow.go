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

import "sync"

// Workflow represents the entire life cycle of a single document.
//
// A workflow begins with the creation of a document, and drives its
// life cycle through a sequence of responses to user actions or other
// external events.
//
// The engine in `flow` is visible primarily through workflows,
// documents and their behaviour.
type Workflow struct {
	id uint64 // globally-unique workflow instance ID

	mutex     sync.Mutex
	node      *Node   // current node in the workflow
	path      []*Node // flow so far, tracked in order
	completed bool
	err       error // reason for failure, if aborted
}
