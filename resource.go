package flow

// Resource represents a collection of documents of a given type.
//
// Resources are the first line of targets for actions and assignment
// of privileges.
type Resource struct {
	id        uint16
	name      string
	endpoint  string
	namespace string // optional
}
