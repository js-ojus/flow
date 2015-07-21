package flow

import (
	"fmt"
)

// User represents any kind of a user invoking or other participating
// in a defined workflow in the system.
//
// User details are expected to be provided by an external identity
// provider application or directory.  `flow` neither defines nor
// manages users.
type User struct {
	id     uint64   // must be globally-unique
	name   string   // for display purposes only
	active bool     // status of the user account
	roles  []*Role  // all roles assigned to this user
	groups []*Group // all groups this user is a part of
}

// NewUser instantiates a user instance in the system.
//
// In most cases, this should be done upon the corresponding user's
// first action in the system.
func NewUser(id uint64, name string) (*User, error) {
	if id == 0 || name == "" {
		return nil, fmt.Errorf("invalid user data -- id: %d, name: %s", id, name)
	}

	return &User{id: id, name: name}, nil
}
