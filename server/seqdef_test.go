package ubom

import (
	"errors"
	"testing"
)

const (
	digitsAlphabet  = "0123456789"
	lettersAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	alnumAlphabet   = digitsAlphabet + lettersAlphabet
)

func TestSeqDefExamples(t *testing.T) {
	tests := []struct {
		name string
		def  SeqDef
		good []string
		bad  []string
	}{
		{
			name: "McMaster style",
			def: NewSeqDef(Choice(
				Concat(Places(5, digitsAlphabet), PlaceSequence(lettersAlphabet), Places(3, digitsAlphabet)),
				Concat(Places(4, digitsAlphabet), Literal("K"), Places(3, digitsAlphabet)),
			)),
			good: []string{"91290A115", "92196A031", "95462A029", "4936K451"},
			bad:  []string{"91290A11", "91290a115", "4936A451"},
		},
		{
			name: "Digi-Key style",
			def: NewSeqDef(Concat(
				Places(3, digitsAlphabet), Literal("-"), Places(4, digitsAlphabet),
				Choice(Literal("-1"), Literal("-2"), Literal("-5"), Literal("-6"), Literal("CT"), Literal("TR"), Literal("DKR")),
				Literal("-ND"),
			)),
			good: []string{"255-1504-1-ND", "255-1504-2-ND", "255-1504-6-ND"},
			bad:  []string{"255-1504-ND", "255-1504-3-ND", "255-1504-1"},
		},
		{
			name: "Mouser style",
			def:  NewSeqDef(Concat(Literal("595-"), PlaceSequence(alnumAlphabet, alnumAlphabet, alnumAlphabet, alnumAlphabet, alnumAlphabet, alnumAlphabet, alnumAlphabet, alnumAlphabet, alnumAlphabet, alnumAlphabet))),
			good: []string{"595-CSD16301Q2"},
			bad:  []string{"595-CSD16301Q", "595-CSD16301Q2-"},
		},
		{
			name: "Murata style",
			def: NewSeqDef(Concat(
				Literal("GRM"),
				PlaceSequence(digitsAlphabet, digitsAlphabet, digitsAlphabet),
				PlaceSequence(lettersAlphabet, digitsAlphabet),
				PlaceSequence(digitsAlphabet, lettersAlphabet),
				PlaceSequence(digitsAlphabet, digitsAlphabet, digitsAlphabet),
				PlaceSequence(lettersAlphabet),
				PlaceSequence(lettersAlphabet, digitsAlphabet, digitsAlphabet),
				PlaceSequence(lettersAlphabet),
			)),
			good: []string{"GRM188R61A226ME15D"},
			bad:  []string{"GRM188R61A226ME15", "GCM188R61A226ME15D"},
		},
		{
			name: "Molex style",
			def: NewSeqDef(Choice(
				Concat(Places(5, digitsAlphabet), Literal("-"), Places(4, digitsAlphabet)),
				Places(10, digitsAlphabet),
			)),
			good: []string{"43025-0400", "0533091070"},
			bad:  []string{"43025-400", "053309107A"},
		},
		{
			name: "Samtec style",
			def: NewSeqDef(Concat(
				Choice(Literal("SFMC"), Literal("IDSS")), Literal("-"),
				Places(3, digitsAlphabet), Literal("-"),
				Places(2, digitsAlphabet), Literal("-"),
				PlaceSequence(lettersAlphabet), Literal("-"),
				PlaceSequence(lettersAlphabet),
			)),
			good: []string{"SFMC-109-03-S-D", "IDSS-008-05-G-D"},
			bad:  []string{"SFMC-109-3-S-D", "QTH-109-03-S-D"},
		},
		{
			name: "internal category schema",
			def: NewSeqDef(Concat(
				Choice(Literal("RES"), Literal("CAP"), Literal("CON"), Literal("DIO")),
				Literal("-"), Places(4, digitsAlphabet), Literal("-"), Places(4, alnumAlphabet),
			)),
			good: []string{"RES-0042-10K0", "CAP-1234-X7R0", "CON-0001-USB1"},
			bad:  []string{"RES-42-10K0", "BAT-0001-AAAA", "RES-0042-10k0"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, input := range test.good {
				if got, err := test.def.Parse(input); err != nil || got != input {
					t.Errorf("Parse(%q) = %q, %v; want the input and no error", input, got, err)
				}
			}
			for _, input := range test.bad {
				if _, err := test.def.Parse(input); err == nil {
					t.Errorf("Parse(%q) accepted invalid input", input)
				}
			}
		})
	}
}

