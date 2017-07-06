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
	groups []*Group // all groups this user is a part of

	mutex sync.RWMutex
}

// NewUser instantiates a user instance in the system.
//
// In most cases, this should be done upon the corresponding user's
// first action in the system.
func NewUser(name string) (*User, error) {
	if name == "" {
		return nil, errors.New("user name should not be empty")
	}

	u := &User{name: name}
	u.groups = make([]*Group, 0, 1)
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

// IsActive answers if this user's account is enabled.
func (u *User) IsActive() bool {
	return u.active
}

// SetStatus can be used to enable or disable a user account
// dynamically.
func (u *User) SetStatus(active bool) {
	u.active = active

	// TODO(js): Persist this.
}

// Groups answers a copy of the groups to which this user currently
// belongs.
func (u *User) Groups() []*Group {
	u.mutex.RLock()
	defer u.mutex.RUnlock()

	gs := make([]*Group, 0, len(u.groups))
	copy(gs, u.groups)
	return gs
}
