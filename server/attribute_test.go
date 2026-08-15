package ubom

import "testing"

func TestValidatePartNumberAttributesResistor(t *testing.T) {
	definitions := []AttributeDef{
		{ID: "footprint", Label: "Footprint", ValueType: AttributeValueEnum, EnumValues: []string{"0603", "0805"}},
		{ID: "resistance", Label: "Resistance", ValueType: AttributeValueDecimal, Unit: "ohm"},
		{ID: "tolerance", Label: "Tolerance", ValueType: AttributeValueDecimal, Unit: "percent"},
		{ID: "power_rating", Label: "Power rating", ValueType: AttributeValueDecimal, Unit: "W"},
	}
	assignments := []AttributeAssignment{
		{AttributeDefID: "footprint", Required: true},
		{AttributeDefID: "resistance", Required: true},
		{AttributeDefID: "tolerance"},
		{AttributeDefID: "power_rating"},
	}
	values := []PartNumberAttribute{
		{AttributeDefID: "footprint", Value: "0603"},
		{AttributeDefID: "resistance", Value: "10000"},
		{AttributeDefID: "tolerance", Value: "1"},
		{AttributeDefID: "power_rating", Value: "0.1"},
	}

	if err := ValidatePartNumberAttributes(values, assignments, definitions); err != nil {
		t.Fatalf("ValidatePartNumberAttributes() error = %v", err)
	}
}

func TestValidatePartNumberAttributesRejectsInvalidValues(t *testing.T) {
	definitions := []AttributeDef{
		{ID: "footprint", Label: "Footprint", ValueType: AttributeValueEnum, EnumValues: []string{"0603"}},
		{ID: "resistance", Label: "Resistance", ValueType: AttributeValueDecimal},
	}
	assignments := []AttributeAssignment{
		{AttributeDefID: "footprint", Required: true},
		{AttributeDefID: "resistance", Required: true},
	}
	tests := []struct {
		name   string
		values []PartNumberAttribute
	}{
		{name: "missing required", values: []PartNumberAttribute{{AttributeDefID: "footprint", Value: "0603"}}},
		{name: "invalid enum", values: []PartNumberAttribute{{AttributeDefID: "footprint", Value: "1210"}, {AttributeDefID: "resistance", Value: "10000"}}},
		{name: "invalid decimal", values: []PartNumberAttribute{{AttributeDefID: "footprint", Value: "0603"}, {AttributeDefID: "resistance", Value: "ten-k"}}},
		{name: "duplicate", values: []PartNumberAttribute{{AttributeDefID: "footprint", Value: "0603"}, {AttributeDefID: "footprint", Value: "0603"}, {AttributeDefID: "resistance", Value: "10000"}}},
		{name: "unknown", values: []PartNumberAttribute{{AttributeDefID: "footprint", Value: "0603"}, {AttributeDefID: "resistance", Value: "10000"}, {AttributeDefID: "package", Value: "0603"}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := ValidatePartNumberAttributes(test.values, assignments, definitions); err == nil {
				t.Fatal("ValidatePartNumberAttributes() accepted invalid values")
			}
		})
	}
}
