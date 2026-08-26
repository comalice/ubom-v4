package ubom

import "errors"

// ErrorCategory is the stable classification of a domain error. Error
// messages are descriptive and are not part of the domain contract.
type ErrorCategory string

const (
	ErrorCategoryValidation     ErrorCategory = "validation"
	ErrorCategoryDefinition     ErrorCategory = "definition"
	ErrorCategorySyntax         ErrorCategory = "syntax"
	ErrorCategoryValue          ErrorCategory = "value"
	ErrorCategoryClassification ErrorCategory = "classification"
	ErrorCategoryOrdering       ErrorCategory = "ordering"
	ErrorCategoryAllocation     ErrorCategory = "allocation"
	ErrorCategoryConflict       ErrorCategory = "conflict"
)

func (c ErrorCategory) Error() string { return string(c) }

// These values are errors.Is targets. They are categories, rather than
// message-bearing errors, so their identity remains stable as messages evolve.
var (
	ErrValidation        = ErrorCategoryValidation
	ErrInvalidDefinition = ErrorCategoryDefinition
	ErrInvalidSyntax     = ErrorCategorySyntax
	ErrInvalidValue      = ErrorCategoryValue
	ErrUnclassified      = ErrorCategoryClassification
	ErrNoOrdering        = ErrorCategoryOrdering
	ErrAllocation        = ErrorCategoryAllocation
	ErrConflict          = ErrorCategoryConflict
)

// DomainError reports a domain failure with a stable category and an
// optional underlying cause. Use errors.Is to inspect Category; do not match
// Error() text.
type DomainError struct {
	Category ErrorCategory
	Message  string
	Cause    error
}

func (e *DomainError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Message == "" {
		return e.Category.Error()
	}
	return e.Message
}

func (e *DomainError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func (e *DomainError) Is(target error) bool {
	if e == nil {
		return target == nil
	}
	switch t := target.(type) {
	case ErrorCategory:
		if t == ErrorCategoryValidation {
			return e.Category == ErrorCategoryDefinition || e.Category == ErrorCategorySyntax || e.Category == ErrorCategoryValue
		}
		return e.Category == t
	case *DomainError:
		return t != nil && e.Category == t.Category
	default:
		return false
	}
}

// NewDomainError constructs a classified domain error. Cause is optional and
// can be supplied with WrapDomainError when a lower-level cause is useful.
func NewDomainError(category ErrorCategory, message string) *DomainError {
	return &DomainError{Category: category, Message: message}
}

// WrapDomainError adds a stable domain category while preserving cause
// inspection through errors.Is and errors.As.
func WrapDomainError(category ErrorCategory, message string, cause error) *DomainError {
	return &DomainError{Category: category, Message: message, Cause: cause}
}

// IsCategory reports whether err, including any wrapped error, has category.
func IsCategory(err error, category ErrorCategory) bool {
	return errors.Is(err, category)
}
