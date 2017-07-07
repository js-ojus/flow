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
	"sync"
)

// AccessContext is a namespace in which groups are mapped to their
// permissions.  This mapping happens through roles.
type AccessContext struct {
	id    uint16             // globally-unique identifier
	name  string             // globally-unique namespace; can be a department, project, location, branch, etc.
	perms map[uint64][]*Role // map of roles assigned to different groups, in this context

	mutex sync.RWMutex
}

// NewAccessContext creates a new access context that determines how
// the workflows that operate in its context run.
//
// Each workflow that operates on a document type is given an
// associated access context.  This context is used to determine
// workflow routing, possible branching and rendezvous points.
func NewAccessContext(name string) (*AccessContext, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("access context name should not be empty")
	}

	return &AccessContext{name: name}, nil
}

// ID answers this access context's unique ID.
func (ac *AccessContext) ID() uint16 {
	return ac.id
}

// Name answers this access context's (unique) namespace.
func (ac *AccessContext) Name() string {
	return ac.name
}

// AddGroupRole assigns the specified role to the given group, if it
// is not already assigned.
func (ac *AccessContext) AddGroupRole(gr uint64, r *Role) error {
	if gr == 0 || r == nil {
		return errors.New("group ID must be a positive integer; role should not be `nil`")
	}

	ac.mutex.Lock()
	defer ac.mutex.Unlock()

	rs, ok := ac.perms[gr]
	if !ok {
		rs = []*Role{r}
		ac.perms[gr] = rs
		return nil
	}

	for _, el := range rs {
		if el.name == r.name {
			return nil
		}
	}
	rs = append(rs, r)
	ac.perms[gr] = rs
	return nil
}

// RemoveGroupRole unassigns the specified role from the given group.
// func (ac *AccessContext)
func (ac *AccessContext) RemoveGroupRole(gr uint64, r *Role) error {
	if gr == 0 || r == nil {
		return errors.New("group ID must be a positive integer; role should not be `nil`")
	}

	ac.mutex.Lock()
	defer ac.mutex.Unlock()

	rs, ok := ac.perms[gr]
	if !ok {
		return nil
	}

	idx := -1
	for i, el := range rs {
		if el.name == r.name {
			idx = i
			break
		}
	}
	if idx == -1 {
		return nil
	}

	rs = append(rs[:idx], rs[idx+1:]...)
	ac.perms[gr] = rs
	return nil
}

// HasPermission answers `true` if the given group has the requested
// action enabled on the specified document type; `false` otherwise.
func (ac *AccessContext) HasPermission(gr uint64, dt DocType, da DocAction) bool {
	if gr == 0 {
		return false
	}

	ac.mutex.RLock()
	defer ac.mutex.RUnlock()

	rs, ok := ac.perms[gr]
	if !ok {
		return false
	}

	for _, el := range rs {
		if el.HasPermission(dt, da) {
			return true
		}
	}
	return false
}