func TestSeqDefRejectsMissingRoot(t *testing.T) {
	if _, err := (SeqDef{}).Parse("anything"); err == nil {
		t.Fatal("Parse() accepted a definition without a root")
	}
}

func TestSeqDefValidate(t *testing.T) {
	tests := []struct {
		name string
		def  SeqDef
		want bool
	}{
		{
			name: "valid",
			def:  NewSeqDef(Bind("number", Range(0, 9))).WithID("pn-v1"),
			want: true,
		},
		{
			name: "missing ID",
			def:  NewSeqDef(Literal("PN-1")),
		},
		{
			name: "missing root",
			def:  SeqDef{ID: "pn-v1"},
		},
		{
			name: "bind missing name",
			def:  NewSeqDef(Bind("", Literal("PN-1"))).WithID("pn-v1"),
		},
		{
			name: "bind missing child",
			def:  NewSeqDef(Bind("value", nil)).WithID("pn-v1"),
		},
		{
			name: "place empty alphabet",
			def:  NewSeqDef(PlaceSequence("ABC", "")).WithID("pn-v1"),
		},
		{
			name: "invalid range",
			def:  NewSeqDef(RangeNode{Min: 10, Max: 1, WidthValue: 1}).WithID("pn-v1"),
		},
		{
			name: "range maximum exceeds width",
			def:  NewSeqDef(Range(0, 100).Width(2)).WithID("pn-v1"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.def.Validate()
			if test.want && err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
			if !test.want && err == nil {
				t.Fatal("Validate() accepted invalid definition")
			}
		})
	}
}

func TestSeqDefVersionMetadataDefaultsAndValidation(t *testing.T) {
	legacy := NewSeqDef(Literal("A")).WithID("legacy")
	if legacy.DefinitionID != "legacy" || legacy.Version != 1 || legacy.ParentID != "" {
		t.Fatalf("legacy metadata = %#v, want identity legacy, version 1, no parent", legacy)
	}
	version := legacy.WithVersionMetadata("stable", 2, legacy.ID)
	if err := version.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := legacy.WithVersionMetadata("stable", 0, "").Validate(); err == nil {
		t.Fatal("accepted zero version")
	}
}

func TestSeqDefValidateRecognizesAllNodeKinds(t *testing.T) {
	tests := []struct {
		name string
		node SeqDefNode
	}{
		{name: "literal", node: LiteralNode{Text: "PN"}},
		{name: "concat", node: ConcatNode{Children: []SeqDefNode{Literal("PN")}}},
		{name: "choice", node: ChoiceNode{Children: []SeqDefNode{Literal("PN")}}},
		{name: "bind", node: BindNode{Name: "value", Child: Literal("PN")}},
		{name: "places", node: PlaceSequenceNode{Alphabets: []string{"ABC"}}},
		{name: "range", node: RangeNode{Min: 0, Max: 9, WidthValue: 1}},
		{name: "range radix", node: RangeRadixNode{Children: []SeqDefNode{Range(0, 9)}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := NewSeqDef(test.node).WithID("seq-v1").Validate(); err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
		})
	}
}

func TestSeqDefValidateClassifiesStructuralErrors(t *testing.T) {
	tests := []SeqDef{
		{},
		{ID: "seq-v1"},
		NewSeqDef(Bind("", Literal("PN"))).WithID("seq-v1"),
		NewSeqDef(Bind("value", nil)).WithID("seq-v1"),
		NewSeqDef(PlaceSequence()).WithID("seq-v1"),
		NewSeqDef(Concat()).WithID("seq-v1"),
		NewSeqDef(Choice()).WithID("seq-v1"),
		NewSeqDef(RangeRadix()).WithID("seq-v1"),
		NewSeqDef(RangeNode{Min: 10, Max: 1, WidthValue: 1}).WithID("seq-v1"),
	}

	for i, definition := range tests {
		err := definition.Validate()
		if err == nil {
			t.Errorf("case %d: Validate() accepted invalid definition", i)
			continue
		}
		if !errors.Is(err, ErrInvalidDefinition) {
			t.Errorf("case %d: errors.Is(Validate(), ErrInvalidDefinition) = false; error = %v", i, err)
		}
	}
}

type unknownSeqDefNode struct{}

func (unknownSeqDefNode) parseAt(string, int, map[string]string) (string, int, error) {
	return "", 0, nil
}

