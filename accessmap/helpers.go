package accessmap

// EdgeAuthCheck creates an annotation for edge access checks with the specified object type
// and requires the edit check
// this should be used for edges that have a mismatch between the edge name and the object type
// that holds the edge access check
func EdgeAuthCheck(objectType string) Annotations {
	return Annotations{
		ObjectType:    objectType,
		SkipEditCheck: false,
	}
}

// EdgeViewCheck creates an annotation for edge access checks with the specified object type
// and requires only view access to the object
// this should be used to for edges that only require view access to the object
// such as system owned edges
func EdgeViewCheck(objectType string) Annotations {
	return Annotations{
		ObjectType:      objectType,
		SkipEditCheck:   true,
		CheckViewAccess: true,
	}
}

// EdgeNoAuthCheck creates an annotation for edges that do not require an auth check
// this should be used for edges that do not require the user having edit access to the edge object
// or it is handled by other means, such as the parent schema
func EdgeNoAuthCheck() Annotations {
	return Annotations{
		SkipEditCheck: true,
	}
}
