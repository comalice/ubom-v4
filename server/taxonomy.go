package ubom

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

/* # Taxonomy: a tree/DAG of catgories, where 'PartNumber' inhabits the leaves.

## Implementation notes:

Presently, we do not allow taxonomy nodes within a tree to be duplicates.

*/

// TaxonomyDefID identifies an immutable taxonomy-definition version.
type TaxonomyDefID string

// TaxonomyNodeID identifies a node within a taxonomy definition.
type TaxonomyNodeID string

type TaxonomyDef struct {
	ID            TaxonomyDefID // durable unique ID
	SeqDef        SeqDefID      // sequence this taxonomy is applied to
	AttributeDefs []AttributeDef
	Taxonomy      Taxonomy
}

// PredicateKind identifies the value comparison made by a taxonomy predicate.
type PredicateKind string

const (
	PredicateExact        PredicateKind = "exact"
	PredicateNumericRange PredicateKind = "numeric-range"
)

// TaxonomyPredicate is a typed condition over one captured SeqDef field.
// NumericRange is inclusive at both ends. Exact values intentionally compare
// the captured source text, preserving significant padding and case.
type TaxonomyPredicate struct {
	Field FieldName
	Kind  PredicateKind
	Value string
	Min   int64
	Max   int64
}

func ExactPredicate(field FieldName, value string) TaxonomyPredicate {
	return TaxonomyPredicate{Field: field, Kind: PredicateExact, Value: value}
}

func NumericRangePredicate(field FieldName, min, max int64) TaxonomyPredicate {
	return TaxonomyPredicate{Field: field, Kind: PredicateNumericRange, Min: min, Max: max}
}

// Taxonomy is a tree that projects parsed bindings into labels.
type Taxonomy struct {
	Root TaxonomyNode
}

type TaxonomyNode struct {
	ID         TaxonomyNodeID
	Label      string
	Matches    map[string]string
	Predicates []TaxonomyPredicate
	Attributes []AttributeAssignment
	Children   []TaxonomyNode
}

// ClassificationPolicy controls whether the generic taxonomy root is an
// acceptable result. Open taxonomies allow it; closed taxonomies require a
// matching descendant.
type ClassificationPolicy struct {
	// RequireClassification makes a generic root result unacceptable.
	RequireClassification bool
}

// Classification is the effective meaning of a parsed sequence. Path,
// Labels, Predicates, and Attributes are ordered from root to the deepest
// matching node.
type Classification struct {
	NodeID     TaxonomyNodeID
	Path       []TaxonomyNodeID
	Labels     []string
	Predicates []TaxonomyPredicate
	Attributes []AttributeAssignment
	Generic    bool
}

// ClassificationResult is a descriptive alias for Classification.
type ClassificationResult = Classification

// Validate checks that taxonomy nodes are well formed.
func (d TaxonomyDef) Validate() error {
	if err := validateTaxonomyID(string(d.ID), "taxonomy definition"); err != nil {
		return err
	}
	if err := validateTaxonomyID(string(d.SeqDef), "sequence definition"); err != nil {
		return err
	}
	seen := map[TaxonomyNodeID]bool{}
	attributeDefs := map[AttributeDefID]bool{}
	for _, attribute := range d.AttributeDefs {
		if err := attribute.Validate(); err != nil {
			return err
		}
		if attributeDefs[attribute.ID] {
			return fmt.Errorf("duplicate attribute definition ID %q", attribute.ID)
		}
		attributeDefs[attribute.ID] = true
	}
	if err := validateTaxonomyNode(d.Taxonomy.Root, seen, attributeDefs, nil); err != nil {
		return err
	}
	return nil
}

// ValidateAgainst validates field references at the taxonomy/SeqDef boundary.
// Keeping this explicit avoids making taxonomy validation depend on a store.
func (d TaxonomyDef) ValidateAgainst(seqDef SeqDef) error {
	if err := d.Validate(); err != nil {
		return err
	}
	if err := seqDef.Validate(); err != nil {
		return err
	}
	if d.SeqDef != seqDef.ID {
		return taxonomyValidation("taxonomy references sequence definition %q, got %q", d.SeqDef, seqDef.ID)
	}
	fields := make(map[string]SeqDefNode)
	for _, field := range seqDef.BindableFields() {
		fields[field.Name] = field.Node
	}
	return validateTaxonomyFields(d.Taxonomy.Root, fields)
}

