package require

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"runtime/debug"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/testing/protocmp"
)

// Matches checks that a string matches a regular-expression.
func Matches(tb testing.TB, expectedMatch string, actual string, msgAndArgs ...interface{}) {
	tb.Helper()
	r, err := regexp.Compile(expectedMatch)
	if err != nil {
		fatal(tb, msgAndArgs, "Match string provided (%v) is invalid", expectedMatch)
	}
	if !r.MatchString(actual) {
		fatal(tb, msgAndArgs, "Actual string (%v) does not match pattern (%v)", actual, expectedMatch)
	}
}

// OneOfMatches checks whether one element of a slice matches a regular-expression.
func OneOfMatches(tb testing.TB, expectedMatch string, actuals []string, msgAndArgs ...interface{}) {
	tb.Helper()
	r, err := regexp.Compile(expectedMatch)
	if err != nil {
		fatal(tb, msgAndArgs, "Match string provided (%v) is invalid", expectedMatch)
	}
	for _, actual := range actuals {
		if r.MatchString(actual) {
			return
		}
	}
	fatal(tb, msgAndArgs, "None of actual strings (%v) match pattern (%v)", actuals, expectedMatch)

}

// NoneMatch checks whether any element of a slice matches a regular-expression
// and returns an error if a match is found.
func NoneMatch(tb testing.TB, shouldNotMatch string, actuals []string, msgAndArgs ...interface{}) {
	tb.Helper()
	r, err := regexp.Compile(shouldNotMatch)
	if err != nil {
		fatal(tb, msgAndArgs, "Match string provided (%v) is invalid", shouldNotMatch)
	}
	for _, actual := range actuals {
		if r.MatchString(actual) {
			fatal(tb, msgAndArgs, "string (%v) should not match pattern (%v)", actual, shouldNotMatch)
		}
	}
}

// NotMatch checks whether actual matches a regular-expression and returns an error if a match is found.
func NotMatch(tb testing.TB, shouldNotMatch string, actual string, msgAndArgs ...interface{}) {
	tb.Helper()
	r, err := regexp.Compile(shouldNotMatch)
	if err != nil {
		fatal(tb, msgAndArgs, "Match string provided (%v) is invalid", shouldNotMatch)
	}
	if r.MatchString(actual) {
		fatal(tb, msgAndArgs, "string (%v) should not match pattern (%v)", actual, shouldNotMatch)
	}
}

// Equal checks the equality of two values
func Equal(tb testing.TB, expected interface{}, actual interface{}, msgAndArgs ...interface{}) {
	tb.Helper()
	if err := EqualOrErr(expected, actual); err != nil {
		fatal(tb, msgAndArgs, err.Error())
	}
}

// EqualOrErr checks equality of two values and returns an error if they're not equal
func EqualOrErr(expected interface{}, actual interface{}) error {
	if diff := cmp.Diff(expected, actual, protocmp.Transform(), cmpopts.EquateErrors()); diff == "" {
		return nil
	}

	eV, aV := reflect.ValueOf(expected), reflect.ValueOf(actual)
	if eV.Type() != aV.Type() {
		return errors.Errorf("Not equal: %T(%#v) (expected)\n"+
			"        != %T(%#v) (actual)", expected, expected, actual, actual)
	}
	if !reflect.DeepEqual(expected, actual) {
		return errors.Errorf(
			"Not equal: %#v (expected)\n"+
				"        != %#v (actual)", expected, actual)
	}
	return nil
}

// NotEqual checks inequality of two values.
func NotEqual(tb testing.TB, expected interface{}, actual interface{}, msgAndArgs ...interface{}) {
	tb.Helper()
	if reflect.DeepEqual(expected, actual) {
		fatal(
			tb,
			msgAndArgs,
			"Equal: (want non-equal)\n"+
				"    %#v (expected)\n"+
				" == %#v (actual)", expected, actual)
	}
}

