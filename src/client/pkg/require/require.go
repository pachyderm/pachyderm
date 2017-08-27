package require

import (
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"testing"
)

// Matches checks that a string matches a regular-expression.
func Matches(tb testing.TB, expectedMatch string, actual string, msgAndArgs ...interface{}) {
	r, err := regexp.Compile(expectedMatch)
	if err != nil {
		fatal(tb, msgAndArgs, "Match string provided (%v) is invalid", expectedMatch)
	}
	if !r.MatchString(actual) {
		fatal(tb, msgAndArgs, "Actual string (%v) does not match pattern (%v)", actual, expectedMatch)
	}
}

// Equal checks equality of two values.
func Equal(tb testing.TB, expected interface{}, actual interface{}, msgAndArgs ...interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		fatal(
			tb,
			msgAndArgs,
			"Not equal: %#v (expected)\n"+
				"        != %#v (actual)", expected, actual)
	}
}

// NotEqual checks inequality of two values.
func NotEqual(tb testing.TB, expected interface{}, actual interface{}, msgAndArgs ...interface{}) {
	if reflect.DeepEqual(expected, actual) {
		fatal(
			tb,
			msgAndArgs,
			"Equal: %#v (expected)\n"+
				"    == %#v (actual)", expected, actual)
	}
}

// EqualOneOf checks if a value is equal to one of the elements of a slice.
func EqualOneOf(tb testing.TB, expecteds []interface{}, actual interface{}, msgAndArgs ...interface{}) {
	equal := false
	for _, expected := range expecteds {
		if reflect.DeepEqual(expected, actual) {
			equal = true
			break
		}
	}
	if !equal {
		fatal(
			tb,
			msgAndArgs,
			"Not equal 1 of: %#v (expecteds)\n"+
				"        != %#v (actual)", expecteds, actual)
	}
}

// oneOfEquals is a helper function for OneOfEquals and NoneEquals, that simply
// returns a bool indicating whether 'expected' is in the slice 'actuals'.
func oneOfEquals(expected interface{}, actuals interface{}) (bool, error) {
	e := reflect.ValueOf(expected)
	as := reflect.ValueOf(actuals)
	if as.Kind() != reflect.Slice {
		return false, fmt.Errorf("\"actuals\" must a be a slice, but instead was %s", as.Type().String())
	}
	if e.Type() != as.Type().Elem() {
		return false, nil
	}
	for i := 0; i < as.Len(); i++ {
		if reflect.DeepEqual(e.Interface(), as.Index(i).Interface()) {
			return true, nil
		}
	}
	return false, nil
}

// OneOfEquals checks whether one element of a slice equals a value.
func OneOfEquals(tb testing.TB, expected interface{}, actuals interface{}, msgAndArgs ...interface{}) {
	equal, err := oneOfEquals(expected, actuals)
	if err != nil {
		fatal(tb, msgAndArgs, err.Error())
	}
	if !equal {
		fatal(tb, msgAndArgs,
			"Not equal : %#v (expected)\n"+
				" one of  != %#v (actuals)", expected, actuals)
	}
}

// NoneEquals checks one element of a slice equals a value.
func NoneEquals(tb testing.TB, expected interface{}, actuals interface{}, msgAndArgs ...interface{}) {
	equal, err := oneOfEquals(expected, actuals)
	if err != nil {
		fatal(tb, msgAndArgs, err.Error())
	}
	if equal {
		fatal(tb, msgAndArgs,
			"Equal : %#v (expected)\n one of == %#v (actuals)", expected, actuals)
	}
}

