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
	"errors"
	"strings"
)

// Workflow represents the entire life cycle of a single document.
//
// A workflow begins with the creation of a document, and drives its
// life cycle through a sequence of responses to user actions or other
// system events.
//
// The engine in `flow` is visible primarily through workflows,
// documents and their behaviour.
//
// Currently, the topology of workflows is a graph, and is determined
// by the node definitions herein.
//
// N.B. It is highly recommended, but not necessary, that workflow
// names be defined in a system of hierarchical namespaces.
type Workflow struct {
	name  string             // globally-unique name of this workflow
	dtype DocType            // document type of which this workflow defines the life cycle
	nodes map[DocState]*Node // all nodes defined in this workflow
	begin DocState           // where this flow begins
}

// NewWorkflow creates and initialises a workflow definition.
func NewWorkflow(name string, dtype DocType, bstate DocState) (*Workflow, error) {
	name = strings.TrimSpace(name)
	tdtype := strings.TrimSpace(string(dtype))
	if name == "" || tdtype == "" {
		return nil, errors.New("workflow name and document type cannot be empty")
	}

	w := &Workflow{name: name, dtype: DocType(tdtype)}
	w.nodes = make(map[DocState]*Node, 2)
	w.begin = bstate
	return w, nil
}

// Name answers this workflow definition's name.
func (w *Workflow) Name() string {
	return w.name
}

// DocType answers the document type for which this defines the
// workflow.
func (w *Workflow) DocType() DocType {
	return w.dtype
}

// BeginState answers the document state in which the execution of
// this workflow begins.
func (w *Workflow) BeginState() DocState {
	return w.begin
}

// AddNode maps the given document state to the specified node.  This
// map is consulted by the workflow when performing a state transition
// of the system.
func (w *Workflow) AddNode(state DocState, node *Node) error {
	if node == nil {
		return errors.New("node should not be `nil`")
	}
	if n, ok := w.nodes[state]; ok {
		return errors.New("state already mapped to node : " + n.name)
	}

	w.nodes[state] = node
	return nil
}