func TestSeqDefValidateClassifiesUnknownNode(t *testing.T) {
	err := NewSeqDef(unknownSeqDefNode{}).WithID("seq-v1").Validate()
	if err == nil {
		t.Fatal("Validate() accepted an unknown sequence node")
	}
	if !errors.Is(err, ErrInvalidDefinition) {
		t.Fatalf("errors.Is(Validate(), ErrInvalidDefinition) = false; error = %v", err)
	}
}

func TestSeqDefRange(t *testing.T) {
	definition := NewSeqDef(Concat(
		Literal("PN-"),
		Range(0, 17).Width(2, '0'),
	))

	for _, input := range []string{"PN-00", "PN-09", "PN-17"} {
		if _, err := definition.Parse(input); err != nil {
			t.Errorf("Parse(%q) error = %v", input, err)
		}
	}
	for _, input := range []string{"PN-18", "PN-1", "PN-AA"} {
		if _, err := definition.Parse(input); err == nil {
			t.Errorf("Parse(%q) accepted invalid input", input)
		}
	}
}

func TestSeqDefRangeDefaultsToMaximumWidth(t *testing.T) {
	definition := NewSeqDef(Range(0, 99))

	for _, input := range []string{"00", "07", "99"} {
		if _, err := definition.Parse(input); err != nil {
			t.Errorf("Parse(%q) error = %v", input, err)
		}
	}
	for _, input := range []string{"0", "000", "100", "AA"} {
		if _, err := definition.Parse(input); err == nil {
			t.Errorf("Parse(%q) accepted invalid input", input)
		}
	}
}

func TestSeqDefRangeRetainsPadding(t *testing.T) {
	node := Range(0, 17).Width(2, 'X')

	if node.Pad != 'X' {
		t.Fatalf("range padding = %q, want %q", node.Pad, 'X')
	}
	if _, err := NewSeqDef(node).Parse("X7"); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if value, err := NewSeqDef(node).Generate(nil); err != nil || value != "X0" {
		t.Fatalf("Generate() = %q, %v; want X0", value, err)
	}
}

func TestSeqDefChoice(t *testing.T) {
	definition := NewSeqDef(Choice(Literal("A"), Literal("BC"), Literal("D")))

	for _, input := range []string{"A", "BC", "D"} {
		if _, err := definition.Parse(input); err != nil {
			t.Errorf("Parse(%q) error = %v", input, err)
		}
	}
	for _, input := range []string{"B", "C", "E", "BCX"} {
		if _, err := definition.Parse(input); err == nil {
			t.Errorf("Parse(%q) accepted invalid input", input)
		}
	}
}

func TestSeqDefRangeRadix(t *testing.T) {
	definition := NewSeqDef(Concat(
		Literal("PN-"),
		RangeRadix(Choice(Literal("A"), Literal("B"), Literal("C")), Range(0, 9)),
	))

	for _, input := range []string{"PN-A0", "PN-B7", "PN-C9"} {
		if _, err := definition.Parse(input); err != nil {
			t.Errorf("Parse(%q) error = %v", input, err)
		}
	}
	if _, err := definition.Parse("PN-D0"); err == nil {
		t.Error("Parse accepted a value outside the radix")
	}
}

func TestSeqDefChoiceNestedGrammar(t *testing.T) {
	definition := NewSeqDef(Choice(
		Concat(Literal("RES-"), Range(0, 99).Width(2)),
		Concat(Literal("CAP-"), Choice(Literal("X5R"), Literal("X7R"))),
	))

	for _, input := range []string{"RES-00", "RES-99", "CAP-X5R", "CAP-X7R"} {
		if _, err := definition.Parse(input); err != nil {
			t.Errorf("Parse(%q) error = %v", input, err)
		}
	}
	for _, input := range []string{"RES-100", "CAP-10", "DIO-01", "CAP-X8R"} {
		if _, err := definition.Parse(input); err == nil {
			t.Errorf("Parse(%q) accepted invalid input", input)
		}
	}
}

func TestSeqDefBind(t *testing.T) {
	definition := NewSeqDef(Concat(
		Bind("category", Choice(Literal("RES"), Literal("CAP"))),
		Literal("-"),
		Bind("number", Range(0, 99).Width(2)),
	))

	result, err := definition.ParseValues("CAP-42")
	if err != nil {
		t.Fatalf("ParseValues() error = %v", err)
	}
	if result.Value != "CAP-42" {
		t.Fatalf("ParseValues().Value = %q, want %q", result.Value, "CAP-42")
	}
	for name, want := range map[string]string{"category": "CAP", "number": "42"} {
		if got := result.Bindings[name]; got != want {
			t.Errorf("binding %q = %q, want %q", name, got, want)
		}
	}
}

