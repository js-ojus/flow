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
	"math"
)

// UserID is the type of unique user identifiers.
type UserID int64

// User represents any kind of a user invoking or otherwise
// participating in a defined workflow in the system.
//
// User details are expected to be provided by an external identity
// provider application or directory.  `flow` neither defines nor
// manages users.
type User struct {
	id    UserID // Must be globally-unique
	name  string // For display purposes only
	email string // E-mail address of this user
}

// ID answers this user's ID.
func (u *User) ID() UserID {
	return u.id
}

// Name answers this user's name.
func (u *User) Name() string {
	return u.name
}

// Email answers this user's e-mail address.
func (u *User) Email() string {
	return u.email
}

// Unexported type, only for convenience methods.
type _Users struct{}

var _users *_Users

func init() {
	_users = &_Users{}
}

// Users provides a resource-like interface to users in the system.
func Users() *_Users {
	return _users
}

// List answers a subset of the users, based on the input
// specification.
//
// Result set begins with ID >= `offset`, and has not more than
// `limit` elements.  A value of `0` for `offset` fetches from the
// beginning, while a value of `0` for `limit` fetches until the end.
func (us *_Users) List(offset, limit int64) ([]*User, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.New("offset and limit must be non-negative integers")
	}
	if limit == 0 {
		limit = math.MaxInt64
	}

	q := `
	SELECT id, fname, lname, email
	FROM wf_users_master
	ORDER BY id
	LIMIT ? OFFSET ?
	`
	rows, err := db.Query(q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fname, lname string
	ary := make([]*User, 0, 10)
	for rows.Next() {
		var elem User
		err = rows.Scan(&elem.id, &fname, &lname, &elem.email)
		if err != nil {
			return nil, err
		}
		elem.name = fname + " " + lname
		ary = append(ary, &elem)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ary, nil
}

// Get instantiates a user instance by reading the database.
func (us *_Users) Get(uid UserID) (*User, error) {
	if uid <= 0 {
		return nil, errors.New("user ID should be a positive integer")
	}

	var elem User
	var fname, lname string
	row := db.QueryRow("SELECT id, first_name, last_name, email FROM wf_users_master_v WHERE id = ?", uid)
	err := row.Scan(&elem.id, &fname, &lname, &elem.email)
	if err != nil {
		return nil, err
	}

	elem.name = fname + " " + lname
	return &elem, nil
}

// IsActive answers `true` if the given user's account is enabled.
func (us *_Users) IsActive(uid UserID) (bool, error) {
	row := db.QueryRow("SELECT status FROM wf_users_master_v WHERE id = ?", uid)
	var status bool
	err := row.Scan(&status)
	if err != nil {
		return false, err
	}

	return status, nil
}

// GroupsOf answers a list of groups that the given user is a member
// of.
func (us *_Users) GroupsOf(uid UserID) ([]GroupID, error) {
	rows, err := db.Query("SELECT group_id FROM wf_group_users WHERE user_id = ?", uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	gids := make([]GroupID, 0, 1)
	var gid GroupID
	for rows.Next() {
		err = rows.Scan(&gid)
		if err != nil {
			return nil, err
		}
		gids = append(gids, gid)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return gids, nil
}

// SingletonGroupOf answers the ID of the given user's singleton
// group.
func (us *_Users) SingletonGroupOf(uid UserID) (GroupID, error) {
	q := `
	SELECT gm.id FROM wf_groups_master gm
	JOIN wf_users_master um ON gm.name = um.email
	WHERE um.id = ?
	`
	var gid GroupID
	row := db.QueryRow(q, uid)
	err := row.Scan(&gid)
	if err != nil {
		return 0, err
	}

	return gid, nil
}
