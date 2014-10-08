package testdata

// unexported struct with unexported field
type dataU struct {
	v interface{}
}

// unexported struct with exported field
type dataE struct {
	V interface{}
}

// exported struct with unexported field
type DataU struct {
	v interface{}
}

// exported struct with exported field
type DataE struct {
	V interface{}
}

// Struct retrieves a pointer to a struct containing v in its only field.
func Struct(v interface{}, exportedStruct, exportedField bool) interface{} {
	if exportedStruct {
		if exportedField {
			return &DataE{v}
		}
		return &DataU{v}
	}
	if exportedField {
		return &dataE{v}
	}
	return &dataU{v}
}