// ValidateWithSeqDef is a descriptive alias for ValidateAgainst.
func (d TaxonomyDef) ValidateWithSeqDef(seqDef SeqDef) error { return d.ValidateAgainst(seqDef) }

func validateTaxonomyID(value, kind string) error {
	if value == "" || strings.IndexFunc(value, unicode.IsSpace) >= 0 {
		return taxonomyValidation("%s has invalid ID %q", kind, value)
	}
	return nil
}

func taxonomyValidation(format string, args ...any) error {
	return WrapDomainError(ErrorCategoryDefinition, fmt.Sprintf(format, args...), nil)
}

func validateTaxonomyFields(node TaxonomyNode, fields map[string]SeqDefNode) error {
	for _, predicate := range taxonomyPredicates(node) {
		fieldNode, ok := fields[predicate.Field.String()]
		if !ok {
			return taxonomyValidation("taxonomy node %q references unknown field %q", node.ID, predicate.Field)
		}
		if predicate.Kind == PredicateNumericRange {
			if _, ok := fieldNode.(RangeNode); !ok {
				return taxonomyValidation("numeric predicate for field %q requires a numeric RangeNode binding", predicate.Field)
			}
		}
	}
	for _, child := range node.Children {
		if err := validateTaxonomyFields(child, fields); err != nil {
			return err
		}
	}
	return nil
}

func validateTaxonomyNode(node TaxonomyNode, seen map[TaxonomyNodeID]bool, attributeDefs map[AttributeDefID]bool, ancestors []TaxonomyPredicate) error {
	if err := validateTaxonomyID(string(node.ID), "taxonomy node"); err != nil {
		return err
	}
	if seen[node.ID] {
		return fmt.Errorf("duplicate taxonomy node ID %q", node.ID)
	}
	seen[node.ID] = true
	predicates, err := validatedTaxonomyPredicates(node)
	if err != nil {
		return err
	}
	for _, predicate := range predicates {
		for _, ancestor := range ancestors {
			if !predicatesCompatible(ancestor, predicate) {
				return taxonomyValidation("taxonomy node %q is unreachable because its predicate conflicts with an ancestor", node.ID)
			}
		}
	}
	assignments := map[AttributeDefID]bool{}
	for _, assignment := range node.Attributes {
		if assignment.AttributeDefID == "" {
			return fmt.Errorf("taxonomy node %q has attribute assignment with no definition ID", node.ID)
		}
		if !attributeDefs[assignment.AttributeDefID] {
			return fmt.Errorf("taxonomy node %q references unknown attribute definition %q", node.ID, assignment.AttributeDefID)
		}
		if assignments[assignment.AttributeDefID] {
			return fmt.Errorf("taxonomy node %q has duplicate attribute assignment %q", node.ID, assignment.AttributeDefID)
		}
		assignments[assignment.AttributeDefID] = true
	}
	for i, child := range node.Children {
		for j := i + 1; j < len(node.Children); j++ {
			left, right := taxonomyPredicates(child), taxonomyPredicates(node.Children[j])
			if predicatesOverlap(left, right) {
				return taxonomyValidation("taxonomy node %q has ambiguous overlapping siblings %q and %q", node.ID, child.ID, node.Children[j].ID)
			}
		}
		if err := validateTaxonomyNode(child, seen, attributeDefs, append(append([]TaxonomyPredicate{}, ancestors...), predicates...)); err != nil {
			return err
		}
	}
	return nil
}

