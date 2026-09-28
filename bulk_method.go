package entx

// BulkCreateMethodAnnotation controls how bulk create helpers create records.
// if Sequential is true, it will preserve the values returned by the hooks ( if any ) instead of calling the bulk builders.
// this is useful for certain bulk creations that need to update existing records like invitations and others.
type BulkCreateMethodAnnotation struct {
	Sequential bool
}

// Name returns the annotation name
func (BulkCreateMethodAnnotation) Name() string {
	return BulkCreateAnnotationName
}

// Decode unmarshalls the BulkCreateMethodAnnotation
func (a *BulkCreateMethodAnnotation) Decode(annotation any) error {
	return DecodeAnnotation(annotation, a)
}

// SequentialBulkCreate instructs the resolver to sequentially create
// the items so we can handle cases where we want to handle
// existing values in the db.
func SequentialBulkCreate() *BulkCreateMethodAnnotation {
	return &BulkCreateMethodAnnotation{Sequential: true}
}
