package ubom

import "fmt"

// This file contains compatibility-only taxonomy behavior. Production
// interpretation uses Taxonomy.Classify with typed Predicates and ParsedSequence.

// validatedTaxonomyPredicates includes the pre-typed Matches representation for
// definition validation and compatibility callers. It is not used by Classify.
func validatedTaxonomyPredicates(node TaxonomyNode) ([]TaxonomyPredicate, error) {
	var result []TaxonomyPredicate
	if node.Predicates != nil {
		result = append(result, node.Predicates...)
	}
	for field, value := range node.Matches {
		legacy := ExactPredicate(FieldName(field), value)
		equivalent := false
		for _, predicate := range result {
			if predicate == legacy {
				equivalent = true
				break
			}
		}
		if !equivalent {
			result = append(result, legacy)
		}
	}
	seen := map[FieldName]bool{}
	for _, predicate := range result {
		if err := predicate.Validate(); err != nil {
			return nil, fmt.Errorf("taxonomy node %q: %w", node.ID, err)
		}
		if seen[predicate.Field] {
			return nil, taxonomyValidation("taxonomy node %q has multiple predicates for field %q", node.ID, predicate.Field)
		}
		seen[predicate.Field] = true
	}
	return result, nil
}

// UpgradeLegacyTaxonomy converts legacy exact-match data into typed predicates
// for an interpretation boundary. Matches is retained so old persisted/API
// representations remain lossless. It returns an error for conflicting data.
func UpgradeLegacyTaxonomy(t Taxonomy) (Taxonomy, error) {
	var upgrade func(TaxonomyNode) (TaxonomyNode, error)
	upgrade = func(node TaxonomyNode) (TaxonomyNode, error) {
		predicates, err := validatedTaxonomyPredicates(node)
		if err != nil {
			return TaxonomyNode{}, err
		}
		node.Predicates = predicates
		node.Children = append([]TaxonomyNode(nil), node.Children...)
		for i, child := range node.Children {
			node.Children[i], err = upgrade(child)
			if err != nil {
				return TaxonomyNode{}, err
			}
		}
		return node, nil
	}
	root, err := upgrade(t.Root)
	if err != nil {
		return Taxonomy{}, err
	}
	t.Root = root
	return t, nil
}

// Project is the deprecated binding-only projection API. Use Schema.Parse or
// Taxonomy.Classify for all new interpretation code.
func (t Taxonomy) Project(parsed ParseResult) []string {
	return projectTaxonomyNode(t.Root, parsed.Bindings)
}

// ProjectNode is the deprecated binding-only projection API. Declaration order
// is retained solely for compatibility; typed validation rejects ambiguity.
func (t Taxonomy) ProjectNode(parsed ParseResult) (TaxonomyNodeID, error) {
	node, ok := projectTaxonomyNodeID(t.Root, parsed.Bindings)
	if !ok {
		return "", fmt.Errorf("bindings do not match a taxonomy node")
	}
	return node, nil
}

func (n TaxonomyNode) predicatesMatch(bindings map[string]string) bool {
	for _, predicate := range legacyTaxonomyPredicates(n) {
		if !predicate.matches(bindings[predicate.Field.String()]) {
			return false
		}
	}
	return true
}

func legacyTaxonomyPredicates(node TaxonomyNode) []TaxonomyPredicate {
	result, _ := validatedTaxonomyPredicates(node)
	return result
}

func projectTaxonomyNode(node TaxonomyNode, bindings map[string]string) []string {
	labels := []string{}
	if node.Label != "" {
		labels = append(labels, node.Label)
	}
	for _, child := range node.Children {
		if child.predicatesMatch(bindings) {
			return append(labels, projectTaxonomyNode(child, bindings)...)
		}
	}
	return labels
}

func projectTaxonomyNodeID(node TaxonomyNode, bindings map[string]string) (TaxonomyNodeID, bool) {
	best := node.ID
	found := node.ID != ""
	for _, child := range node.Children {
		if !child.predicatesMatch(bindings) {
			continue
		}
		if childBest, childFound := projectTaxonomyNodeID(child, bindings); childFound {
			return childBest, true
		}
		if child.ID != "" {
			return child.ID, true
		}
	}
	return best, found
}
