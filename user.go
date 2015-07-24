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

// User represents any kind of a user invoking or otherwise
// participating in a defined workflow in the system.
//
// User details are expected to be provided by an external identity
// provider application or directory.  `flow` neither defines nor
// manages users.
type User struct {
	id     uint64   // must be globally-unique
	name   string   // for display purposes only
	active bool     // status of the user account
	roles  []*Role  // all roles assigned to this user
	groups []*Group // all groups this user is a part of
}

// NewUser instantiates a user instance in the system.
//
// In most cases, this should be done upon the corresponding user's
// first action in the system.
func NewUser(id uint64, name string) (*User, error) {
	if id == 0 || name == "" {
		return nil, fmt.Errorf("invalid user data -- id: %d, name: %s", id, name)
	}

	u := &User{id: id, name: name}
	u.roles = make([]*Role, 1)
	u.groups = make([]*Group, 1)
	return u, nil
}

// ID answers this user's ID.
func (u *User) ID() uint64 {
	return u.id
}

// Name answers this user's name.
func (u *User) Name() string {
	return u.name
}

// Active answers if this user's account is enabled.
func (u *User) Active() bool {
	return u.active
}

// SetStatus can be used to enable or disable a user account
// dynamically.
func (u *User) SetStatus(active bool) {
	u.active = active
}

// AssignRole adds the given role to this user, if it is not assigned
// already.
func (u *User) AssignRole(r *Role) bool {
	for _, el := range u.roles {
		if el.id == r.id {
			return false
		}
	}

	u.roles = append(u.roles, r)
	return true
}

// UnassignRole removes the given role from this user, if it is
// assigned currently.
func (u *User) UnassignRole(r *Role) bool {
	found := false
	idx := -1
	for i, el := range u.roles {
		if el.id == r.id {
			found = true
			idx = i
			break
		}
	}
	if !found {
		return false
	}

	u.roles = append(u.roles[:idx], u.roles[idx+1:]...)
	return true
}

// Roles answers a copy of the roles currently assigned to this user.
func (u *User) Roles() []*Role {
	rs := make([]*Role, len(u.roles))
	copy(rs, u.roles)
	return rs
}

// AddToGroup records that this user participates in the given group,
// if (s)he does not already.
func (u *User) AddToGroup(g *Group) bool {
	for _, el := range u.groups {
		if el.id == g.id {
			return false
		}
	}

	u.groups = append(u.groups, g)
	return true
}

// RemoveFromGroup records that this user no longer participates in
// the given group, if (s)he does currently.
func (u *User) RemoveFromGroup(g *Group) bool {
	found := false
	idx := -1
	for i, el := range u.groups {
		if el.id == g.id {
			found = true
			idx = i
			break
		}
	}
	if !found {
		return false
	}

	u.groups = append(u.groups[:idx], u.groups[idx+1:]...)
	return true
}

// Groups answers a copy of the groups to which this user currently
// belongs.
func (u *User) Groups() []*Group {
	gs := make([]*Group, len(u.groups))
	copy(gs, u.groups)
	return gs
}
