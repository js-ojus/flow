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
