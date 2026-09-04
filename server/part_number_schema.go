package ubom

import (
	"fmt"
	"strings"
	"unicode"
)

// PartNumberSchemaID identifies an immutable, named interpretation schema.
type PartNumberSchemaID string

// PartNumberSchema is the durable assignment of a taxonomy revision and
// classification policy. The sequence definition is derived from TaxonomyDef.
type PartNumberSchema struct {
	ID          PartNumberSchemaID `json:"id"`
	Name        string             `json:"name"`
	TaxonomyDef TaxonomyDefID      `json:"taxonomyDef"`
	Policy      SchemaPolicy       `json:"policy"`
}

func (s PartNumberSchema) Validate() error {
	if string(s.ID) == "" || strings.IndexFunc(string(s.ID), unicode.IsSpace) >= 0 {
		return NewDomainError(ErrorCategoryDefinition, fmt.Sprintf("part-number schema has invalid ID %q", s.ID))
	}
	if strings.TrimSpace(s.Name) == "" {
		return NewDomainError(ErrorCategoryDefinition, "part-number schema name is required")
	}
	if string(s.TaxonomyDef) == "" || strings.IndexFunc(string(s.TaxonomyDef), unicode.IsSpace) >= 0 {
		return NewDomainError(ErrorCategoryDefinition, "part-number schema taxonomy definition is required")
	}
	return nil
}