// ElementsEqual checks whether the elements of the slice "expecteds" are
// exactly the elements of the slice "actuals", ignoring order (i.e.
// setwise-equal)
func ElementsEqual(tb testing.TB, expecteds interface{}, actuals interface{}, msgAndArgs ...interface{}) {
	if equal := func() bool {
		es := reflect.ValueOf(expecteds)
		as := reflect.ValueOf(actuals)
		if es.Kind() != reflect.Slice {
			fatal(tb, msgAndArgs, "ElementsEqual must be called with a slice, but \"expected\" was %s", es.Type().String())
			return false
		}
		if as.Kind() != reflect.Slice {
			fatal(tb, msgAndArgs, "ElementsEqual must be called with a slice, but \"actual\" was %s", as.Type().String())
			return false
		}
		if es.Type().Elem() != as.Type().Elem() {
			return false
		}
		expectedCt := reflect.MakeMap(reflect.MapOf(es.Type().Elem(), reflect.TypeOf(int(0))))
		actualCt := reflect.MakeMap(reflect.MapOf(as.Type().Elem(), reflect.TypeOf(int(0))))
		for i := 0; i < es.Len(); i++ {
			v := es.Index(i)
			if !expectedCt.MapIndex(v).IsValid() {
				expectedCt.SetMapIndex(v, reflect.ValueOf(1))
			} else {
				newCt := expectedCt.MapIndex(v).Int() + 1
				expectedCt.SetMapIndex(v, reflect.ValueOf(newCt))
			}
		}
		for i := 0; i < as.Len(); i++ {
			v := as.Index(i)
			if !actualCt.MapIndex(v).IsValid() {
				actualCt.SetMapIndex(v, reflect.ValueOf(1))
			} else {
				newCt := actualCt.MapIndex(v).Int() + 1
				actualCt.SetMapIndex(v, reflect.ValueOf(newCt))
			}
		}
		if expectedCt.Len() != actualCt.Len() {
			return false
		}
		for _, key := range expectedCt.MapKeys() {
			ec := expectedCt.MapIndex(key)
			ac := actualCt.MapIndex(key)
			if !ec.IsValid() || !ac.IsValid() ||
				!reflect.DeepEqual(ec.Interface(), ac.Interface()) {
				return false
			}
		}
		return true
	}(); !equal {
		fatal(tb, msgAndArgs,
			"Not equal: %#v (expecteds)\n"+
				"      != %#v (actual)", expecteds, actuals)
	}
}

// NoError checks for no error.
func NoError(tb testing.TB, err error, msgAndArgs ...interface{}) {
	if err != nil {
		fatal(tb, msgAndArgs, "No error is expected but got %v", err)
	}
}

// YesError checks for an error.
func YesError(tb testing.TB, err error, msgAndArgs ...interface{}) {
	if err == nil {
		fatal(tb, msgAndArgs, "Error is expected but got %v", err)
	}
}

// NotNil checks a value is non-nil.
func NotNil(tb testing.TB, object interface{}, msgAndArgs ...interface{}) {
	success := true

	if object == nil {
		success = false
	} else {
		value := reflect.ValueOf(object)
		kind := value.Kind()
		if kind >= reflect.Chan && kind <= reflect.Slice && value.IsNil() {
			success = false
		}
	}

	if !success {
		fatal(tb, msgAndArgs, "Expected value not to be nil.")
	}
}

// Nil checks a value is nil.
func Nil(tb testing.TB, object interface{}, msgAndArgs ...interface{}) {
	if object == nil {
		return
	}
	value := reflect.ValueOf(object)
	kind := value.Kind()
	if kind >= reflect.Chan && kind <= reflect.Slice && value.IsNil() {
		return
	}

	fatal(tb, msgAndArgs, "Expected value to be nil.")
}

// True checks a value is true.
func True(tb testing.TB, value bool, msgAndArgs ...interface{}) {
	if !value {
		fatal(tb, msgAndArgs, "Should be true.")
	}
}

// False checks a value is false.
func False(tb testing.TB, value bool, msgAndArgs ...interface{}) {
	if value {
		fatal(tb, msgAndArgs, "Should be false.")
	}
}

func logMessage(tb testing.TB, msgAndArgs []interface{}) {
	if len(msgAndArgs) == 1 {
		tb.Logf(msgAndArgs[0].(string))
	}
	if len(msgAndArgs) > 1 {
		tb.Logf(msgAndArgs[0].(string), msgAndArgs[1:]...)
	}
}

func fatal(tb testing.TB, userMsgAndArgs []interface{}, msgFmt string, msgArgs ...interface{}) {
	logMessage(tb, userMsgAndArgs)
	_, file, line, ok := runtime.Caller(2)
	if ok {
		tb.Logf("%s:%d", file, line)
	}
	tb.Fatalf(msgFmt, msgArgs...)
}
