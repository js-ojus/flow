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

// Group represents a specified collection of users.  A user belongs
// to zero or more groups.
type Group struct {
	id          uint64              // globally-unique ID
	name        string              // globally-unique name
	users       map[uint64]struct{} // users included in this group
	isUserGroup bool                // is this a user-specific group?

	mutex sync.RWMutex
}

// NewGroup creates and initialises a group.
//
// Usually, all available groups should be loaded during system
// initialization.  Only groups created during runtime should be added
// dynamically.
func NewGroup(name string) (*Group, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("group name should not be empty")
	}

	g := &Group{name: name}
	g.users = make(map[uint64]struct{})
	return g, nil
}

// newUserGroup creates and initialises a group that is exclusive to
// the given user.
//
// Usually, all available groups should be loaded during system
// initialization.  Only groups created during runtime should be added
// dynamically.
func newUserGroup(name string, u uint64) (*Group, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("group name should not be empty")
	}
	if u == 0 {
		return nil, errors.New("user ID should be a positive integer")
	}

	g := &Group{name: name, isUserGroup: true}
	g.users = make(map[uint64]struct{})
	return g, nil
}

// ID answers this group's identifier.
func (g *Group) ID() uint64 {
	return g.id
}

// Name answers this group's name.
func (g *Group) Name() string {
	return g.name
}

// IsUserGroup answers `true` if this group was auto-created as the
// native group of a user account; `false` otherwise.
func (g *Group) IsUserGroup() bool {
	return g.isUserGroup
}

// AddUser includes the given user in this group.
//
// Answers `true` if the user was not already included in this group;
// `false` otherwise.
func (g *Group) AddUser(u uint64) bool {
	if g.isUserGroup {
		return false
	}

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
	if g.isUserGroup {
		return false
	}

	g.mutex.Lock()
	defer g.mutex.Unlock()

	if _, ok := g.users[u]; !ok {
		return false
	}

	delete(g.users, u)
	return true
}

// Users answers a copy of the list of users included in this group,
// as their IDs.
func (g *Group) Users() []uint64 {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	us := make([]uint64, 0, len(g.users))
	for u := range g.users {
		us = append(us, u)
	}
	return us
}

// HasUser answers `true` if this group includes the given user;
// `false` otherwise.
func (g *Group) HasUser(u uint64) bool {
	_, ok := g.users[u]
	return ok
}

// AddGroup includes all the users in the given group to this group.
//
// Answers `true` if at least one user from the other group did not
// already exist in this group; `false` otherwise.
func (g *Group) AddGroup(other *Group) bool {
	if g.isUserGroup {
		return false
	}

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
	if g.isUserGroup {
		return false
	}

	g.mutex.Lock()
	defer g.mutex.Unlock()

	l1 := len(g.users)
	for u := range other.users {
		delete(g.users, u)
	}
	l2 := len(g.users)
	return l2 < l1
}
