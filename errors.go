package flow

// Error defines `flow`-specific errors, and satisfies the `error`
// interface.
type Error string

// Error implements the `error` interface.
func (e Error) Error() string {
	return string(e)
}

//

const (
	// ErrUnknown : unknown internal error
	ErrUnknown = Error("ErrUnknown : unknown internal error")

	// ErrDocEventRedundant : another equivalent event has already effected this action
	ErrDocEventRedundant = Error("ErrDocEventRedundant : another equivalent event has already applied this action")
	// ErrDocEventDocTypeMismatch : document's type does not match event's type
	ErrDocEventDocTypeMismatch = Error("ErrDocEventDocTypeMismatch : document's type does not match event's type")
	// ErrDocEventStateMismatch : document's state does not match event's state
	ErrDocEventStateMismatch = Error("ErrDocEventStateMismatch : document's state does not match event's state")
	// ErrDocEventAlreadyApplied : event already applied; nothing to do
	ErrDocEventAlreadyApplied = Error("ErrDocEventAlreadyApplied : event already applied; nothing to do")

	// ErrWorkflowInactive : this workflow is currently inactive
	ErrWorkflowInactive = Error("ErrWorkflowInactive : this workflow is currently inactive")
	// ErrWorkflowInvalidAction : given action cannot be performed on this document's current state
	ErrWorkflowInvalidAction = Error("ErrWorkflowInvalidAction : given action cannot be performed on this document's current state")

	// ErrMessageNoRecipients : list of recipients is empty
	ErrMessageNoRecipients = Error("ErrMessageNoRecipients : list of recipients is empty")
)