// ElementsEqualOrErr returns nil if the elements of the slice "expecteds" are
// exactly the elements of the slice "actuals", ignoring order (i.e.
// setwise-equal), and an error otherwise.
//
// Unlike other require.* functions, this returns an error, so that if the
// caller is polling e.g. ListCommit or ListAdmins, they can wrap
// ElementsEqualOrErr in a retry loop.
//
// Also, like ElementsEqual, treat 'nil' and the empty slice as equivalent (for
// convenience)
func ElementsEqualOrErr(expecteds interface{}, actuals interface{}) error {
	es := reflect.ValueOf(expecteds)
	as := reflect.ValueOf(actuals)

	// If either slice is empty, check that both are empty
	esIsEmpty := expecteds == nil || es.IsNil() || (es.Kind() == reflect.Slice && es.Len() == 0)
	asIsEmpty := actuals == nil || as.IsNil() || (as.Kind() == reflect.Slice && as.Len() == 0)
	if esIsEmpty && asIsEmpty {
		return nil
	} else if esIsEmpty {
		return errors.Errorf("expected 0 elements, but got %d: %v", as.Len(), actuals)
	} else if asIsEmpty {
		return errors.Errorf("expected %d elements, but got 0\n  expected: %v", es.Len(), expecteds)
	}

	// Both slices are nonempty--compare elements
	if es.Kind() != reflect.Slice {
		return errors.Errorf("\"expecteds\" must be a slice, but was %s", es.Type().String())
	}
	if as.Kind() != reflect.Slice {
		return errors.Errorf("\"actuals\" must be a slice, but was %s", as.Type().String())
	}

	// Make sure expecteds and actuals are slices of the same type, modulo
	// pointers (*T ~= T in this function)
	esArePtrs := es.Type().Elem().Kind() == reflect.Ptr
	asArePtrs := as.Type().Elem().Kind() == reflect.Ptr
	esElemType, asElemType := es.Type().Elem(), as.Type().Elem()
	if esArePtrs {
		esElemType = es.Type().Elem().Elem()
	}
	if asArePtrs {
		asElemType = as.Type().Elem().Elem()
	}
	if esElemType != asElemType {
		return errors.Errorf("expected []%s but got []%s", es.Type().Elem(), as.Type().Elem())
	}

	if es.Len() != as.Len() {
		// slight abuse of error: contains newlines so final output prints well
		return errors.Errorf("expected %d elements, but got %d\n  expected: %v\n  actual: %v",
			es.Len(), as.Len(), expecteds, actuals)
	}

	// Count up elements of expecteds
	intType := reflect.TypeOf(int64(0))
	expectedCt := reflect.MakeMap(reflect.MapOf(esElemType, intType))
	for i := 0; i < es.Len(); i++ {
		v := es.Index(i)
		if esArePtrs {
			v = v.Elem()
		}
		if !expectedCt.MapIndex(v).IsValid() {
			expectedCt.SetMapIndex(v, reflect.ValueOf(int64(1)))
		} else {
			newCt := expectedCt.MapIndex(v).Int() + 1
			expectedCt.SetMapIndex(v, reflect.ValueOf(newCt))
		}
	}

	// Count up elements of actuals
	actualCt := reflect.MakeMap(reflect.MapOf(asElemType, intType))
	for i := 0; i < as.Len(); i++ {
		v := as.Index(i)
		if asArePtrs {
			v = v.Elem()
		}
		if !actualCt.MapIndex(v).IsValid() {
			actualCt.SetMapIndex(v, reflect.ValueOf(int64(1)))
		} else {
			newCt := actualCt.MapIndex(v).Int() + 1
			actualCt.SetMapIndex(v, reflect.ValueOf(newCt))
		}
	}
	for _, key := range expectedCt.MapKeys() {
		ec := expectedCt.MapIndex(key)
		ac := actualCt.MapIndex(key)
		if !ec.IsValid() || !ac.IsValid() || ec.Int() != ac.Int() {
			ecInt, acInt := int64(0), int64(0)
			if ec.IsValid() {
				ecInt = ec.Int()
			}
			if ac.IsValid() {
				acInt = ac.Int()
			}
			// slight abuse of error: contains newlines so final output prints well
			return errors.Errorf("expected %d copies of %v, but got %d copies\n  expected: %v\n  actual: %v", ecInt, key, acInt, expecteds, actuals)
		}
	}
	return nil
}

