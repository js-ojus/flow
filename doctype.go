package flow

// DocType enumerates the types of documents in the system, as defined
// by the consuming application.
//
// Accordingly, `flow` does not assume anything about the specifics of
// the any document type.  Instead, it treats document types as plain,
// but controlled, vocabulary.
//
// All document types must be defined as constant strings.
type DocType string
