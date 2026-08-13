package ubom

// Part intentionally left bare for now. Leaf node in a BOM tree.
// TODO sort out _what_ goes into Part. Eventually the concept of Artifact, ArtifactSource,
// mfr mpn, etc?
type Part struct {
}

type LineItem struct {
	PartNumber   PartNumber
	PartRevision PartRevision
}

type PartRevision struct {
	BOM  []LineItem
	Part Part
}

// PartNumber, a unique value, taxonomy, and schema used to identify a given part. May be a
// partial construction or fully qualified -- where a fully qualified part number possesses values
// for every level of the related taxonomy definition.
type PartNumber struct {
	Value         string         // SeqDef parseable canonical part number
	PartRevision  []PartRevision // Part revision.
	SeqDefID      SeqDefID       // SeqDef version used to generate this part number.
	TaxonomyDefID TaxonomyDefID  // Taxonomy applied to this part number.
	// TODO, post MVP, sort out attributes and their types.
	// AttributeSchemas []AttributeSchemaID // Attribute schema attached to this part number.
}

// Next generates the next part number from the part number's SeqDef.
// TODO MVP+1 wire in attribute fields -- require an attribute object with (key,value) pairs for
// all defined attribues in the taxonomy tree AND/OR on the part number.

// NextFromPartial generates the next part number from the part number's SeqDef AND relevant SeqDef
// field selections.
// TODO MVP+1 wire in attribute fields.

// Validate consumes a part number struct OR part number string + SeqDef and determines if the part
// number could be produced from the SeqDef.

// ProjectTaxonomy produces a PartNumberTaxonomy object with taxonomy nodes attached to each
// segment of the part number.
