package protoextensions

import (
	"testing"

	"go.uber.org/zap/zapcore"
)

type testEncoder struct {
	zapcore.ObjectEncoder
	string string
}

func (e *testEncoder) AddString(key, value string) {
	e.string = value
}

func TestHalf(t *testing.T) {
	testData := []struct {
		input, want string
	}{
		{
			input: "",
			want:  "",
		},
		{
			input: "1",
			want:  ".../1",
		},
		{
			input: "12",
			want:  "1.../2",
		},
		{
			input: "abccba",
			want:  "abc.../6",
		},
		{
			input: "abcdcba",
			want:  "abc.../7",
		},
	}

	for _, test := range testData {
		t.Run(test.input, func(t *testing.T) {
			e := new(testEncoder)
			AddHalfString(e, "", test.input)
			if got, want := e.string, test.want; got != want {
				t.Errorf("half:\n  got: %v\n want: %v", got, want)
			}
		})
	}
}
