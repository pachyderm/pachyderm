package env

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
)

var (
	testEnvRestrictTo = []string{
		"REQUIRED_STRING",
		"OPTIONAL_STRING",
		"OPTIONAL_INT",
		"OPTIONAL_BOOL",
		"STRUCT_OPTIONAL_INT",
		"OPTIONAL_STRUCT",
		"OPTIONAL_STRUCT_TWO",
	}
	testEnvRestrictToWithoutOptionalBool = []string{
		"OPTIONAL_STRING",
		"OPTIONAL_INT",
		"REQUIRED_STRING",
		"STRUCT_OPTIONAL_INT",
		"OPTIONAL_STRUCT",
		"OPTIONAL_STRUCT_TWO",
	}
)

type testEnv struct {
	RequiredString string `env:"REQUIRED_STRING,required"`
	OptionalString string `env:"OPTIONAL_STRING"`
	OptionalInt    int    `env:"OPTIONAL_INT"`
	OptionalBool   bool   `env:"OPTIONAL_BOOL"`
	OptionalStruct struct {
		StructOptionalInt int `env:"STRUCT_OPTIONAL_INT"`
	} `env:"OPTIONAL_STRUCT,match=FOO"`
	OptionalStructTwo struct {
		StructOptionalInt int `env:"STRUCT_OPTIONAL_INT"`
	} `env:"OPTIONAL_STRUCT_TWO,match=^(FOO|BAR)$"`
}

// TODO(pedge): if tests are run in parallel, this is affecting global state

func TestBasic(t *testing.T) {
	testState := newTestState(testEnvRestrictTo)
	defer testState.reset()
	testSetenv(map[string]string{
		"REQUIRED_STRING": "foo",
		"OPTIONAL_STRING": "",
		"OPTIONAL_INT":    "1234",
	})
	testEnv := populateTestEnv(t)
	checkEqual(t, "", testEnv.OptionalString)
	checkEqual(t, 1234, testEnv.OptionalInt)
	checkEqual(t, "foo", testEnv.RequiredString)
}

func TestBasicWithDefault(t *testing.T) {
	testState := newTestState(testEnvRestrictTo)
	defer testState.reset()
	testSetenv(map[string]string{
		"REQUIRED_STRING": "foo",
		"OPTIONAL_STRING": "",
	})
	testEnv := populateTestEnvLong(t, testEnvRestrictTo, nil, map[string]string{"OPTIONAL_INT": "1234"})
	checkEqual(t, "", testEnv.OptionalString)
	checkEqual(t, 1234, testEnv.OptionalInt)
	checkEqual(t, "foo", testEnv.RequiredString)
}

func TestBasicWithDefaultRequiredMissing(t *testing.T) {
	testState := newTestState(testEnvRestrictTo)
	defer testState.reset()
	testSetenv(map[string]string{
		"OPTIONAL_STRING": "",
	})
	populateTestEnvExpectErrorLong(t, envKeyNotSetWhenRequiredErr, testEnvRestrictTo, nil, map[string]string{"OPTIONAL_INT": "1234"})
}

func TestMissing(t *testing.T) {
	testState := newTestState(testEnvRestrictTo)
	defer testState.reset()
	testSetenv(map[string]string{
		"REQUIRED_STRING": "",
	})
	populateTestEnvExpectError(t, envKeyNotSetWhenRequiredErr)
}

func TestOutsideOfRestrictToRange(t *testing.T) {
	testState := newTestState(testEnvRestrictToWithoutOptionalBool)
	defer testState.reset()
	testSetenv(map[string]string{
		"REQUIRED_STRING": "foo",
		"OPTIONAL_BOOL":   "1",
	})
	populateTestEnvExpectErrorLong(t, invalidTagRestrictToErr, testEnvRestrictToWithoutOptionalBool, nil, nil)
}

func TestCannotParse(t *testing.T) {
	testState := newTestState(testEnvRestrictTo)
	defer testState.reset()
	testSetenv(map[string]string{
		"REQUIRED_STRING": "foo",
		"OPTIONAL_INT":    "abc",
	})
	populateTestEnvExpectError(t, cannotParseErr)
}

func TestParsingBool(t *testing.T) {
	testState := newTestState(testEnvRestrictTo)
	defer testState.reset()
	testSetenv(map[string]string{
		"REQUIRED_STRING": "foo",
		"OPTIONAL_BOOL":   "1",
	})
	testEnv := populateTestEnv(t)
	checkEqual(t, true, testEnv.OptionalBool)
	testSetenv(map[string]string{
		"REQUIRED_STRING": "foo",
		"OPTIONAL_BOOL":   "",
	})
	testEnv = populateTestEnv(t)
	checkEqual(t, false, testEnv.OptionalBool)
	testSetenv(map[string]string{
		"REQUIRED_STRING": "foo",
		"OPTIONAL_BOOL":   "false",
	})
	testEnv = populateTestEnv(t)
	checkEqual(t, false, testEnv.OptionalBool)
}

