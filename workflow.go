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

import (
	"sync"
	"time"
)

// Workflow represents the entire life cycle of a single document.
//
// A workflow begins with the creation of a document, and drives its
// life cycle through a sequence of responses to user actions or other
// external events.
//
// The engine in `flow` is visible primarily through workflows,
// documents and their behaviour.
type Workflow struct {
	id   uint64        // globally-unique workflow instance ID
	defn *WfDefinition // definition of this flow

	mutex     sync.Mutex
	node      *Node   // current node in the workflow
	path      []*Node // flow so far, tracked in order
	completed bool
	err       error // reason for failure, if aborted
}

// newWorkflow creates and initialises a workflow instance using the
// given definition.
func newWorkflow(wd *WfDefinition) *Workflow {
	// WARNING: In a truly busy application, this manner of generating
	// IDs could lead to clashes.
	t := time.Now().UTC().UnixNano()
	return &Workflow{id: uint64(t), defn: wd, path: make([]*Node, 0, 2)}
}

// moveToNode adds the first node to the path of this workflow, as a
// completed step of processing.  It then marks the second as the
// current node in the workflow.
//
// If `n2` is `nil`, then the workflow is deemed to have completed.
func (wf *Workflow) moveToNode(n1, n2 *Node) bool {
	for _, el := range wf.path {
		if el.id == n1.id {
			return false
		}
	}

	wf.mutex.Lock()
	defer wf.mutex.Unlock()

	wf.path = append(wf.path, n1)
	wf.node = n2
	if n2 == nil {
		wf.completed = true
	}
	return true
}

// abort marks this workflow as aborted, and records the error
// message.
func (wf *Workflow) abort(err error) {
	if err == nil {
		return
	}

	wf.mutex.Lock()
	defer wf.mutex.Unlock()

	wf.completed = true
	wf.err = err
}

// complete marks this workflow as completed normally.  This method
// should be used only when a manual intervention halts this workflow.
func (wf *Workflow) complete() {
	wf.mutex.Lock()
	defer wf.mutex.Unlock()

	wf.node = nil
	wf.completed = true
}
