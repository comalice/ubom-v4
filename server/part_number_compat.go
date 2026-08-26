package ubom

// NewPartNumber is the deprecated compatibility adapter for
// Schema.NewPartNumber. New code should use Schema.NewPartNumber so attribute
// validation and schema policy are explicit.
//
// The historical API has no attribute parameter and callers commonly fill
// attributes immediately afterward. Keep that behavior while sharing schema
// parsing and typed classification.
func NewPartNumber(value string, seqDef SeqDef, taxonomy TaxonomyDef) (PartNumber, error) {
	return NewSchema(seqDef, taxonomy, SchemaPolicy{}).materializePartNumber(value, nil, false)
}
