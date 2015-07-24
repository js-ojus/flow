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

import "fmt"

// Resource represents a collection of documents of a given type.
//
// Resources are the first line of targets for actions and assignment
// of privileges.
type Resource struct {
	id        uint16 // globally-unique identifier
	name      string // convenient name
	endpoint  string // globally-unique end point path
	namespace string // optional
}

// NewResource instantiates a resource.
func NewResource(id uint16, name, ep, ns string) (*Resource, error) {
	if id == 0 || name == "" || ep == "" {
		return nil, fmt.Errorf("invalid resource information")
	}

	return &Resource{id, name, ep, ns}, nil
}

// ID answers this resource's globally-unique ID.
func (r *Resource) ID() uint16 {
	return r.id
}

// Name answers this resource's name.
func (r *Resource) Name() string {
	return r.name
}

// EndPoint answers this resource's globally-unique endpoint.
func (r *Resource) EndPoint() string {
	return r.endpoint
}

// Namespace answers this resource's namespace, if any.
func (r *Resource) Namespace() string {
	return r.namespace
}
