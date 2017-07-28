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

// GetUser instantiates a user instance by reading the database.
func GetUser(uid UserID) (*User, error) {
	if uid <= 0 {
		return nil, errors.New("user ID should be a positive integer")
	}

	var tid int64
	var fname string
	var lname string
	var email string
	row := db.QueryRow("SELECT id, first_name, last_name, email FROM wf_users_master_v WHERE id = ?", uid)
	err := row.Scan(&tid, &fname, &lname, &email)
	if err != nil {
		return nil, err
	}
	u := &User{id: uid, name: fname + " " + lname, email: email}
	return u, nil
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

// IsActive answers `true` if this user's account is enabled.
func (u *User) IsActive() (bool, error) {
	row := db.QueryRow("SELECT status FROM wf_users_master_v WHERE id = ?", u.id)
	var status bool
	err := row.Scan(&status)
	if err != nil {
		return false, err
	}

	return status, nil
}

// IsUserActive answers `true` if the given user's account is enabled.
func IsUserActive(uid UserID) (bool, error) {
	row := db.QueryRow("SELECT status FROM wf_users_master_v WHERE id = ?", uid)
	var status bool
	err := row.Scan(&status)
	if err != nil {
		return false, err
	}

	return status, nil
}

// Groups answers a list of groups that this user is a member of.
func (u *User) Groups() ([]GroupID, error) {
	active, err := u.IsActive()
	if !active {
		return nil, errors.New("this user is currently not active")
	}

	rows, err := db.Query("SELECT group_id FROM wf_group_users WHERE user_id = ?", u.id)
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

// SingletonGroup answers the ID of this user's singleton group.
func (u *User) SingletonGroup() (GroupID, error) {
	active, err := u.IsActive()
	if err != nil {
		return 0, err
	}
	if !active {
		return 0, errors.New("this user is currently not active")
	}

	row := db.QueryRow("SELECT id from wf_groups_master WHERE name = ?", u.email)
	var gid GroupID
	err = row.Scan(&gid)
	if err != nil {
		return 0, err
	}

	return gid, nil
}
