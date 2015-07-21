package flow

import "fmt"

// Role represents a collection of privileges.
//
// Each user in the system has one or more roles assigned.
type Role struct {
	id    uint16
	name  string
	privs []*Privilege
}

// NewRole creates and initialises a role.
//
// Usually, all available roles should be loaded during system
// initialization.  Only roles created during runtime should be added
// dynamically.
func NewRole(id uint16, name string) (*Role, error) {
	if id == 0 || name == "" {
		return nil, fmt.Errorf("invalid role data -- id: %d, name: %s", id, name)
	}

	return &Role{id: id, name: name}, nil
}

// AddPrivilege includes the given privilege in the set of privileges
// assigned to this role.
func (r *Role) AddPrivilege(p *Privilege) bool {
	for _, el := range r.privs {
		if el.IsOnSameTargetAs(p) {
			return false
		}
	}

	r.privs = append(r.privs, p)
	return true
}

// RemovePrivilegesOn removes the privileges that this role has on the
// given target.
func (r *Role) RemovePrivilegesOn(res *Resource, doc *Document) bool {
	found := false
	idx := -1
	for i, el := range r.privs {
		if el.IsOnTarget(res, doc) {
			found = true
			idx = i
			break
		}
	}
	if !found {
		return false
	}

	r.privs = append(r.privs[:idx], r.privs[idx+1:]...)
	return true
}

// ReplacePrivilege any current privilege on the given target, with
// the given privilege.
func (r *Role) ReplacePrivilege(p *Privilege) bool {
	if !r.RemovePrivilegesOn(p.resource, p.doc) {
		return false
	}

	r.privs = append(r.privs, p)
	return true
}