func TestSeqDefBindDoesNotLeakFromFailedChoice(t *testing.T) {
	definition := NewSeqDef(Choice(
		Concat(Bind("kind", Literal("OLD")), Literal("-"), Literal("bad")),
		Concat(Bind("kind", Literal("NEW")), Literal("-"), Literal("good")),
	))

	result, err := definition.ParseValues("NEW-good")
	if err != nil {
		t.Fatalf("ParseValues() error = %v", err)
	}
	if got := result.Bindings["kind"]; got != "NEW" {
		t.Fatalf("binding kind = %q, want %q", got, "NEW")
	}
}

func TestSeqDefStructuredCapturesPreserveOrderAndNormalization(t *testing.T) {
	definition := NewSeqDef(Concat(
		Bind("category", Literal("CAP")),
		Literal("-"),
		Bind("number", Range(0, 99).Width(2)),
	))

	result, err := definition.ParseValues("CAP-07")
	if err != nil {
		t.Fatalf("ParseValues() error = %v", err)
	}
	if len(result.Fields) != 2 {
		t.Fatalf("Fields length = %d, want 2", len(result.Fields))
	}
	if result.Fields[0].Name != "category" || result.Fields[0].Source != "CAP" || result.Fields[0].Normalized != "CAP" {
		t.Errorf("category capture = %#v", result.Fields[0])
	}
	if result.Fields[1].Name != "number" || result.Fields[1].Source != "07" || result.Fields[1].Normalized != 7 {
		t.Errorf("number capture = %#v, want source 07 and normalized 7", result.Fields[1])
	}
	if result.Bindings["number"] != "07" {
		t.Errorf("compatibility binding number = %q, want 07", result.Bindings["number"])
	}
}

func TestSeqDefChoiceRollbackRemovesStructuredCaptures(t *testing.T) {
	definition := NewSeqDef(Choice(
		Concat(Bind("kind", Literal("OLD")), Literal("-"), Literal("bad")),
		Concat(Bind("kind", Literal("NEW")), Literal("-"), Literal("good")),
	))

	result, err := definition.ParseValues("NEW-good")
	if err != nil {
		t.Fatalf("ParseValues() error = %v", err)
	}
	if len(result.Fields) != 1 || result.Fields[0].Name != "kind" || result.Fields[0].Source != "NEW" {
		t.Fatalf("Fields = %#v, want only the successful alternative capture", result.Fields)
	}
}

func TestSeqDefRejectsDuplicateBindingsOnSuccessfulPath(t *testing.T) {
	definition := NewSeqDef(Concat(Bind("value", Literal("A")), Bind("value", Literal("B"))))
	if _, err := definition.ParseValues("AB"); err == nil {
		t.Fatal("ParseValues() accepted duplicate binding")
	}

	// Reusing a name in alternatives is unambiguous because only the selected
	// alternative contributes captures.
	choice := NewSeqDef(Choice(Bind("value", Literal("A")), Bind("value", Literal("B"))))
	if result, err := choice.ParseValues("B"); err != nil || len(result.Fields) != 1 || result.Fields[0].Source != "B" {
		t.Fatalf("alternative duplicate binding = %#v, %v", result, err)
	}
}

func TestSeqDefStructuredRangeCaptureRetainsPadding(t *testing.T) {
	result, err := NewSeqDef(Bind("number", Range(0, 17).Width(2, 'X'))).ParseValues("X7")
	if err != nil {
		t.Fatalf("ParseValues() error = %v", err)
	}
	field := result.Fields[0]
	if field.Source != "X7" || field.Normalized != 7 {
		t.Fatalf("capture = %#v, want source X7 and normalized 7", field)
	}
}

func TestSeqDefParseRequiresFullInput(t *testing.T) {
	definition := NewSeqDef(Bind("value", Literal("A")))
	if _, err := definition.ParseValues("AB"); err == nil || !errors.Is(err, ErrInvalidSyntax) {
		t.Fatalf("ParseValues() error = %v, want invalid syntax", err)
	}
}

func TestSeqDefBindableFieldsInChoice(t *testing.T) {
	definition := NewSeqDef(Choice(
		Concat(Bind("category", Literal("RES")), Bind("number", Range(0, 99).Width(2))),
		Bind("category", Literal("CAP")),
	))

	fields := definition.BindableFields()
	if len(fields) != 2 {
		t.Fatalf("BindableFields() returned %d fields, want 2", len(fields))
	}
	if fields[0].Name != "category" || fields[1].Name != "number" {
		t.Fatalf("BindableFields() names = %q, %q, want category, number", fields[0].Name, fields[1].Name)
	}
	if fields[0].Node == nil || fields[1].Node == nil {
		t.Fatal("BindableFields() returned a nil field node")
	}
}

