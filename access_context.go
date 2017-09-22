// (c) Copyright 2015-2017 JONNALAGADDA Srinivas
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
	"log"
	"sync"
)

// AccessContext is a namespace in which groups are mapped to their
// permissions.  This mapping happens through roles.
type AccessContext struct {
	nsID  int64                // globally-unique namespace; can be a department, project, location, branch, etc.
	perms map[GroupID][]RoleID // map of roles assigned to different groups, in this context

	mutex sync.RWMutex
}

// NewAccessContext creates a new in-memory access context.
func NewAccessContext(nsID int64) (*AccessContext, error) {
	if nsID <= 0 {
		return nil, errors.New("namespace ID must be a positive integer")
	}

	ac := &AccessContext{nsID: nsID}
	ac.perms = make(map[GroupID][]RoleID)
	return ac, nil
}

// GetAccessContext fetches the requested access context that
// determines how the workflows that operate in its context run.
//
// Each workflow that operates on a document type is given an
// associated access context.  This context is used to determine
// workflow routing, possible branching and rendezvous points.
//
// For complex organisational requirements, the name of the access
// context can be made hierarchical with a suitable delimiter.  For
// example:
//     IN:south:HYD:BR-101
//     sbu-08/client-0249/prj-006348
func GetAccessContext(nsID int64) (*AccessContext, error) {
	if nsID <= 0 {
		return nil, errors.New("namespace ID should not be negative")
	}
	rows, err := db.Query("SELECT * FROM wf_access_contexts WHERE ns_id = ?", nsID)
	if err != nil {
		log.Printf("error fetching access context '%d' : %s\n", nsID, err.Error())
		return nil, err
	}
	defer rows.Close()

	ac := &AccessContext{nsID: nsID}
	ac.perms = make(map[GroupID][]RoleID)

	var acID int64
	var tnsID int64
	var gid GroupID
	var rid RoleID
	for rows.Next() {
		err = rows.Scan(&acID, &tnsID, &gid, &rid)
		if err != nil {
			log.Printf("error fetching details for access context '%d' : %s\n", nsID, err.Error())
			return nil, err
		}
		roles, ok := ac.perms[gid]
		if !ok {
			roles = make([]RoleID, 0, DefACRoleCount)
		}
		roles = append(roles, rid)
		ac.perms[gid] = roles
	}
	err = rows.Err()
	if err != nil {
		log.Printf("error fetching details for access context '%d' : %s\n", nsID, err.Error())
		return nil, err
	}

	return ac, nil
}

// Namespace answers this access context's namespace ID.
func (ac *AccessContext) Namespace() int64 {
	return ac.nsID
}

// AddGroupRole assigns the specified role to the given group, if it
// is not already assigned.
func (ac *AccessContext) AddGroupRole(gid GroupID, rid RoleID) error {
	if gid == 0 || rid == 0 {
		return errors.New("group ID and role ID must be positive integers")
	}

	save := func(gid GroupID, rid RoleID) error {
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		_, err = tx.Exec("INSERT INTO wf_access_contexts(ns_id, group_id, role_id) VALUES(?, ?, ?)", ac.nsID, gid, rid)
		if err != nil {
			return err
		}
		err = tx.Commit()
		if err != nil {
			return err
		}

		return nil
	}

	ac.mutex.Lock()
	defer ac.mutex.Unlock()

	rs, ok := ac.perms[gid]
	if !ok {
		err := save(gid, rid)
		if err != nil {
			return err
		}

		rs = []RoleID{rid}
		ac.perms[gid] = rs
		return nil
	}

	for _, el := range rs {
		if el == rid {
			return nil
		}
	}

	err := save(gid, rid)
	if err != nil {
		return err
	}
	rs = append(rs, rid)
	ac.perms[gid] = rs

	return nil
}

// RemoveGroupRole unassigns the specified role from the given group.
func (ac *AccessContext) RemoveGroupRole(gid GroupID, rid RoleID) error {
	if gid == 0 || rid == 0 {
		return errors.New("group ID and role ID must be positive integers")
	}

	save := func(gid GroupID, rid RoleID) error {
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		_, err = tx.Exec("DELETE FROM wf_access_contexts WHERE ns_id = ? AND group_id = ? AND role_id = ?", ac.nsID, gid, rid)
		if err != nil {
			return err
		}
		err = tx.Commit()
		if err != nil {
			return err
		}

		return nil
	}

	ac.mutex.Lock()
	defer ac.mutex.Unlock()

	rs, ok := ac.perms[gid]
	if !ok {
		return nil
	}

	idx := -1
	for i, el := range rs {
		if el == rid {
			idx = i
			break
		}
	}
	if idx == -1 {
		return nil
	}

	err := save(gid, rid)
	if err != nil {
		return err
	}

	rs = append(rs[:idx], rs[idx+1:]...)
	ac.perms[gid] = rs
	return nil
}

// HasPermission answers `true` if the given group has the requested
// action enabled on the specified document type; `false` otherwise.
func (ac *AccessContext) HasPermission(gid GroupID, dtype DocTypeID, action DocActionID) bool {
	if gid == 0 {
		return false
	}

	ac.mutex.RLock()
	defer ac.mutex.RUnlock()

	rs, ok := ac.perms[gid]
	if !ok {
		return false
	}

	for _, el := range rs {
		if ok, _ := Roles().HasPermission(el, dtype, action); ok {
			return true
		}
	}
	return false
}