func TestOptionalStruct(t *testing.T) {
	testState := newTestState(testEnvRestrictTo)
	defer testState.reset()
	testSetenv(map[string]string{
		"REQUIRED_STRING":     "foo",
		"OPTIONAL_STRUCT":     "FOO",
		"STRUCT_OPTIONAL_INT": "1234",
	})
	testEnv := populateTestEnv(t)
	checkEqual(t, 1234, testEnv.OptionalStruct.StructOptionalInt)
	testSetenv(map[string]string{
		"REQUIRED_STRING":     "foo",
		"OPTIONAL_STRUCT":     "",
		"STRUCT_OPTIONAL_INT": "1234",
	})
	testEnv = populateTestEnv(t)
	checkEqual(t, 0, testEnv.OptionalStruct.StructOptionalInt)
	testSetenv(map[string]string{
		"REQUIRED_STRING":     "foo",
		"OPTIONAL_STRUCT":     "BAR",
		"STRUCT_OPTIONAL_INT": "1234",
	})
	testEnv = populateTestEnv(t)
	checkEqual(t, 0, testEnv.OptionalStruct.StructOptionalInt)
	testSetenv(map[string]string{
		"REQUIRED_STRING":     "foo",
		"OPTIONAL_STRUCT":     "FOOO",
		"STRUCT_OPTIONAL_INT": "1234",
	})
	testEnv = populateTestEnv(t)
	checkEqual(t, 1234, testEnv.OptionalStruct.StructOptionalInt)
	testSetenv(map[string]string{
		"REQUIRED_STRING":     "foo",
		"OPTIONAL_STRUCT":     "FOO$",
		"STRUCT_OPTIONAL_INT": "1234",
	})
	testEnv = populateTestEnv(t)
	checkEqual(t, 1234, testEnv.OptionalStruct.StructOptionalInt)
	testEnv = populateTestEnv(t)
	checkEqual(t, 1234, testEnv.OptionalStruct.StructOptionalInt)
	testSetenv(map[string]string{
		"REQUIRED_STRING":     "foo",
		"OPTIONAL_STRUCT_TWO": "BAR",
		"STRUCT_OPTIONAL_INT": "1234",
	})
	testEnv = populateTestEnv(t)
	checkEqual(t, 1234, testEnv.OptionalStructTwo.StructOptionalInt)
}

func TestEnvFileDecoderBasic(t *testing.T) {
	reader := getTestReader(t, "_testdata/env.env")
	m, err := newEnvFileDecoder(reader).Decode()
	if err != nil {
		t.Fatal(err)
	}
	checkEqual(t, "bar", m["FOO"])
	checkEqual(t, "baz", m["BAR"])
	checkEqual(t, "", m["BAZ"])
	checkEqual(t, "9", m["BAT"])
	checkEqual(t, "false", m["BAN"])

}

func TestJSONDecoderBasic(t *testing.T) {
	reader := getTestReader(t, "_testdata/env.json")
	m, err := newJSONDecoder(reader).Decode()
	if err != nil {
		t.Fatal(err)
	}
	checkEqual(t, "bar", m["FOO"])
	checkEqual(t, "baz", m["BAR"])
	checkEqual(t, "", m["BAZ"])
	checkEqual(t, "10", m["BAT"])
	checkEqual(t, "false", m["BAN"])
}

func TestPopulateDecoders(t *testing.T) {
	testState := newTestState(testEnvRestrictTo)
	defer testState.reset()
	decoders := []Decoder{
		newEnvFileDecoder(getTestReader(t, "_testdata/env.env")),
		newJSONDecoder(getTestReader(t, "_testdata/env.json")),
	}
	testSetenv(map[string]string{
		"REQUIRED_STRING": "foo",
	})
	testEnv := populateTestEnvLong(t, testEnvRestrictTo, decoders, nil)
	checkEqual(t, "BAZ", testEnv.RequiredString)
	checkEqual(t, "BAR", testEnv.OptionalString)
}

type testState struct {
	originalEnv map[string]string
}

func newTestState(restrictTo []string) *testState {
	originalEnv := make(map[string]string)
	for _, elem := range restrictTo {
		originalEnv[elem] = os.Getenv(elem)
	}
	return &testState{
		originalEnv,
	}
}

func (t *testState) reset() {
	testSetenv(t.originalEnv)
}

func testSetenv(env map[string]string) {
	for key, value := range env {
		_ = os.Setenv(key, value)
	}
}

func populateTestEnv(t *testing.T) *testEnv {
	return populateTestEnvLong(t, testEnvRestrictTo, nil, nil)
}

func populateTestEnvLong(t *testing.T, restrictTo []string, decoders []Decoder, defaults map[string]string) *testEnv {
	testEnv := &testEnv{}
	if err := Populate(
		testEnv,
		PopulateOptions{
			RestrictTo: restrictTo,
			Decoders:   decoders,
			Defaults:   defaults,
		},
	); err != nil {
		t.Error(err)
	}
	return testEnv
}

func populateTestEnvExpectError(t *testing.T, expected string) {
	populateTestEnvExpectErrorLong(t, expected, testEnvRestrictTo, nil, nil)
}

func populateTestEnvExpectErrorLong(t *testing.T, expected string, restrictTo []string, decoders []Decoder, defaults map[string]string) {
	testEnv := &testEnv{}
	err := Populate(
		testEnv,
		PopulateOptions{
			RestrictTo: restrictTo,
			Decoders:   decoders,
			Defaults:   defaults,
		},
	)
	if err == nil {
		t.Error("expected error")
	} else if !strings.HasPrefix(err.Error(), expected) {
		t.Errorf("expected error type %s, got error %s", expected, err.Error())
	}
}

func getTestReader(t *testing.T, filePath string) io.Reader {
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatal(err)
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		if err := file.Close(); err != nil {
			t.Error(err)
		}
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	return bytes.NewBuffer(data)
}

func checkEqual(t *testing.T, expected interface{}, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}
