package ubom

/* Taxonomy: a tree/DAG of catgories, where 'PartNumber' inhabits the leaves.

 */

type TaxonomyDefID string

type TaxonomyNodeID string

type TaxonomyDef struct {
	ID     TaxonomyDefID // durable unique ID
	SeqDef SeqDefID      // sequence this taxonomy is applied to
}

// Taxonomy is a tree that projects parsed bindings into labels.
type Taxonomy struct {
	Root TaxonomyNode
}

type TaxonomyNode struct {
	ID       TaxonomyNodeID
	Label    string
	Matches  map[string]string
	Children []TaxonomyNode
}

// Project returns the labels along the first matching path from Root.
// Unknown binding values are valid but stop projection at the last match.
func (t Taxonomy) Project(parsed ParseResult) []string {
	return projectTaxonomyNode(t.Root, parsed.Bindings)
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

func taxonomyMatches(matches, bindings map[string]string) bool {
	for name, want := range matches {
		if bindings[name] != want {
			return false
		}
	}
	return true
}
