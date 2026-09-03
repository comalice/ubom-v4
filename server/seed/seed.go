package seed

import (
	"fmt"
	"math/rand"

	ubom "ubom-v4"
	"ubom-v4/store"
)

type Options struct {
	Parts        int
	MaxRevisions int
	MaxBOMDepth  int
	Seed         int64
}

type Result struct {
	Parts     int
	Revisions int
	BOMLines  int
}

func Populate(s store.Store, options Options) (Result, error) {
	if options.Parts < 0 || options.MaxRevisions < 0 || options.MaxBOMDepth < 0 {
		return Result{}, fmt.Errorf("seed counts and depths cannot be negative")
	}
	if options.Parts > 10000 {
		return Result{}, fmt.Errorf("parts cannot exceed 10000")
	}

	seqDef := ubom.NewSeqDef(ubom.Concat(
		ubom.Literal("PN-"),
		ubom.Bind("category", ubom.Choice(ubom.Literal("A"), ubom.Literal("B"))),
		ubom.Bind("number", ubom.Range(0, 9999).Width(4)),
	)).WithID("sample-pn-v1")
	if err := s.CreateSeqDef(seqDef); err != nil {
		return Result{}, err
	}
	gtdPNSeqDef := ubom.NewSeqDef(ubom.Choice(
		ubom.Concat(
			ubom.Bind("family", ubom.Range(100, 999).Width(3)),
			ubom.Literal("-"),
			ubom.Bind("group", ubom.Range(0, 99999).Width(5)),
			ubom.Literal("-"),
			ubom.Bind("uid", ubom.Range(0, 999).Width(3)),
		),
		ubom.Concat(
			ubom.Bind("family", ubom.Range(0, 99).Width(3)),
			ubom.Literal("-"),
			ubom.Bind("uid", ubom.Range(0, 99999).Width(5)),
		),
	)).WithID("gtd-pn-v1")
	if err := s.CreateSeqDef(gtdPNSeqDef); err != nil {
		return Result{}, err
	}
	gtdPNTaxonomyDef := ubom.TaxonomyDef{
		ID:     "gtd-pn-taxonomy-v1",
		SeqDef: gtdPNSeqDef.ID,
		Taxonomy: ubom.Taxonomy{Root: ubom.TaxonomyNode{
			ID:    "gtd-part-numbers",
			Label: "Part Numbers",
			Children: []ubom.TaxonomyNode{
				{ID: "raw-material", Label: "Raw Material", Predicates: []ubom.TaxonomyPredicate{ubom.NumericRangePredicate("family", 0, 99)}},
				{
					ID: "family-100-199", Label: "100-199", Predicates: []ubom.TaxonomyPredicate{ubom.NumericRangePredicate("family", 100, 199)},
					Children: []ubom.TaxonomyNode{
						{ID: "family-100", Label: "100", Predicates: []ubom.TaxonomyPredicate{ubom.ExactPredicate("family", "100")}},
						{ID: "family-105", Label: "105", Predicates: []ubom.TaxonomyPredicate{ubom.ExactPredicate("family", "105")}},
						{ID: "family-150", Label: "150", Predicates: []ubom.TaxonomyPredicate{ubom.ExactPredicate("family", "150")}},
						{ID: "family-151", Label: "151", Predicates: []ubom.TaxonomyPredicate{ubom.ExactPredicate("family", "151")}},
						{ID: "family-155", Label: "155", Predicates: []ubom.TaxonomyPredicate{ubom.ExactPredicate("family", "155")}},
						{ID: "family-199", Label: "199", Predicates: []ubom.TaxonomyPredicate{ubom.ExactPredicate("family", "199")}},
					},
				},
				{ID: "family-200", Label: "200", Predicates: []ubom.TaxonomyPredicate{ubom.ExactPredicate("family", "200")}},
				{
					ID: "family-300-399", Label: "300-399", Predicates: []ubom.TaxonomyPredicate{ubom.NumericRangePredicate("family", 300, 399)},
					Children: []ubom.TaxonomyNode{
						{ID: "family-310", Label: "310", Predicates: []ubom.TaxonomyPredicate{ubom.ExactPredicate("family", "310")}},
						{ID: "family-320", Label: "320", Predicates: []ubom.TaxonomyPredicate{ubom.ExactPredicate("family", "320")}},
						{ID: "family-330", Label: "330", Predicates: []ubom.TaxonomyPredicate{ubom.ExactPredicate("family", "330")}},
					},
				},
				{ID: "family-400", Label: "400", Predicates: []ubom.TaxonomyPredicate{ubom.ExactPredicate("family", "400")}},
				{
					ID: "family-500-599", Label: "500-599", Predicates: []ubom.TaxonomyPredicate{ubom.NumericRangePredicate("family", 500, 599)},
					Children: []ubom.TaxonomyNode{
						{ID: "family-500", Label: "500", Predicates: []ubom.TaxonomyPredicate{ubom.ExactPredicate("family", "500")}},
						{ID: "family-505", Label: "505", Predicates: []ubom.TaxonomyPredicate{ubom.ExactPredicate("family", "505")}},
					},
				},
				{ID: "family-600", Label: "600", Predicates: []ubom.TaxonomyPredicate{ubom.ExactPredicate("family", "600")}},
				{ID: "family-700", Label: "700", Predicates: []ubom.TaxonomyPredicate{ubom.ExactPredicate("family", "700")}},
				{ID: "family-800", Label: "800", Predicates: []ubom.TaxonomyPredicate{ubom.ExactPredicate("family", "800")}},
				{ID: "family-900", Label: "900", Predicates: []ubom.TaxonomyPredicate{ubom.ExactPredicate("family", "900")}},
				{ID: "family-910", Label: "910", Predicates: []ubom.TaxonomyPredicate{ubom.ExactPredicate("family", "910")}},
				{ID: "family-911", Label: "911", Predicates: []ubom.TaxonomyPredicate{ubom.ExactPredicate("family", "911")}},
				{ID: "family-912", Label: "912", Predicates: []ubom.TaxonomyPredicate{ubom.ExactPredicate("family", "912")}},
				{ID: "family-950", Label: "950", Predicates: []ubom.TaxonomyPredicate{ubom.ExactPredicate("family", "950")}},
				{ID: "family-990", Label: "990", Predicates: []ubom.TaxonomyPredicate{ubom.ExactPredicate("family", "990")}},
			},
		}},
	}
	if err := s.CreateTaxonomyDef(gtdPNTaxonomyDef); err != nil {
		return Result{}, err
	}
	revisionSeqDef := ubom.NewSeqDef(ubom.Range(1, 9999).Width(4)).WithID("sample-revision-v1")
	if err := s.CreateSeqDef(revisionSeqDef); err != nil {
		return Result{}, err
	}
	taxonomyDef := ubom.TaxonomyDef{
		ID:     "sample-taxonomy-v1",
		SeqDef: seqDef.ID,
		AttributeDefs: []ubom.AttributeDef{
			{ID: "footprint", Label: "Footprint", ValueType: ubom.AttributeValueEnum, EnumValues: []string{"0603", "0805"}},
			{ID: "resistance", Label: "Resistance value", ValueType: ubom.AttributeValueDecimal, Unit: "ohm"},
			{ID: "tolerance", Label: "Tolerance", ValueType: ubom.AttributeValueDecimal, Unit: "percent"},
			{ID: "power_rating", Label: "Power rating", ValueType: ubom.AttributeValueDecimal, Unit: "W"},
		},
		Taxonomy: ubom.Taxonomy{Root: ubom.TaxonomyNode{
			ID:    "components",
			Label: "Components",
			Attributes: []ubom.AttributeAssignment{
				{AttributeDefID: "footprint", Required: true},
			},
			Children: []ubom.TaxonomyNode{
				{ID: "resistors", Label: "Resistors", Predicates: []ubom.TaxonomyPredicate{ubom.ExactPredicate("category", "A")}, Attributes: []ubom.AttributeAssignment{
					{AttributeDefID: "resistance", Required: true},
					{AttributeDefID: "tolerance"},
					{AttributeDefID: "power_rating"},
				}},
				{ID: "capacitors", Label: "Capacitors", Predicates: []ubom.TaxonomyPredicate{ubom.ExactPredicate("category", "B")}},
			},
		}},
	}
	if err := s.CreateTaxonomyDef(taxonomyDef); err != nil {
		return Result{}, err
	}
	schema := ubom.NewSchema(seqDef, taxonomyDef, ubom.SchemaPolicy{})

	parts := make([]ubom.PartNumber, 0, options.Parts)
	for i := 0; i < options.Parts; i++ {
		category := "A"
		if i%2 == 1 {
			category = "B"
		}
		attributes := []ubom.PartNumberAttribute{{AttributeDefID: "footprint", Value: "0603"}}
		if category == "A" {
			attributes = append(attributes,
				ubom.PartNumberAttribute{AttributeDefID: "resistance", Value: fmt.Sprintf("%d", 1000+(i/2)*1000)},
				ubom.PartNumberAttribute{AttributeDefID: "tolerance", Value: "1"},
				ubom.PartNumberAttribute{AttributeDefID: "power_rating", Value: "0.1"},
			)
		}
		part, err := schema.NewPartNumber(fmt.Sprintf("PN-%s%04d", category, i/2), attributes)
		if err != nil {
			return Result{}, err
		}
		part, err = s.CreatePartNumber(part)
		if err != nil {
			return Result{}, err
		}
		parts = append(parts, part)
	}

	rng := rand.New(rand.NewSource(options.Seed))
	createdRevisions := make([][]ubom.PartRevision, len(parts))
	depths := make([]int, len(parts))
	result := Result{Parts: len(parts)}
	for i, part := range parts {
		if options.MaxBOMDepth > 0 {
			depths[i] = (i % options.MaxBOMDepth) + 1
		}
		revisionCount := 0
		if options.MaxRevisions > 0 {
			revisionCount = 1 + rng.Intn(options.MaxRevisions)
		}
		for range revisionCount {
			revision := ubom.PartRevision{
				PartNumberID:     part.ID,
				Revision:         fmt.Sprintf("%04d", len(createdRevisions[i])+1),
				RevisionSeqDefID: revisionSeqDef.ID,
			}
			if depths[i] > 0 && i > 0 {
				candidateCount := rng.Intn(3)
				for j := 0; j < candidateCount; j++ {
					candidates := make([]int, 0, i)
					for childIndex := 0; childIndex < i; childIndex++ {
						if depths[childIndex] < depths[i] && len(createdRevisions[childIndex]) > 0 {
							candidates = append(candidates, childIndex)
						}
					}
					if len(candidates) == 0 {
						break
					}
					childIndex := candidates[rng.Intn(len(candidates))]
					childRevision := createdRevisions[childIndex][rng.Intn(len(createdRevisions[childIndex]))]
					revision.BOM = append(revision.BOM, ubom.LineItem{
						PartNumberID:   parts[childIndex].ID,
						PartRevisionID: childRevision.ID,
					})
				}
			}
			created, err := s.CreatePartRevision(revision)
			if err != nil {
				return Result{}, err
			}
			createdRevisions[i] = append(createdRevisions[i], created)
			result.Revisions++
			result.BOMLines += len(created.BOM)
		}
	}
	return result, nil
}
