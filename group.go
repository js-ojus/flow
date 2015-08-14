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

// Group represents a specified collection of users.
//
// A user belongs to zero or more groups.  Groups can have associated
// privileges, too.
type Group struct {
	id    uint16
	name  string
	privs []*Privilege
}

// NewGroup creates and initialises a group.
//
// Usually, all available groups should be loaded during system
// initialization.  Only groups created during runtime should be added
// dynamically.
func NewGroup(id uint16, name string) (*Group, error) {
	if id == 0 || name == "" {
		return nil, fmt.Errorf("invalid group data -- id: %d, name: %s", id, name)
	}

	g := &Group{id: id, name: name}
	g.privs = make([]*Privilege, 0, 1)
	return g, nil
}

// AddPrivilege includes the given privilege in the set of privileges
// assigned to this group.
func (g *Group) AddPrivilege(p *Privilege) bool {
	for _, el := range g.privs {
		if el.IsOnSameTargetAs(p) {
			return false
		}
	}

	g.privs = append(g.privs, p)
	return true
}

// RemovePrivilegesOn removes the privileges that this group has on
// the given target.
func (g *Group) RemovePrivilegesOn(res *Resource, doc *Document) bool {
	found := false
	idx := -1
	for i, el := range g.privs {
		if el.IsOnTarget(res, doc) {
			found = true
			idx = i
			break
		}
	}
	if !found {
		return false
	}

	g.privs = append(g.privs[:idx], g.privs[idx+1:]...)
	return true
}

// ReplacePrivilege any current privilege on the given target, with
// the given privilege.
func (g *Group) ReplacePrivilege(p *Privilege) bool {
	if !g.RemovePrivilegesOn(p.resource, p.doc) {
		return false
	}

	g.privs = append(g.privs, p)
	return true
}
