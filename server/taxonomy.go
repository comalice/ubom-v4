package ubom

import "fmt"

/* # Taxonomy: a tree/DAG of catgories, where 'PartNumber' inhabits the leaves.

## Implementation notes:

Presently, we do not allow taxonomy nodes within a tree to be duplicates.

*/

type TaxonomyDefID string

type TaxonomyNodeID string

type TaxonomyDef struct {
	ID            TaxonomyDefID // durable unique ID
	SeqDef        SeqDefID      // sequence this taxonomy is applied to
	AttributeDefs []AttributeDef
	Taxonomy      Taxonomy
}

// Taxonomy is a tree that projects parsed bindings into labels.
type Taxonomy struct {
	Root TaxonomyNode
}

type TaxonomyNode struct {
	ID         TaxonomyNodeID
	Label      string
	Matches    map[string]string
	Attributes []AttributeAssignment
	Children   []TaxonomyNode
}

// Validate checks that taxonomy nodes are well formed.
func (d TaxonomyDef) Validate() error {
	if d.ID == "" {
		return fmt.Errorf("taxonomy definition has no ID")
	}
	if d.SeqDef == "" {
		return fmt.Errorf("taxonomy definition has no sequence definition ID")
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
	if err := validateTaxonomyNode(d.Taxonomy.Root, seen, attributeDefs); err != nil {
		return err
	}
	return nil
}

func validateTaxonomyNode(node TaxonomyNode, seen map[TaxonomyNodeID]bool, attributeDefs map[AttributeDefID]bool) error {
	if node.ID == "" {
		return fmt.Errorf("taxonomy node has no ID")
	}
	if seen[node.ID] {
		return fmt.Errorf("duplicate taxonomy node ID %q", node.ID)
	}
	seen[node.ID] = true
	for _, assignment := range node.Attributes {
		if assignment.AttributeDefID == "" {
			return fmt.Errorf("taxonomy node %q has attribute assignment with no definition ID", node.ID)
		}
		if !attributeDefs[assignment.AttributeDefID] {
			return fmt.Errorf("taxonomy node %q references unknown attribute definition %q", node.ID, assignment.AttributeDefID)
		}
	}
	for _, child := range node.Children {
		if err := validateTaxonomyNode(child, seen, attributeDefs); err != nil {
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

// Project returns the labels along the first matching path from Root.
// Unknown binding values are valid but stop projection at the last match.
func (t Taxonomy) Project(parsed ParseResult) []string {
	return projectTaxonomyNode(t.Root, parsed.Bindings)
}

// ProjectNode returns the deepest matching taxonomy node.
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
		if taxonomyMatches(child.Matches, bindings) {
			return append(labels, projectTaxonomyNode(child, bindings)...)
		}
	}
	return labels
}

func projectTaxonomyNodeID(node TaxonomyNode, bindings map[string]string) (TaxonomyNodeID, bool) {
	best := node.ID
	found := node.ID != ""
	for _, child := range node.Children {
		if !taxonomyMatches(child.Matches, bindings) {
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