func validatedTaxonomyPredicates(node TaxonomyNode) ([]TaxonomyPredicate, error) {
	result := append([]TaxonomyPredicate{}, node.Predicates...)
	for field, value := range node.Matches {
		result = append(result, ExactPredicate(FieldName(field), value))
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

func taxonomyPredicates(node TaxonomyNode) []TaxonomyPredicate {
	result, _ := validatedTaxonomyPredicates(node)
	return result
}

func (p TaxonomyPredicate) Validate() error {
	if err := p.Field.Validate(); err != nil {
		return err
	}
	switch p.Kind {
	case PredicateExact:
		if p.Value == "" {
			return taxonomyValidation("exact predicate for field %q has an empty value", p.Field)
		}
	case PredicateNumericRange:
		if p.Min > p.Max {
			return taxonomyValidation("numeric predicate for field %q has invalid bounds %d..%d", p.Field, p.Min, p.Max)
		}
	default:
		return taxonomyValidation("predicate for field %q has unsupported kind %q", p.Field, p.Kind)
	}
	return nil
}

func predicatesCompatible(a, b TaxonomyPredicate) bool {
	if a.Field != b.Field {
		return true
	}
	if a.Kind == PredicateExact && b.Kind == PredicateExact {
		return a.Value == b.Value
	}
	if a.Kind == PredicateExact {
		n, err := strconv.ParseInt(a.Value, 10, 64)
		return err == nil && n >= b.Min && n <= b.Max
	}
	if b.Kind == PredicateExact {
		n, err := strconv.ParseInt(b.Value, 10, 64)
		return err == nil && n >= a.Min && n <= a.Max
	}
	return a.Min <= b.Max && b.Min <= a.Max
}

func predicatesOverlap(a, b []TaxonomyPredicate) bool {
	for _, left := range a {
		for _, right := range b {
			if !predicatesCompatible(left, right) {
				return false
			}
		}
	}
	return true
}

func (p TaxonomyPredicate) matches(value string) bool {
	if p.Kind == PredicateExact {
		return value == p.Value
	}
	n, err := strconv.ParseInt(value, 10, 64)
	return err == nil && n >= p.Min && n <= p.Max
}

func (p TaxonomyPredicate) matchesParsed(parsed ParsedSequence) bool {
	for _, field := range parsed.Fields {
		if field.Name != p.Field.String() {
			continue
		}
		if p.Kind == PredicateExact {
			return field.Source == p.Value
		}
		if value, ok := numericCapture(field.Normalized); ok {
			return value >= p.Min && value <= p.Max
		}
		return p.matches(field.Source)
	}
	if parsed.Bindings != nil {
		return p.matches(parsed.Bindings[p.Field.String()])
	}
	return false
}

func numericCapture(value any) (int64, bool) {
	switch n := value.(type) {
	case int:
		return int64(n), true
	case int8:
		return int64(n), true
	case int16:
		return int64(n), true
	case int32:
		return int64(n), true
	case int64:
		return n, true
	case uint:
		if uint64(n) <= uint64(^uint64(0)>>1) {
			return int64(n), true
		}
	case uint8:
		return int64(n), true
	case uint16:
		return int64(n), true
	case uint32:
		if uint64(n) <= uint64(^uint64(0)>>1) {
			return int64(n), true
		}
	case uint64:
		if n <= uint64(^uint64(0)>>1) {
			return int64(n), true
		}
	}
	return 0, false
}

func (n TaxonomyNode) predicatesMatch(bindings map[string]string) bool {
	for _, predicate := range taxonomyPredicates(n) {
		if !predicate.matches(bindings[predicate.Field.String()]) {
			return false
		}
	}
	return true
}

// Classify returns the deepest matching path. Sibling order is retained as a
// deterministic priority for compatibility with the legacy projection API;
// validated taxonomies reject overlapping siblings.
func (t Taxonomy) Classify(parsed ParsedSequence) (*Classification, error) {
	if err := validateClassificationPredicates(t.Root); err != nil {
		return nil, err
	}
	if t.Root.ID == "" {
		return nil, WrapDomainError(ErrorCategoryClassification, "taxonomy has no root classification", nil)
	}
	if !taxonomyNodePredicatesMatch(t.Root, parsed) {
		return nil, WrapDomainError(ErrorCategoryClassification, "parsed value does not match the taxonomy root", nil)
	}
	path := []TaxonomyNode{t.Root}
	for node := t.Root; ; {
		matched := -1
		for i := range node.Children {
			if taxonomyNodePredicatesMatch(node.Children[i], parsed) {
				matched = i
				break
			}
		}
		if matched < 0 {
			break
		}
		node = node.Children[matched]
		path = append(path, node)
	}

	result := &Classification{NodeID: path[len(path)-1].ID}
	result.Path = make([]TaxonomyNodeID, 0, len(path))
	positions := map[AttributeDefID]int{}
	for _, node := range path {
		result.Path = append(result.Path, node.ID)
		if node.Label != "" {
			result.Labels = append(result.Labels, node.Label)
		}
		result.Predicates = append(result.Predicates, taxonomyPredicates(node)...)
		for _, assignment := range node.Attributes {
			if position, exists := positions[assignment.AttributeDefID]; exists {
				result.Attributes[position] = assignment
			} else {
				positions[assignment.AttributeDefID] = len(result.Attributes)
				result.Attributes = append(result.Attributes, assignment)
			}
		}
	}
	result.Generic = len(path) == 1
	return result, nil
}

// ClassifyWithPolicy applies the open/closed taxonomy policy to Classify.
func (t Taxonomy) ClassifyWithPolicy(parsed ParsedSequence, policy ClassificationPolicy) (*Classification, error) {
	result, err := t.Classify(parsed)
	if err != nil {
		return nil, err
	}
	if policy.RequireClassification && result.Generic {
		return nil, WrapDomainError(ErrorCategoryClassification, "parsed value has no acceptable specific classification", nil)
	}
	return result, nil
}

func (d TaxonomyDef) Classify(parsed ParsedSequence) (*Classification, error) {
	return d.Taxonomy.Classify(parsed)
}

func (d TaxonomyDef) ClassifyWithPolicy(parsed ParsedSequence, policy ClassificationPolicy) (*Classification, error) {
	return d.Taxonomy.ClassifyWithPolicy(parsed, policy)
}

func taxonomyNodePredicatesMatch(node TaxonomyNode, parsed ParsedSequence) bool {
	for _, predicate := range taxonomyPredicates(node) {
		if !predicate.matchesParsed(parsed) {
			return false
		}
	}
	return true
}

// Classification validates all predicate-bearing nodes before evaluating the
// value. The legacy projection helpers intentionally retain their tolerant
// taxonomyPredicates adapter, but public classification must not hide errors.
func validateClassificationPredicates(node TaxonomyNode) error {
	if _, err := validatedTaxonomyPredicates(node); err != nil {
		return err
	}
	for _, child := range node.Children {
		if err := validateClassificationPredicates(child); err != nil {
			return err
		}
	}
	return nil
}

// EffectiveAttributes returns assignments from root to node. A child replaces
// an ancestor assignment with the same definition ID, like an overlay file.
func (t Taxonomy) EffectiveAttributes(nodeID TaxonomyNodeID) ([]AttributeAssignment, error) {
	path, ok := taxonomyNodePath(t.Root, nodeID)
	if !ok {
		return nil, fmt.Errorf("taxonomy node %q not found", nodeID)
	}
	result := []AttributeAssignment{}
	positions := map[AttributeDefID]int{}
	for _, node := range path {
		for _, assignment := range node.Attributes {
			if position, exists := positions[assignment.AttributeDefID]; exists {
				result[position] = assignment
				continue
			}
			positions[assignment.AttributeDefID] = len(result)
			result = append(result, assignment)
		}
	}
	return result, nil
}

func taxonomyNodePath(node TaxonomyNode, id TaxonomyNodeID) ([]TaxonomyNode, bool) {
	if node.ID == id {
		return []TaxonomyNode{node}, true
	}
	for _, child := range node.Children {
		path, ok := taxonomyNodePath(child, id)
		if ok {
			return append([]TaxonomyNode{node}, path...), true
		}
	}
	return nil, false
}

// Project returns the labels along the first matching path from Root. When
// sibling rules overlap, declaration order is the deterministic priority.
// Unknown binding values are valid but stop projection at the last match.
func (t Taxonomy) Project(parsed ParseResult) []string {
	return projectTaxonomyNode(t.Root, parsed.Bindings)
}

// ProjectNode returns the deepest matching taxonomy node, using declaration
// order as the priority when sibling rules overlap.
func (t Taxonomy) ProjectNode(parsed ParseResult) (TaxonomyNodeID, error) {
	node, ok := projectTaxonomyNodeID(t.Root, parsed.Bindings)
	if !ok {
		return "", fmt.Errorf("bindings do not match a taxonomy node")
	}
	return node, nil
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

func taxonomyMatches(matches, bindings map[string]string) bool {
	for name, want := range matches {
		if bindings[name] != want {
			return false
		}
	}
	return true
}
