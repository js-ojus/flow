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
	"log"
)

// DocState is one of a set of enumerated states for a document, as
// defined by the consuming application.
//
// `flow`, therefore, does not assume anything about the specifics of
// any state.  Applications can, for example, embed `DocState` in a
// struct that provides more context.
//
// Document states should be loaded during application initialisation.
type DocState struct {
	dtype      DocType // for namespace purposes
	name       string
	successors []*DocState // possible next states
}

// NewDocState creates an enumerated state as defined by the consuming
// application.
func NewDocState(dtype DocType, name string) (*DocState, error) {
	if string(dtype) == "" || name == "" {
		return nil, fmt.Errorf("invalid initialisation data -- type: %s, name: %s", dtype, name)
	}

	ds := &DocState{dtype: dtype, name: name}
	ds.successors = make([]*DocState, 0, 1)
	return ds, nil
}

// Type answers the document type for which this defines a state.
func (s *DocState) Type() DocType {
	return s.dtype
}

// Name answers this state's name.
func (s *DocState) Name() string {
	return s.name
}

// AddSuccessor adds a possible successor state for this document
// state, if it is not already defined.
func (s *DocState) AddSuccessor(ds *DocState) bool {
	if ds.dtype != s.dtype {
		log.Printf("mismatched DocType -- current: %s, given: %s", s.dtype, ds.dtype)
		return false
	}

	for _, el := range s.successors {
		if el.name == ds.name {
			return false
		}
	}

	s.successors = append(s.successors, ds)
	return true
}

// Successors answers a copy of this state's possible successor
// states.
func (s *DocState) Successors() []*DocState {
	ds := make([]*DocState, 0, len(s.successors))
	copy(ds, s.successors)
	return ds
}
