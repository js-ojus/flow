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
	"sync"
)

// GroupID is the type of unique group identifiers.
type GroupID int64

// Group represents a specified collection of users.  A user belongs
// to zero or more groups.
//
// Group details are expected to be provided by an external identity
// provider application or directory.  `flow` neither defines nor
// manages groups.
type Group struct {
	id        GroupID // Globally-unique ID
	name      string  // Globally-unique name
	groupType string  // Is this a user-specific group? Etc.

	mutex sync.RWMutex
}

// GetGroup initialises the group by reading from database.
//
// Usually, all available groups should be loaded during system
// initialization.  Only groups created during runtime should be added
// dynamically.
func GetGroup(gid GroupID) (*Group, error) {
	if gid <= 0 {
		return nil, errors.New("group ID should be a positive integer")
	}

	var tid GroupID
	var name string
	var gtype string
	row := db.QueryRow("SELECT id, name, group_type FROM groups_master WHERE id = ?", gid)
	err := row.Scan(&tid, &name, &gtype)
	if err != nil {
		return nil, err
	}

	g := &Group{id: gid, name: name, groupType: gtype}
	return g, nil
}

// ID answers this group's identifier.
func (g *Group) ID() GroupID {
	return g.id
}

// Name answers this group's name.
func (g *Group) Name() string {
	return g.name
}

// GroupType answers the nature of this group.  For instance, if this
// group was auto-created as the native group of a user account, or is
// a collection of users, etc.
func (g *Group) GroupType() string {
	return g.groupType
}

// HasUser answers `true` if this group includes the given user;
// `false` otherwise.
func (g *Group) HasUser(uid UserID) (bool, error) {
	var count int64
	row := db.QueryRow("SELECT COUNT(*) FROM group_user WHERE group_id = ? AND user_id = ? LIMIT 1", g.id, uid)
	err := row.Scan(&count)
	if err != nil {
		return false, err
	}

	if count == 0 {
		return false, nil
	}
	return true, nil
}
