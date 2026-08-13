package ubom

/* Taxonomy: a tree/DAG of catgories, where 'PartNumber' inhabits the leaves.

 */

type TaxonomyDefID string

type TaxonomyDef struct {
	ID     TaxonomyDefID // durable unique ID
	SeqDef SeqDefID      // sequence this taxonomy is applied to
}