// ImagesEqual is similar to 'ElementsEqualUnderFn', but it applies 'f' to both
// 'expecteds' and 'actuals'. This is useful for doing before/after
// comparisons. This can also compare 'T' and '*T' , but 'f' should expect
// interfaces wrapping the underlying types of 'expecteds' (e.g. if 'expecteds'
// has type []*T and 'actuals' has type []T, then 'f' should cast its argument
// to '*T')
//
// Like ElementsEqual, treat 'nil' and the empty slice as equivalent (for
// convenience)
func ImagesEqual(tb testing.TB, expecteds interface{}, actuals interface{}, f func(interface{}) interface{}, msgAndArgs ...interface{}) {
	tb.Helper()
	as := reflect.ValueOf(actuals)
	es := reflect.ValueOf(expecteds)

	// Check if 'actuals' is empty; if so, just pass nil (no need to transform)
	if actuals != nil && !as.IsNil() && as.Kind() != reflect.Slice {
		fatal(tb, msgAndArgs, fmt.Sprintf("\"actuals\" must be a slice, but was %s", as.Type().String()))
	} else if actuals == nil || as.IsNil() || as.Len() == 0 {
		// Just pass 'nil' for 'actuals'
		if err := ElementsEqualOrErr(expecteds, nil); err != nil {
			fatal(tb, msgAndArgs, err.Error())
		}
		return
	}

	// Check if 'expecteds' is empty: if so, return an error (since 'actuals' is
	// not empty)
	if expecteds != nil && !es.IsNil() && es.Kind() != reflect.Slice {
		fatal(tb, msgAndArgs, fmt.Sprintf("\"expecteds\" must be a slice, but was %s", as.Type().String()))
	} else if expecteds == nil || es.IsNil() || es.Len() == 0 {
		fatal(tb, msgAndArgs, fmt.Sprintf("expected 0 distinct elements, but got %d\n elements (before function is applied): %v", as.Len(), actuals))
	}

	// Make sure expecteds and actuals are slices of the same type, modulo
	// pointers (*T ~= T in this function). This is better than some kind of
	// opaque reflection error from calling 'f' on mismatched types.
	esArePtrs := es.Type().Elem().Kind() == reflect.Ptr
	asArePtrs := as.Type().Elem().Kind() == reflect.Ptr
	esUnderlyingType, asUnderlyingType := es.Type().Elem(), as.Type().Elem()
	if esArePtrs {
		esUnderlyingType = es.Type().Elem().Elem()
	}
	if asArePtrs {
		asUnderlyingType = as.Type().Elem().Elem()
	}
	if esUnderlyingType != asUnderlyingType {
		fatal(tb, msgAndArgs, "expected []%s but got []%s", esUnderlyingType, asUnderlyingType)
	}

	if es.Len() != as.Len() {
		// slight abuse of error: contains newlines so final output prints well
		fatal(tb, msgAndArgs, "expected %d elements, but got %d\n  expected: %v\n  actual: %v",
			es.Len(), as.Len(), expecteds, actuals)
	}

	// apply 'f' to both 'es' and 'as'. Make 'es[i]' have the same underlying
	// type as 'as[i]' (may need to take address or dereference elements) so 'f'
	// can apply to both.
	newExpecteds, newActuals := make([]interface{}, 0, as.Len()), make([]interface{}, 0, as.Len())
	for i := 0; i < es.Len(); i++ {
		switch {
		case asArePtrs && !esArePtrs:
			newExpecteds = append(newExpecteds, f(es.Index(i).Addr().Interface()))
		case !asArePtrs && esArePtrs:
			newExpecteds = append(newExpecteds, f(es.Index(i).Elem().Interface()))
		default:
			newExpecteds = append(newExpecteds, f(es.Index(i).Interface()))
		}
	}
	for i := 0; i < as.Len(); i++ {
		newActuals = append(newActuals, f(as.Index(i).Interface()))
	}
	if err := ElementsEqualOrErr(newExpecteds, newActuals); err != nil {
		fatal(tb, msgAndArgs, err.Error())
	}
}

