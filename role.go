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

// RoleID is the type of unique role identifiers.
type RoleID int64

// Role represents a collection of privileges.
//
// Each user in the system has one or more roles assigned.
type Role struct {
	id    RoleID                  // globally-unique ID of this role
	name  string                  // name of this role
	perms map[DocType][]DocAction // actions allowed to perform on each document type

	mutex sync.RWMutex
}

// GetRole loads the role object corresponding to the given role ID
// from the database, and answers that.
func GetRole(rid RoleID) (*Role, error) {
	// TODO(js): implement
	return nil, nil
}

// ID answers this role's identifier.
func (r *Role) ID() RoleID {
	return r.id
}

// Name answers this role's name.
func (r *Role) Name() string {
	return r.name
}

// AddPermissions adds the given actions to this role, for the given
// document type.
func (r *Role) AddPermissions(dt DocType, das []DocAction) error {
	if len(das) == 0 {
		return errors.New("list of permitted actions should not be empty")
	}

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
	if len(das) == 0 {
		return errors.New("list of permitted actions should not be empty")
	}

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

// HasPermission answers `true` if this role has the queried
// permission for the given document type.
func (r *Role) HasPermission(dt DocType, da DocAction) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	das, ok := r.perms[dt]
	if !ok {
		return false
	}

	for _, el := range das {
		if el == da {
			return true
		}
	}
	return false
}
