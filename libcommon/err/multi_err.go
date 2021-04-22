package err

import "fmt"

type MultiErr struct {
	errs 	[]error
}

func (e *MultiErr) Error() string {
	var content string

	content = "multiple error: {\n"
	for _, err := range e.errs {
		content += fmt.Sprintf("%v\n", err)
	}
	content += "}"

	return content
}

func (e *MultiErr) Append(err error) {
	if len(e.errs) == 0 {
		e.errs = make([]error, 0)
	}

	e.errs = append(e.errs, err)
}

func (e *MultiErr) Zero() bool {
	return len(e.errs) == 0
}