// ElementsEqualUnderFn checks that the elements of the slice 'expecteds' are
// the same as the elements of the slice 'actuals' under 'f', ignoring order
// (i.e.  'expecteds' and 'map(f, actuals)' are setwise-equal, but respecting
// duplicates). This is useful for cases where ElementsEqual doesn't quite work,
// e.g. because the type in 'expecteds'/'actuals' contains a pointer, or
// 'actuals' contains superfluous data which you wish to discard.
//
// Like ElementsEqual, treat 'nil' and the empty slice as equivalent (for
// convenience)
func ElementsEqualUnderFn(tb testing.TB, expecteds interface{}, actuals interface{}, f func(interface{}) interface{}, msgAndArgs ...interface{}) {
	tb.Helper()
	as := reflect.ValueOf(actuals)
	es := reflect.ValueOf(expecteds)

	// Check if 'actuals' is empty; if so, just pass nil (no need to transform)
	if actuals != nil && !as.IsNil() && as.Kind() != reflect.Slice {
		fatal(tb, msgAndArgs, fmt.Sprintf("\"actuals\" must be a slice, but was %s", as.Type().String()))
	} else if actuals == nil || as.IsNil() || as.Len() == 0 {
		// Just pass 'nil' for 'actuals'
		if err := ElementsEqualOrErr(expecteds, nil); err != nil {
			fatal(tb, msgAndArgs, err.Error())
		}
		return
	}

	// Check if 'expecteds' is empty: if so, return an error (since 'actuals' is
	// not empty)
	if expecteds != nil && !es.IsNil() && es.Kind() != reflect.Slice {
		fatal(tb, msgAndArgs, fmt.Sprintf("\"expecteds\" must be a slice, but was %s", as.Type().String()))
	} else if expecteds == nil || es.IsNil() || es.Len() == 0 {
		fatal(tb, msgAndArgs, fmt.Sprintf("expected 0 distinct elements, but got %d\n elements (before function is applied): %v", as.Len(), actuals))
	}

	// Neither 'expecteds' nor 'actuals' is empty--apply 'f' to 'actuals'
	newActuals := reflect.MakeSlice(reflect.SliceOf(es.Type().Elem()), as.Len(), as.Len())
	for i := 0; i < as.Len(); i++ {
		newActuals.Index(i).Set(reflect.ValueOf(f(as.Index(i).Interface())))
	}
	if err := ElementsEqualOrErr(expecteds, newActuals.Interface()); err != nil {
		fatal(tb, msgAndArgs, err.Error())
	}
}

// ElementsEqual checks that the elements of the slice "expecteds" are
// exactly the elements of the slice "actuals", ignoring order (i.e.
// setwise-equal, but respecting duplicates).
//
// Note that if the elements of 'expecteds' and 'actuals' are pointers,
// ElementsEqual will unwrap the pointers before comparing them, so that the
// output of e.g. ListCommit(), which returns []*pfs.Commit can easily be
// verfied.
//
// Also, treat 'nil' and the empty slice as equivalent, so that callers can
// pass 'nil' for 'expecteds'.
func ElementsEqual(tb testing.TB, expecteds interface{}, actuals interface{}, msgAndArgs ...interface{}) {
	tb.Helper()
	if err := ElementsEqualOrErr(expecteds, actuals); err != nil {
		fatal(tb, msgAndArgs, err.Error())
	}
}

