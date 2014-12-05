package envflag

import "fmt"

type errslice []error

func (es errslice) String() string {
	switch len(es) {
	case 0:
		return "<no error>"
	case 1:
		return es[0].Error()
	}
	return fmt.Sprintf("%#v", es)
}

func (es errslice) Join() error {
	if len(es) == 0 {
		return nil
	}
	return fmt.Errorf(es.String())
}
