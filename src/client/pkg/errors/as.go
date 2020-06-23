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
		vp := reflect.New(v.Type())

		// Attempt unwrapping a nested pointer
		if v.Elem().Kind() == reflect.Ptr && tryAs(err, v.Elem()) {
			fmt.Printf("ret 0\n")
			return true
		}

		// Attempt wrapping with an additional pointer
		if tryAs(err, vp) {
			fmt.Printf("ret 1\n")
			v.Elem().Set(vp.Elem().Elem())
			return true
		}

		// Attempt the passed target as-is
		if tryAs(err, v) {
			fmt.Printf("ret 2\n")
			return true
		}
	}
	return false
}
