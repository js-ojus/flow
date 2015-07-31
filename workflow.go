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
	"fmt"
	"sync"
)

// WfDefinition holds the definition of a specific workflow.  A single
// definition can be used to instantiate any number of workflow
// instances.
//
// Currently, the topology is a graph, and is determined by the return
// values of the node functions in the node definitions herein.
type WfDefinition struct {
	ns    string                     // globally-unique namespace identifier
	name  string                     // unique within its namespace
	nodes map[string]*NodeDefinition // all nodes defined in this workflow
	begin *NodeDefinition            // where this flow begins
}

// NewWfDefinition creates and initialises a workflow definition.
func NewWfDefinition(ns, name string) (*WfDefinition, error) {
	if ns == "" || name == "" {
		return nil, fmt.Errorf("invalid initialisation data -- namespace: %s, name: %s", ns, name)
	}

	wd := &WfDefinition{ns: ns, name: name}
	wd.nodes = make(map[string]*NodeDefinition, 2)
	return wd, nil
}

// Namespace answers this workflow definition's globally-unique
// namespace.
func (wd *WfDefinition) Namespace() string {
	return wd.ns
}

// Name answers this workflow definition's name.
func (wd *WfDefinition) Name() string {
	return wd.name
}

// AddNodeDefn adds the given node definition to this workflow, if it
// is not already included.
func (wd *WfDefinition) AddNodeDefn(name string, ntype NodeType, nfunc NodeFunc) (*NodeDefinition, error) {
	if name == "" || nfunc == nil {
		return nil, fmt.Errorf("invalid initialisation data -- workflow: (%s, %s), name: %s, nfunc: %v", wd.ns, wd.name, name, nfunc)
	}

	if _, ok := wd.nodes[name]; ok {
		return nil, fmt.Errorf("node definition already exists: %s", name)
	}

	nd := newNodeDefinition(wd, name, ntype, nfunc)
	wd.nodes[name] = nd
	return nd, nil
}

// SetBeginNode sets the node definition having the given name as the
// node at which an instance of this workflow begins execution.
func (wd *WfDefinition) SetBeginNode(name string) error {
	if name == "" {
		return fmt.Errorf("empty node name given")
	}

	if nd, ok := wd.nodes[name]; ok {
		wd.begin = nd
		return nil
	}

	return fmt.Errorf("node definition not registered: %s", name)
}

// BeginNodeDefn answers the definition of the node where this
// workflow has to begin.
func (wd *WfDefinition) BeginNodeDefn() *NodeDefinition {
	return wd.begin
}

// NodeDefinition answers the definition having the given name.
func (wd *WfDefinition) NodeDefinition(name string) *NodeDefinition {
	if name == "" {
		return nil
	}

	if nd, ok := wd.nodes[name]; ok {
		return nd
	}
	return nil
}

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
