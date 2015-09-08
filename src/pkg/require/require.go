package require

import (
	"reflect"
	"testing"
)

func logMessage(tb testing.TB, msgAndArgs ...interface{}) {
	if len(msgAndArgs) == 1 {
		tb.Logf(msgAndArgs[0].(string))
	}
	if len(msgAndArgs) > 1 {
		tb.Logf(msgAndArgs[0].(string), msgAndArgs[1:]...)
	}
}

func Equal(tb testing.TB, expected interface{}, actual interface{}, msgAndArgs ...interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		logMessage(tb, msgAndArgs...)
		tb.Errorf("Not equal: %#v (expected)\n"+
			"        != %#v (actual)", expected, actual)
	}
}

func NoError(tb testing.TB, err error, msgAndArgs ...interface{}) {
	if err != nil {
		logMessage(tb, msgAndArgs...)
		tb.Errorf("No error is expected but got %v", err)
	}
}

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
		logMessage(tb, msgAndArgs...)
		tb.Errorf("Expected value not to be nil.")
	}
}

func Nil(tb testing.TB, object interface{}, msgAndArgs ...interface{}) {
	if object == nil {
		return
	}
	value := reflect.ValueOf(object)
	kind := value.Kind()
	if kind >= reflect.Chan && kind <= reflect.Slice && value.IsNil() {
		return
	}

	logMessage(tb, msgAndArgs...)
	tb.Errorf("Expected value to be nil.")
}

func True(tb testing.TB, value bool, msgAndArgs ...interface{}) {
	if !value {
		logMessage(tb, msgAndArgs...)
		tb.Errorf("Should be true")
	}
}
