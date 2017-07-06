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

import "sync"

// Group represents a specified collection of users.  A user belongs
// to zero or more groups.
type Group struct {
	id    uint16              // globally-unique ID
	name  string              // globally-unique name
	users map[uint64]struct{} // users included in this group

	mutex sync.RWMutex
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
	g.users = make(map[uint64]struct{})
	return g, nil
}

// AddUser includes the given user in this group.
//
// Answers `true` if the user was not already included in this group;
// `false` otherwise.
func (g *Group) AddUser(u uint64) bool {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if _, ok := g.users[u]; ok {
		return false
	}

	g.users[u] = struct{}{}
	return true
}

// RemoveUser removes the given user from this group.
//
// Answers `true` if the user was removed from this group now; `false`
// if the user was not a part of this group.
func (g *Group) RemoveUser(u uint64) bool {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if _, ok := g.users[u]; !ok {
		return false
	}

	delete(g.users, u)
	return true
}

// AddGroup includes all the users in the given group to this group.
//
// Answers `true` if at least one user from the other group did not
// already exist in this group; `false` otherwise.
func (g *Group) AddGroup(other *Group) bool {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	l1 := len(g.users)
	for u := range other.users {
		g.users[u] = struct{}{}
	}
	l2 := len(g.users)
	return l2 > l1
}

// RemoveGroup removes all the users in the given group from this
// group.
//
// Answers `true` if at least one user from the other group existed in
// this group; `false` otherwise.
func (g *Group) RemoveGroup(other *Group) bool {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	l1 := len(g.users)
	for u := range other.users {
		delete(g.users, u)
	}
	l2 := len(g.users)
	return l2 < l1
}
