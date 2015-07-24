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

// Role represents a collection of privileges.
//
// Each user in the system has one or more roles assigned.
type Role struct {
	id    uint16
	name  string
	privs []*Privilege
}

// NewRole creates and initialises a role.
//
// Usually, all available roles should be loaded during system
// initialization.  Only roles created during runtime should be added
// dynamically.
func NewRole(id uint16, name string) (*Role, error) {
	if id == 0 || name == "" {
		return nil, fmt.Errorf("invalid role data -- id: %d, name: %s", id, name)
	}

	r := &Role{id: id, name: name}
	r.privs = make([]*Privilege, 1)
	return r, nil
}

// AddPrivilege includes the given privilege in the set of privileges
// assigned to this role.
func (r *Role) AddPrivilege(p *Privilege) bool {
	for _, el := range r.privs {
		if el.IsOnSameTargetAs(p) {
			return false
		}
	}

	r.privs = append(r.privs, p)
	return true
}

// RemovePrivilegesOn removes the privileges that this role has on the
// given target.
func (r *Role) RemovePrivilegesOn(res *Resource, doc *Document) bool {
	found := false
	idx := -1
	for i, el := range r.privs {
		if el.IsOnTarget(res, doc) {
			found = true
			idx = i
			break
		}
	}
	if !found {
		return false
	}

	r.privs = append(r.privs[:idx], r.privs[idx+1:]...)
	return true
}

// ReplacePrivilege any current privilege on the given target, with
// the given privilege.
func (r *Role) ReplacePrivilege(p *Privilege) bool {
	if !r.RemovePrivilegesOn(p.resource, p.doc) {
		return false
	}

	r.privs = append(r.privs, p)
	return true
}
