package event

var typeOfError = reflect.TypeOf((*error)(nil)).Elem()

// asError converts val to an error without wrapping it.
func asError(val interface{}) error {
	switch x := val.(type) {
	case nil:
		return nil
	case error:
		return x
	default:
		return fmt.Errorf("%v", x)
	}
}

func combineErrors(errs []error) error {
	switch len(errs) {
	case 0:
		return nil
	case 1:
		return errs[0]
	}
	var b strings.Builder
	for i, err := range errs {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(err.Error())
	}
	return errors.New(b.String())
}