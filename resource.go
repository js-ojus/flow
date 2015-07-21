package flow

import "fmt"

// Resource represents a collection of documents of a given type.
//
// Resources are the first line of targets for actions and assignment
// of privileges.
type Resource struct {
	id        uint16 // globally-unique identifier
	name      string // convenient name
	endpoint  string // globally-unique end point path
	namespace string // optional
}

// NewResource instantiates a resource.
func NewResource(id uint16, name, ep, ns string) (*Resource, error) {
	if id == 0 || name == "" || ep == "" {
		return nil, fmt.Errorf("invalid resource information")
	}

	return &Resource{id, name, ep, ns}, nil
}

// ID answers this resource's globally-unique ID.
func (r *Resource) ID() uint16 {
	return r.id
}

// Name answers this resource's name.
func (r *Resource) Name() string {
	return r.name
}

// EndPoint answers this resource's globally-unique endpoint.
func (r *Resource) EndPoint() string {
	return r.endpoint
}

// Namespace answers this resource's namespace, if any.
func (r *Resource) Namespace() string {
	return r.namespace
}
