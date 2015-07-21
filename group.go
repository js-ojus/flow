package flow

// Group represents a specified collection of users.
//
// A user belongs to zero or more groups.  Groups can have associated
// privileges, too.
type Group struct {
	id    uint16
	name  string
	privs []*Privilege
}
