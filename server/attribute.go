package ubom

import (
	"fmt"
	"math/big"
)

type AttributeDefID string

type AttributeValueType string

const (
	AttributeValueString  AttributeValueType = "string"
	AttributeValueDecimal AttributeValueType = "decimal"
	AttributeValueEnum    AttributeValueType = "enum"
)

// AttributeDef describes a value that taxonomy nodes may require or expose.
type AttributeDef struct {
	ID         AttributeDefID
	Label      string
	ValueType  AttributeValueType
	Unit       string
	EnumValues []string
}

type AttributeAssignment struct {
	AttributeDefID AttributeDefID
	Required       bool
}

// PartNumberAttribute is a canonical value owned by a part number. The
// definition supplies its type, unit, and any enum vocabulary.
type PartNumberAttribute struct {
	AttributeDefID AttributeDefID
	Value          string
}

func (d AttributeDef) Validate() error {
	if d.ID == "" {
		return fmt.Errorf("attribute definition has no ID")
	}
	if d.Label == "" {
		return fmt.Errorf("attribute definition %q has no label", d.ID)
	}
	switch d.ValueType {
	case AttributeValueString, AttributeValueDecimal:
		if len(d.EnumValues) != 0 {
			return fmt.Errorf("attribute definition %q has enum values but type is %q", d.ID, d.ValueType)
		}
	case AttributeValueEnum:
		if len(d.EnumValues) == 0 {
			return fmt.Errorf("attribute definition %q has no enum values", d.ID)
		}
	default:
		return fmt.Errorf("attribute definition %q has unsupported value type %q", d.ID, d.ValueType)
	}
	return nil
}

// ValidatePartNumberAttributes checks values against the effective taxonomy
// assignments for a part number. It is pure domain validation: persistence
// and API concerns stay outside this function.
func ValidatePartNumberAttributes(values []PartNumberAttribute, assignments []AttributeAssignment, definitions []AttributeDef) error {
	defs := make(map[AttributeDefID]AttributeDef, len(definitions))
	for _, definition := range definitions {
		defs[definition.ID] = definition
	}
	assigned := make(map[AttributeDefID]AttributeAssignment, len(assignments))
	for _, assignment := range assignments {
		assigned[assignment.AttributeDefID] = assignment
	}
	seen := make(map[AttributeDefID]bool, len(values))
	for _, value := range values {
		if value.AttributeDefID == "" {
			return fmt.Errorf("part number attribute has no definition ID")
		}
		if value.Value == "" {
			return fmt.Errorf("part number attribute %q has no value", value.AttributeDefID)
		}
		if seen[value.AttributeDefID] {
			return fmt.Errorf("duplicate part number attribute %q", value.AttributeDefID)
		}
		seen[value.AttributeDefID] = true
		definition, ok := defs[value.AttributeDefID]
		if !ok {
			return fmt.Errorf("part number attribute %q references unknown definition", value.AttributeDefID)
		}
		if _, ok := assigned[value.AttributeDefID]; !ok {
			return fmt.Errorf("part number attribute %q is not assigned by taxonomy", value.AttributeDefID)
		}
		switch definition.ValueType {
		case AttributeValueString:
		case AttributeValueDecimal:
			if _, ok := new(big.Rat).SetString(value.Value); !ok {
				return fmt.Errorf("part number attribute %q is not a decimal", value.AttributeDefID)
			}
		case AttributeValueEnum:
			valid := false
			for _, option := range definition.EnumValues {
				if option == value.Value {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("part number attribute %q has invalid enum value %q", value.AttributeDefID, value.Value)
			}
		}
	}
	for _, assignment := range assignments {
		if assignment.Required && !seen[assignment.AttributeDefID] {
			return fmt.Errorf("required part number attribute %q is missing", assignment.AttributeDefID)
		}
	}
	return nil
}
