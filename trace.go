package envflag

// scantracer provides error tracing functionality.
type scantracer struct {
	path

	// pointers maps each pointer to the paths leading to it.
	pointers map[interface{}][]string

	// skipped contains paths to skipped values.
	skipped []string

	// duplicates stores the number of addresses targeted by more than one pointer.
	duplicates int
}

// warning retrieves problems occuring during the scan in an accessible format.
func (s *scantracer) warning() *ScanWarnings {
	if s == nil || s.duplicates == 0 && len(s.skipped) == 0 {
		return nil
	}
	var dups [][]string
	if s.duplicates > 0 {
		dups = make([][]string, 0, s.duplicates)
		for _, group := range s.pointers {
			if len(group) > 1 {
				dups = append(dups, group)
			}
		}
	}
	return &ScanWarnings{
		Duplicates: dups,
		Skipped:    s.skipped,
	}
}

func (s *scantracer) register(ptr interface{}) bool {
	paths, found := s.pointers[ptr]
	if len(paths) == 1 {
		s.duplicates++
	}
	s.pointers[ptr] = append(paths, s.String())
	return found
}

func (s *scantracer) skip() {
	s.skipped = append(s.skipped, s.String())
}

// ScanWarnings provides warnings generated during the scanning process.
// It holds paths to problematic fields, where a path is the sequence of
// struct field names starting at the root that have to be traversed to reach a field.
//
// A path starts at the root element and contains field names or slice indices separated
// by a slash ('/').
type ScanWarnings struct {

	// Duplicates contains slices with paths to fields holding pointers
	// to the same address.
	Duplicates [][]string

	// Skipped contains paths to skipped fields.
	//
	// A struct field is skipped if it is a duplicate and already contained in
	// a parameter, it is unexported or nil or if it can neither be
	// converted by ValueOf nor scanned as an inner struct.
	Skipped []string
}

func (w *ScanWarnings) Error() string {
	msg := []byte{}
	if len(w.Duplicates) > 0 {
		msg = append(msg, "found duplicates ("...)
		for i, group := range w.Duplicates {
			if i > 0 {
				msg = append(msg, "), ("...)
			}
			msg = appendgroup(msg, group)
		}
		msg = append(msg, ')')
	}
	if len(w.Skipped) > 0 {
		if len(w.Duplicates) > 0 {
			msg = append(msg, " and "...)
		}
		msg = append(msg, "skipped "...)
		msg = appendgroup(msg, w.Skipped)
	}
	return string(msg)
}

func appendgroup(buf []byte, group []string) []byte {
	for i, path := range group {
		if i > 0 {
			buf = append(buf, ',', ' ')
		}
		buf = append(buf, path...)
	}
	return buf
}
