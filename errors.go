package envflag

import (
	"fmt"
	"strings"
)

type errors struct {
	errs []error
}

func (e *errors) add(err error) {
	if err == nil {
		return
	}
	e.errs = append(e.errs, err)
}

func (e *errors) has() bool {
	return len(e.errs) > 0
}

func (e *errors) get() error {
	msgs := make([]string, len(e.errs))
	for i, err := range e.errs {
		msgs[i] = err.Error()
	}
	return fmt.Errorf(strings.Join(msgs, "\n"))
}
