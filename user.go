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
	"database/sql"
	"errors"
	"math"
	"strings"
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
	ID        UserID `json:"ID"`               // Must be globally-unique
	FirstName string `json:"FirstName"`        // For display purposes only
	LastName  string `json:"LastName"`         // For display purposes only
	Email     string `json:"Email"`            // E-mail address of this user
	Active    bool   `json:"Active,omitempty"` // Is this user account active?
}

// Unexported type, only for convenience methods.
type _Users struct{}

// Users provides a resource-like interface to users in the system.
var Users *_Users

// List answers a subset of the users, based on the input
// specification.
//
// Result set begins with ID >= `offset`, and has not more than
// `limit` elements.  A value of `0` for `offset` fetches from the
// beginning, while a value of `0` for `limit` fetches until the end.
func (us *_Users) List(prefix string, offset, limit int64) ([]*User, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.New("offset and limit must be non-negative integers")
	}
	if limit == 0 {
		limit = math.MaxInt64
	}

	var q string
	var rows *sql.Rows
	var err error

	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		q = `
		SELECT id, first_name, last_name, email, active
		FROM wf_users_master
		ORDER BY id
		LIMIT ? OFFSET ?
		`
		rows, err = db.Query(q, limit, offset)
	} else {
		q = `
		SELECT id, first_name, last_name, email, active
		FROM wf_users_master
		WHERE first_name LIKE ?
		UNION
		SELECT id, first_name, last_name, email, active
		FROM wf_users_master
		WHERE last_name LIKE ?
		ORDER BY id
		LIMIT ? OFFSET ?
		`
		rows, err = db.Query(q, prefix+"%", prefix+"%", limit, offset)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ary := make([]*User, 0, 10)
	for rows.Next() {
		var elem User
		err = rows.Scan(&elem.ID, &elem.FirstName, &elem.LastName, &elem.Email, &elem.Active)
		if err != nil {
			return nil, err
		}
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
	row := db.QueryRow("SELECT id, first_name, last_name, email, active FROM wf_users_master WHERE id = ?", uid)
	err := row.Scan(&elem.ID, &elem.FirstName, &elem.LastName, &elem.Email, &elem.Active)
	if err != nil {
		return nil, err
	}

	return &elem, nil
}

// GetByEmail retrieves user information from the database, by looking
// up the given e-mail address.
func (us *_Users) GetByEmail(email string) (*User, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return nil, errors.New("e-mail address should be non-empty")
	}

	var elem User
	row := db.QueryRow("SELECT id, first_name, last_name, email, active FROM wf_users_master WHERE email = ?", email)
	err := row.Scan(&elem.ID, &elem.FirstName, &elem.LastName, &elem.Email, &elem.Active)
	if err != nil {
		return nil, err
	}

	return &elem, nil
}

// IsActive answers `true` if the given user's account is enabled.
func (us *_Users) IsActive(uid UserID) (bool, error) {
	row := db.QueryRow("SELECT active FROM wf_users_master WHERE id = ?", uid)
	var active bool
	err := row.Scan(&active)
	if err != nil {
		return false, err
	}

	return active, nil
}

// GroupsOf answers a list of groups that the given user is a member
// of.
func (us *_Users) GroupsOf(uid UserID) ([]*Group, error) {
	q := `
	SELECT gm.id, gm.name, gm.group_type
	FROM wf_groups_master gm
	JOIN wf_group_users gus ON gus.group_id = gm.id
	JOIN wf_users_master um ON um.id = gus.user_id
	WHERE um.id = ?
	`
	rows, err := db.Query(q, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ary := make([]*Group, 0, 2)
	for rows.Next() {
		var elem Group
		err = rows.Scan(&elem.ID, &elem.Name, &elem.GroupType)
		if err != nil {
			return nil, err
		}
		ary = append(ary, &elem)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return ary, nil
}

// SingletonGroupOf answers the ID of the given user's singleton
// group.
func (us *_Users) SingletonGroupOf(uid UserID) (*Group, error) {
	q := `
	SELECT gm.id, gm.name, gm.group_type
	FROM wf_groups_master gm
	JOIN wf_users_master um ON gm.name = um.email
	WHERE um.id = ?
	`
	var elem Group
	row := db.QueryRow(q, uid)
	err := row.Scan(&elem.ID, &elem.Name, &elem.GroupType)
	if err != nil {
		return nil, err
	}

	return &elem, nil
}
