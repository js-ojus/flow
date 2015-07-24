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

import "fmt"

// PrivilegeType enumerates the possible operations on resources and
// documents, closely modeling REST conventions.
type PrivilegeType byte

const (
	PrivList PrivilegeType = iota + 1
	PrivCreate
	PrivRead
	PrivUpdate
	PrivDelete
	PrivUndelete
	PrivArchive
	PrivRestore
)

// Privilege represents an authorisation to perform a specific action
// on a specified set of documents.
//
// Privileges can be held by individual users, roles and groups.
type Privilege struct {
	resource *Resource
	doc      *Document // only if not on a resource
	privs    []PrivilegeType
}

// NewPrivilege creates and initialises a set of permissions on a
// target, identified as a resource and - optionally - a document.
func NewPrivilege(res *Resource, doc *Document) (*Privilege, error) {
	if res == nil {
		return nil, fmt.Errorf("resource not specified")
	}

	p := &Privilege{resource: res, doc: doc}
	p.privs = make([]PrivilegeType, 4)
	return p, nil
}

// Resource answers the resource part of this privilege's target.
func (p *Privilege) Resource() *Resource {
	return p.resource
}

// Document answers the document part of this privilege's target.
func (p *Privilege) Document() *Document {
	return p.doc
}

// AddPrivilegeType includes the given permission in this privilege.
func (p *Privilege) AddPrivilegeType(pt PrivilegeType) bool {
	for _, el := range p.privs {
		if el == pt {
			return false
		}
	}

	p.privs = append(p.privs, pt)
	return true
}

// PrivilegeTypes answers a copy of this privilege's permissions.
func (p *Privilege) PrivilegeTypes() []PrivilegeType {
	pts := make([]PrivilegeType, len(p.privs))
	copy(pts, p.privs)
	return pts
}

// IsOnSameTargetAs answers if this privilege operates on the same
// resource/document as the given one.
func (p *Privilege) IsOnSameTargetAs(p2 *Privilege) bool {
	return p.IsOnTarget(p2.resource, p2.doc)
}

// IsOnTarget answers if this privilege operates on the given
// resource/document as the given ones.
func (p *Privilege) IsOnTarget(res *Resource, doc *Document) bool {
	if p.resource.id != res.id {
		return false
	}
	if (p.doc != nil && doc == nil) ||
		(p.doc == nil && doc != nil) {
		return false
	}
	if p.doc != nil {
		if p.doc.id != doc.id {
			return false
		}
	}

	return true
}