// oneOfEquals is a helper function for EqualOneOf, OneOfEquals and NoneEquals, that simply
// returns a bool indicating whether 'elem' is in 'slice'. 'sliceName' is used for errors
func oneOfEquals(sliceName string, slice interface{}, elem interface{}) (bool, error) {
	e := reflect.ValueOf(elem)
	sl := reflect.ValueOf(slice)
	if slice == nil || sl.IsNil() {
		sl = reflect.MakeSlice(reflect.SliceOf(e.Type()), 0, 0)
	}
	if sl.Kind() != reflect.Slice {
		return false, errors.Errorf("\"%s\" must a be a slice, but instead was %s", sliceName, sl.Type().String())
	}
	if e.Type() != sl.Type().Elem() {
		return false, nil
	}
	arePtrs := e.Kind() == reflect.Ptr
	for i := 0; i < sl.Len(); i++ {
		if !arePtrs && reflect.DeepEqual(e.Interface(), sl.Index(i).Interface()) {
			return true, nil
		} else if arePtrs && reflect.DeepEqual(e.Elem().Interface(), sl.Index(i).Elem().Interface()) {
			return true, nil
		}
	}
	return false, nil
}

// EqualOneOf checks if a value is equal to one of the elements of a slice. Note
// that if expecteds and actual are a slice of pointers and a pointer
// respectively, then the pointers are unwrapped before comparison (so this
// functions works for e.g. *pfs.Commit and []*pfs.Commit)
func EqualOneOf(tb testing.TB, expecteds interface{}, actual interface{}, msgAndArgs ...interface{}) {
	tb.Helper()
	equal, err := oneOfEquals("expecteds", expecteds, actual)
	if err != nil {
		fatal(tb, msgAndArgs, err.Error())
	}
	if !equal {
		fatal(
			tb,
			msgAndArgs,
			"None of : %#v (expecteds)\n"+
				"              == %#v (actual)", expecteds, actual)
	}
}

// OneOfEquals checks whether one element of a slice equals a value. Like
// EqualsOneOf, OneOfEquals unwraps pointers
func OneOfEquals(tb testing.TB, expected interface{}, actuals interface{}, msgAndArgs ...interface{}) {
	tb.Helper()
	equal, err := oneOfEquals("actuals", actuals, expected)
	if err != nil {
		fatal(tb, msgAndArgs, err.Error())
	}
	if !equal {
		fatal(tb, msgAndArgs,
			"Not equal : %#v (expected)\n"+
				" one of  != %#v (actuals)", expected, actuals)
	}
}

// NoneEquals checks one element of a slice equals a value. Like
// EqualsOneOf, NoneEquals unwraps pointers.
func NoneEquals(tb testing.TB, expected interface{}, actuals interface{}, msgAndArgs ...interface{}) {
	tb.Helper()
	equal, err := oneOfEquals("actuals", actuals, expected)
	if err != nil {
		fatal(tb, msgAndArgs, err.Error())
	}
	if equal {
		fatal(tb, msgAndArgs,
			"Equal : %#v (expected)\n == one of %#v (actuals)", expected, actuals)
	}
}

// NoError checks for no error.
func NoError(tb testing.TB, err error, msgAndArgs ...interface{}) {
	tb.Helper()
	if err != nil {
		fatal(tb, msgAndArgs, "No error is expected but got %v", err)
	}
}

// NoErrorWithinT checks that 'f' finishes within time 't' and does not emit an
// error
func NoErrorWithinT(tb testing.TB, t time.Duration, f func() error, msgAndArgs ...interface{}) {
	tb.Helper()
	errCh := make(chan error)
	go func() {
		// This goro will leak if the timeout is exceeded, but it's okay because the
		// test is failing anyway
		errCh <- f()
	}()
	select {
	case err := <-errCh:
		if err != nil {
			fatal(tb, msgAndArgs, "No error is expected but got %v", err)
		}
	case <-time.After(t):
		fatal(tb, msgAndArgs, "operation did not finish within %s", t.String())
	}
}

// NoErrorWithinTRetry checks that 'f' finishes within time 't' and does not
// emit an error. Unlike NoErrorWithinT if f does error, it will retry it. Uses an exponential backoff.
func NoErrorWithinTRetry(tb testing.TB, t time.Duration, f func() error, msgAndArgs ...interface{}) {
	tb.Helper()
	noErrorWithinTRetry(tb, t, f, backoff.NewExponentialBackOff(), msgAndArgs...)
}

