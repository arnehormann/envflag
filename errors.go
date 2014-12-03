package envflag

import "fmt"

type Errors struct {
	errs []error
}

func (es *Errors) add(err error) {
	if err == nil {
		return
	}
	es.errs = append(es.errs, err)
}

func (es Errors) Len() int {
	return len(es.errs)
}

func (es Errors) Err(i int) error {
	if i < 0 || i >= len(es.errs) {
		return nil
	}
	return es.errs[i]
}

func (es Errors) String() string {
	switch len(es.errs) {
	case 0:
		return "<no error>"
	case 1:
		return es.errs[0].Error()
	}
	return fmt.Sprintf("%#v", es.errs)
}

func (es Errors) Error() error {
	if len(es.errs) == 0 {
		return nil
	}
	return fmt.Errorf(es.String())
}
