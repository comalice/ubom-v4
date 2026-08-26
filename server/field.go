package ubom

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// FieldName identifies a named capture shared by SeqDef and taxonomy
// predicates. Names are non-empty and may contain punctuation, but may not
// contain whitespace.
type FieldName string

func NewFieldName(value string) (FieldName, error) {
	name := FieldName(value)
	if err := name.Validate(); err != nil {
		return "", err
	}
	return name, nil
}

func (n FieldName) String() string { return string(n) }

func (n FieldName) Validate() error {
	if n == "" {
		return WrapDomainError(ErrorCategoryValue, "field name is empty", nil)
	}
	if strings.TrimSpace(string(n)) != string(n) {
		return WrapDomainError(ErrorCategoryValue, fmt.Sprintf("field name %q has surrounding whitespace", n), nil)
	}
	for _, r := range string(n) {
		if unicode.IsSpace(r) {
			return WrapDomainError(ErrorCategoryValue, fmt.Sprintf("field name %q contains whitespace", n), nil)
		}
	}
	return nil
}

// FieldValue is the textual value captured from a sequence. Empty values are
// allowed: whether an empty capture is structurally valid belongs to SeqDef.
type FieldValue string

func NewFieldValue(value string) (FieldValue, error) {
	fieldValue := FieldValue(value)
	if err := fieldValue.Validate(); err != nil {
		return "", err
	}
	return fieldValue, nil
}

func (v FieldValue) String() string { return string(v) }

func (v FieldValue) Validate() error {
	if !utf8.ValidString(string(v)) {
		return WrapDomainError(ErrorCategoryValue, "field value is not valid UTF-8", nil)
	}
	return nil
}
