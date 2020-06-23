package errors

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
)

var errorType = reflect.TypeOf((*error)(nil)).Elem()

func tryAs(err error, targetVal reflect.Value) bool {
	e := targetVal.Type().Elem()
	if e.Kind() == reflect.Interface || e.Implements(errorType) {
		res := errors.As(err, targetVal.Interface())
		fmt.Printf("as (%s): %v\n", targetVal.Type(), res)
		return res
	}
	return false
}

// As finds the first error in err's chain that matches the target's type, and
// if so, sets target to that error value and returns true.
// As is a wrapper for the underlying errors.As function, which may panic or
// return unexpected results based on how err was constructed (with or without
// a pointer).  This works by inspecting the type of target and attempting
// multiple errors.As calls if necessary.
func As(err error, target interface{}) bool {
	// Check the type of target, it must be a pointer to an error, or a pointer to a pointer to an error
	v := reflect.ValueOf(target)
	fmt.Printf("%s (%v): %v\n", reflect.TypeOf(target), v.Kind(), target)

	switch v.Kind() {
	case reflect.Ptr:
		e := v.Type().Elem()
		vp := reflect.New(v.Type())
		vpe := vp.Type().Elem()
		fmt.Printf("%s (%v): %v\n", vp.Type(), vp.Kind(), vp.Interface())

		// Attempt unwrapping a nested pointer
		if v.Elem().Kind() == reflect.Ptr {
			if tryAs(err, v.Elem()) {
				fmt.Printf("ret 0\n")
				return true
			}
		} else if tryAs(err, vp) {
			fmt.Printf("ret 1\n")
			v.Elem().Set(vp.Elem().Elem())
			return true
		} else if tryAs(err, v) {
			fmt.Printf("ret 2\n")
			return true
		}
			fmt.Printf("as 1: %s\n", vp.Type())
			if errors.As(err, vp.Interface()) {
				fmt.Printf("ret 1\n")
				return true
			}
		}

		if e.Kind() == reflect.Interface || e.Implements(errorType) {
			fmt.Printf("as 2: %s\n", v.Type())
			if errors.As(err, v.Interface()) {
				fmt.Printf("ret 2\n")
				return true
			}
		}
	}

	fmt.Printf("ret false\n")
	return false

	switch v.Kind() {
	case reflect.Struct:
	case reflect.Ptr:
		// Unwrap inner type
		vi := v.Elem()

		if _, ok := v.Interface().(error); ok {
			// Wrap target in an extra pointer layer
			// x := reflect.New(reflect.TypeOf(target))
			// if errors.As(err, x.Interface()) {
			// 	fmt.Printf("ret 3 (%v), err(%s): %v, target(%s): %v, x(%s): %v\n", true, reflect.TypeOf(err), err, reflect.TypeOf(target), target, reflect.TypeOf(x.Interface()), x.Interface())
			// 	return true
			// }
			x := reflect.New(v.Type())
			if errors.As(err, x.Interface()) {
				v.Elem().Set(x)
				fmt.Printf("ret 3 (%v), err(%s): %v, target(%s): %v, x(%s): %v\n", true, reflect.TypeOf(err), err, reflect.TypeOf(target), target, reflect.TypeOf(x.Interface()), x.Interface())
				return true
			}

			if _, ok := vi.Interface().(error); ok {
				res := errors.As(err, target)
				fmt.Printf("ret 4 (%v), err(%s): %v, target(%s): %v\n", res, reflect.TypeOf(err), err, reflect.TypeOf(target), target)
				return res
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
