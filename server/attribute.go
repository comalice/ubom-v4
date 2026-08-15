package ubom

import "fmt"

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