// NoErrorWithinTRetryConstant checks that 'f' finishes within time 't' and does not
// emit an error. Will retry at a constant specified interval rather than using exponential backoff
func NoErrorWithinTRetryConstant(tb testing.TB, t time.Duration, f func() error, interval time.Duration, msgAndArgs ...interface{}) {
	tb.Helper()
	noErrorWithinTRetry(tb, t, f, backoff.NewConstantBackOff(interval), msgAndArgs...)
}

// Same as NoErrorWithinTRetry but accepts a custom backoff
func noErrorWithinTRetry(tb testing.TB, t time.Duration, f func() error, bo backoff.BackOff, msgAndArgs ...interface{}) {
	tb.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), t)
	defer cancel()
	var userError error
	err := backoff.RetryUntilCancel(ctx, func() error {
		userError = f()
		if userError != nil {
			tb.Logf("retryable: %v", userError)
		}
		return userError
	}, bo, nil)
	if err != nil {
		fatal(tb, msgAndArgs, "operation did not finish within %s - last error: %v", t.String(), userError)
	}
}

// YesError checks for an error.
func YesError(tb testing.TB, err error, msgAndArgs ...interface{}) {
	tb.Helper()
	if err == nil {
		fatal(tb, msgAndArgs, "Error is expected but got %v", err)
	}
}

// ErrorIs checks that the errors.Is returns true
func ErrorIs(tb testing.TB, err, target error, msgAndArgs ...interface{}) {
	tb.Helper()
	if !errors.Is(err, target) {
		fatal(tb, msgAndArgs, "errors.Is(%v, %v) should be true", err, target)
	}
}

// ErrorContains checks that the error contains a specified substring.
func ErrorContains(tb testing.TB, err error, contains string, msgAndArgs ...interface{}) {
	require.ErrorContains(tb, err, contains, msgAndArgs...)
}

// NotNil checks a value is non-nil.
func NotNil(tb testing.TB, object interface{}, msgAndArgs ...interface{}) {
	tb.Helper()
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
	tb.Helper()
	if object == nil {
		return
	}
	value := reflect.ValueOf(object)
	kind := value.Kind()
	if kind >= reflect.Chan && kind <= reflect.Slice && value.IsNil() {
		return
	}

	fatal(tb, msgAndArgs, "Expected value to be nil, but was %v", object)
}

// True checks a value is true.
func True(tb testing.TB, value bool, msgAndArgs ...interface{}) {
	tb.Helper()
	if !value {
		fatal(tb, msgAndArgs, "Should be true.")
	}
}

// False checks a value is false.
func False(tb testing.TB, value bool, msgAndArgs ...interface{}) {
	tb.Helper()
	if value {
		fatal(tb, msgAndArgs, "Should be false.")
	}
}

// YesPanic checks that the callback panics.
func YesPanic(tb testing.TB, cb func(), msgAndArgs ...interface{}) {
	defer func() {
		r := recover()
		if r == nil {
			fatal(tb, msgAndArgs, "Should have panicked.")
		}
	}()

	cb()
}

// Len asserts the the provided object x has length l ( len(x) == l )
func Len(tb testing.TB, x interface{}, l int, msgAndArgs ...interface{}) {
	require.Len(tb, x, l, msgAndArgs...)
}

func logMessage(tb testing.TB, msgAndArgs []interface{}) {
	tb.Helper()
	if len(msgAndArgs) == 1 {
		tb.Logf(msgAndArgs[0].(string))
	}
	if len(msgAndArgs) > 1 {
		tb.Logf(msgAndArgs[0].(string), msgAndArgs[1:]...)
	}
}

func fatal(tb testing.TB, userMsgAndArgs []interface{}, msgFmt string, msgArgs ...interface{}) {
	tb.Helper()
	logMessage(tb, userMsgAndArgs)
	tb.Logf(msgFmt, msgArgs...)
	if len(msgArgs) > 0 {
		err, ok := msgArgs[0].(error)
		if ok {
			errors.ForEachStackFrame(err, func(frame errors.Frame) {
				tb.Logf("%+v\n", frame)
			})
		}
	}
	tb.Fatalf("current stack:\n%s", string(debug.Stack()))
}
