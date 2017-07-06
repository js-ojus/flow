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

// Role represents a collection of privileges.
//
// Each user in the system has one or more roles assigned.
type Role struct {
	id    uint16                  // globally-unique ID of this role
	name  string                  // name of this role
	perms map[DocType][]DocAction // actions allowed to perform on each document type

	mutex sync.RWMutex
}

// NewRole creates and initialises a role.
//
// Usually, all available roles should be loaded during system
// initialization.  Only roles created during runtime should be added
// dynamically.
func NewRole(name string) (*Role, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("role name should not be empty")
	}

	r := &Role{name: name}
	r.perms = make(map[DocType][]DocAction)
	return r, nil
}

// AddPermissions adds the given actions to this role, for the given
// document type.
func (r *Role) AddPermissions(dt DocType, das []DocAction) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, ok := r.perms[dt]; ok {
		return errors.New("given document type already has permissions set")
	}

	r.perms[dt] = append([]DocAction{}, das...)
	return nil
}

// RemovePermissions removes all permissions from this role, for the
// given document type.
func (r *Role) RemovePermissions(dt DocType) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, ok := r.perms[dt]; !ok {
		return errors.New("given document type does not have any permissions set")
	}

	delete(r.perms, dt)
	return nil
}

// UpdatePermissions updates the set of permissions this role has, for
// the given document type.
func (r *Role) UpdatePermissions(dt DocType, das []DocAction) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, ok := r.perms[dt]; !ok {
		return errors.New("given document type does not have any permissions set")
	}

	r.perms[dt] = nil
	delete(r.perms, dt)
	r.perms[dt] = append([]DocAction{}, das...)
	return nil
}

// Permissions answers the current set of permissions this role has,
// for the given document type.  It answers `nil` in case the given
// document type does not have any permissions set in this role.
func (r *Role) Permissions(dt DocType) []DocAction {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if as, ok := r.perms[dt]; ok {
		return append([]DocAction{}, as...)
	}

	return nil
}
