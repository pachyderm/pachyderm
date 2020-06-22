package errors

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
)

// As finds the first error in err's chain that matches the target's type, and
// if so, sets target to that error value and returns true.
// As is a wrapper for the underlying errors.As function, which may panic or
// return unexpected results based on how err was constructed (with or without
// a pointer).  This works by inspecting the type of target and attempting
// multiple errors.As calls if necessary.
func As(err error, target interface{}) bool {
	// Check the type of target, it must be a pointer to an error, or a pointer to a pointer to an error
	v := reflect.ValueOf(target)

	switch v.Kind() {
	case reflect.Struct:
		x := &target
		if errors.As(err, x) {
			fmt.Printf("ret 1\n")
			return true
		}

		// TODO: this branch never triggers
		fmt.Printf("ret 2\n")
		return errors.As(err, &x)
	case reflect.Ptr:
		// Unwrap inner type
		vi := v.Elem()

		if _, ok := v.Interface().(error); ok {
			// Wrap target in an extra pointer layer
			x := reflect.New(reflect.TypeOf(target))
			if errors.As(err, x.Interface()) {
				fmt.Printf("ret 3, target(%s): %v, x(%s): %v\n", reflect.TypeOf(target), target, reflect.TypeOf(x.Interface()), x.Interface())
				return true
			}

			if _, ok := vi.Interface().(error); ok {
				fmt.Printf("ret 4\n")
				return errors.As(err, target)
			}

			fmt.Printf("ret 4b\n")
			return false
		}

		if _, ok := vi.Interface().(error); ok {
			if errors.As(err, target) {
				fmt.Printf("ret 5\n")
				return true
			}
			fmt.Printf("ret 6\n")
			return errors.As(err, vi.Interface())
		}
	}
	fmt.Printf("Kind(): %v\n", v.Kind())
	panic("target must be a non-nil pointer to an error type, or a pointer to a pointer to an error type")
}
