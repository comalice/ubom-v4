package ubom

import (
	"errors"
	"testing"
)

func TestDomainErrorCategorySurvivesWrapping(t *testing.T) {
	leaf := errors.New("parser detail")
	err := WrapDomainError(ErrorCategoryValue, "invalid captured field", leaf)

	if !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("errors.Is(error, ErrInvalidValue) = false")
	}
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("errors.Is(error, ErrValidation) = false")
	}
	if IsCategory(err, ErrorCategoryDefinition) {
		t.Fatalf("value error classified as definition error")
	}
	if !errors.Is(err, leaf) {
		t.Fatalf("errors.Is(error, cause) = false")
	}
	var classified *DomainError
	if !errors.As(err, &classified) || classified.Category != ErrorCategoryValue {
		t.Fatalf("errors.As() did not recover the domain error: %#v", classified)
	}
}

func TestFieldNameValidation(t *testing.T) {
	valid, err := NewFieldName("category-code")
	if err != nil || valid != "category-code" {
		t.Fatalf("NewFieldName() = %q, %v", valid, err)
	}

	for _, input := range []string{"", " category", "category ", "category code"} {
		if _, err := NewFieldName(input); err == nil || !errors.Is(err, ErrInvalidValue) {
			t.Errorf("NewFieldName(%q) error = %v, want ErrInvalidValue", input, err)
		}
	}
}

func TestFieldValueValidation(t *testing.T) {
	empty, err := NewFieldValue("")
	if err != nil || empty.String() != "" {
		t.Fatalf("empty FieldValue = %q, %v; empty values should be allowed", empty, err)
	}
	if _, err := NewFieldValue(string([]byte{0xff})); err == nil || !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("invalid UTF-8 value error = %v, want ErrInvalidValue", err)
	}
}
