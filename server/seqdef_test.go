package ubom

import "testing"

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

	if node.pad != 'X' {
		t.Fatalf("range padding = %q, want %q", node.pad, 'X')
	}
	if _, err := NewSeqDef(node).Parse("07"); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
}

func TestSeqDefValues(t *testing.T) {
	definition := NewSeqDef(Values("A", "BC", "D"))

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
		RangeRadix(Values("A", "B", "C"), Range(0, 9)),
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

func TestSeqDefBranch(t *testing.T) {
	definition := NewSeqDef(Branch(
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

func Places(count int, alphabet string) SeqNode {
	places := make([]string, count)
	for i := range places {
		places[i] = alphabet
	}
	return PlaceSequence(places...)
}
