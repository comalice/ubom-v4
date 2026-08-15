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
		ubom.Range(0, 9999).Width(4),
	)).WithID("sample-pn-v1")
	if err := s.CreateSeqDef(seqDef); err != nil {
		return Result{}, err
	}
	revisionSeqDef := ubom.NewSeqDef(ubom.Range(1, 9999).Width(4)).WithID("sample-revision-v1")
	if err := s.CreateSeqDef(revisionSeqDef); err != nil {
		return Result{}, err
	}
	taxonomyDef := ubom.TaxonomyDef{
		ID:     "sample-taxonomy-v1",
		SeqDef: seqDef.ID,
		Taxonomy: ubom.Taxonomy{Root: ubom.TaxonomyNode{
			ID:    "components",
			Label: "Components",
			Children: []ubom.TaxonomyNode{
				{ID: "resistors", Label: "Resistors", Matches: map[string]string{"category": "A"}},
				{ID: "capacitors", Label: "Capacitors", Matches: map[string]string{"category": "B"}},
			},
		}},
	}
	if err := s.CreateTaxonomyDef(taxonomyDef); err != nil {
		return Result{}, err
	}

	parts := make([]ubom.PartNumber, 0, options.Parts)
	for i := 0; i < options.Parts; i++ {
		category := "A"
		if i%2 == 1 {
			category = "B"
		}
		part, err := ubom.NewPartNumber(fmt.Sprintf("PN-%s%04d", category, i/2), seqDef, taxonomyDef)
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
