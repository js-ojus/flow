package flow

// Role represents a collection of privileges.
//
// Each user in the system has one or more roles assigned.
type Role struct {
	id    uint16
	name  string
	privs []*Privilege
}
