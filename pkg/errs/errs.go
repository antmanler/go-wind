package errs

import multierror "github.com/hashicorp/go-multierror"

// And chains functions together and retuns one error
func And(err error, errors ...error) error {
	if len(errors) == 0 {
		return err
	}
	for _, e := range errors {
		if e != nil {
			err = multierror.Append(err, e)
		}
	}

	return err
}