func TestSeqDefGenerate(t *testing.T) {
	definition := NewSeqDef(Concat(
		Literal("PN-"),
		Bind("category", Choice(Literal("A"), Literal("B"))),
		RangeRadix(Bind("number", Range(0, 99).Width(2)), Literal("-")),
		PlaceSequence("XYZ", digitsAlphabet),
	))

	value, err := definition.Generate(map[string]string{"category": "B", "number": "42"})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if value != "PN-B42-X0" {
		t.Fatalf("Generate() = %q, want %q", value, "PN-B42-X0")
	}

	value, err = definition.Generate(map[string]string{"category": "A", "number": "07"})
	if err != nil || value != "PN-A07-X0" {
		t.Fatalf("Generate() = %q, %v; want PN-A07-X0", value, err)
	}
}

func TestSeqDefGenerateRejectsInvalidBindings(t *testing.T) {
	definition := NewSeqDef(Concat(Bind("kind", Choice(Literal("A"), Literal("B"))), Bind("number", Range(0, 9))))
	tests := []struct {
		name     string
		bindings map[string]string
	}{
		{name: "missing", bindings: map[string]string{"kind": "A"}},
		{name: "invalid choice", bindings: map[string]string{"kind": "C", "number": "0"}},
		{name: "invalid range", bindings: map[string]string{"kind": "A", "number": "x"}},
		{name: "unknown", bindings: map[string]string{"kind": "A", "number": "0", "extra": "x"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := definition.Generate(test.bindings); err == nil {
				t.Fatal("Generate() accepted invalid bindings")
			} else if !errors.Is(err, ErrInvalidValue) {
				t.Fatalf("Generate() error = %v, want invalid value", err)
			}
		})
	}
}

func TestSeqDefGenerateRejectsBindingsIgnoredByChoice(t *testing.T) {
	definition := NewSeqDef(Choice(Literal("A"), Bind("value", Literal("B"))))
	if got, err := definition.Generate(map[string]string{"value": "B"}); err != nil || got != "B" {
		t.Fatalf("Generate() = %q, %v; want B", got, err)
	}
}

func TestSeqDefGenerateRejectsNoCompatibleChoiceAlternative(t *testing.T) {
	definition := NewSeqDef(Choice(Bind("value", Literal("A")), Bind("value", Literal("B"))))
	if _, err := definition.Generate(map[string]string{"value": "C"}); err == nil || !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("Generate() error = %v, want invalid value", err)
	}
}

func TestSeqDefSuccessorUsesRangeRadixOrdering(t *testing.T) {
	definition := NewSeqDef(Concat(
		Literal("PN-"),
		RangeRadix(Bind("major", Range(0, 1).Width(2)), Bind("minor", Range(0, 2).Width(2))),
	))
	tests := map[string]string{
		"PN-0000": "PN-0001",
		"PN-0001": "PN-0002",
		"PN-0002": "PN-0100",
	}
	for input, want := range tests {
		if got, err := definition.Successor(input); err != nil || got != want {
			t.Errorf("Successor(%q) = %q, %v; want %q", input, got, err, want)
		}
	}
}

func TestSeqDefSuccessorExhaustion(t *testing.T) {
	definition := NewSeqDef(Concat(Literal("PN-"), RangeRadix(Range(0, 1).Width(1), Range(0, 1).Width(1))))
	if _, err := definition.Successor("PN-11"); err == nil || !errors.Is(err, ErrNoOrdering) {
		t.Fatalf("Successor() error = %v, want ordering exhaustion", err)
	}
}

func TestSeqDefSuccessorRequiresMeaningfulOrdering(t *testing.T) {
	tests := []struct {
		definition SeqDef
		input      string
	}{
		{definition: NewSeqDef(Literal("PN-1")), input: "PN-1"},
		{definition: NewSeqDef(PlaceSequence("ABC")), input: "A"},
		{definition: NewSeqDef(Choice(Literal("A"), Literal("B"))), input: "A"},
	}
	for _, test := range tests {
		if _, err := test.definition.Successor(test.input); err == nil || !errors.Is(err, ErrNoOrdering) {
			t.Errorf("Successor() error = %v, want no-ordering error", err)
		}
	}
}

func Places(count int, alphabet string) SeqDefNode {
	places := make([]string, count)
	for i := range places {
		places[i] = alphabet
	}
	return PlaceSequence(places...)
}
