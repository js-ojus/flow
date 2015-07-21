package flow

// PrivilegeType enumerates the possible operations on resources and
// documents, closely modeling REST conventions.
type PrivilegeType string

const (
	PrivList     PrivilegeType = "LIST"
	PrivCreate                 = "CREATE"
	PrivRead                   = "READ"
	PrivUpdate                 = "UPDATE"
	PrivDelete                 = "DELETE"
	PrivUndelete               = "UNDELETE"
	PrivArchive                = "ARCHIVE"
	PrivRestore                = "RESTORE"
)

// Privilege represents an authorisation to perform a specific action
// on a specified set of documents.
//
// Privileges can be held by individual users, roles and groups.
type Privilege struct {
	privs    []PrivilegeType
	resource *Resource
	doc      *Document // only if not on a resource
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
