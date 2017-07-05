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
import "sort"
import "sync"

// Group represents a specified collection of users.  A user belongs
// to zero or more groups.
type Group struct {
	id    uint16   // globally-unique ID
	name  string   // globally-unique name
	users []uint64 // users included in this group

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
	return g, nil
}

// AddUser includes the given user in this group
func (g *Group) AddUser(u uint64) bool {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	return g.addUser(u)
}

// addUser includes the given user in this group
func (g *Group) addUser(u uint64) bool {
	idx := sort.Search(len(g.users), func(i int) bool { return g.users[i] >= u })
	if idx < len(g.users) && g.users[idx] == u {
		return false
	}

	g.users = append(g.users, 0)
	copy(g.users[idx+1:], g.users[idx:])
	g.users[idx] = u
	return true
}

// RemoveUser removes the given user from this group.
func (g *Group) RemoveUser(u uint64) bool {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	return g.removeUser(u)
}

// removeUser removes the given user from this group.
func (g *Group) removeUser(u uint64) bool {
	idx := sort.Search(len(g.users), func(i int) bool { return g.users[i] >= u })
	if idx < len(g.users) && g.users[idx] == u {
		return false
	}

	g.users = append(g.users[:idx], g.users[idx+1:]...)
	return true
}

// AddGroup includes all the users in the given group to this group.
func (g *Group) AddGroup(other *Group) bool {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	ret := true
	for _, u := range other.users {
		if ok := g.addUser(u); !ok {
			ret = false
		}
	}

	return ret
}

// RemoveGroup removes all the users in the given group from this
// group.
func (g *Group) RemoveGroup(other *Group) bool {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	ret := true
	for _, u := range other.users {
		if ok := g.removeUser(u); !ok {
			ret = false
		}
	}

	return ret
}
