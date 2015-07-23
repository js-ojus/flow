package flow

// DocState is one of a set of enumerated states for a document, as
// defined by the consuming application.
//
// `flow`, therefore, does not assume anything about the specifics of
// any state.
type DocState struct {
	dtype      DocType // for namespace purposes
	name       string
	successors []*DocState // possible next states
}
