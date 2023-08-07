package server

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"google.golang.org/protobuf/proto"
)

func TestJSONMergePatch(t *testing.T) {
	var testCases = []struct{ target, patch, result string }{
		// first example in RFC 7396
		{`{"a": "b", "c": {"d": "e", "f": "g"}}`, `{"a":"z", "c": {"f": null}}`, `{"a": "z", "c": {"d": "e"}}`},

		// second example in RFC 7396
		{
			`{"title": "Goodbye!", "author": {"givenName": "John", "familyName": "Doe"}, "tags":["example", "sample"], "content": "This will be unchanged"}`, ` {"title": "Hello!", "phoneNumber": "+01-123-456-7890", "author": {"familyName": null}, "tags": ["example"]}`, `{"title": "Hello!", "author": {"givenName": "John"}, "tags": ["example"], "content": "This will be unchanged", "phoneNumber": "+01-123-456-7890"}`,
		},

		// test cases from Appendix A of RFC 7396
		{`{"a":"b"}`, `{"a":"c"}`, `{"a":"c"}`},

		{`{"a":"b"}`, `{"b":"c"}`, `{"a":"b", "b":"c"}`},

		{`{"a":"b"}`, `{"a":null}`, `{}`},

		{`{"a":"b","b":"c"}`, `{"a":null}`, `{"b":"c"}`},

		{`{"a":["b"]}`, `{"a":"c"}`, `{"a":"c"}`},

		{`{"a":"c"}`, `{"a":["b"]}`, `{"a":["b"]}`},

		{`{"a": {"b": "c"}}`, `{"a": {  "b": "d","c": null}}`, `{"a": {"b": "d"}}`},

		{`{"a": [{"b":"c"}]}`, `{"a": [1]}`, `{"a": [1]}`},

		{`["a","b"]`, `["c","d"]`, `["c","d"]`},

		{`{"a":"b"}`, `["c"] `, `["c"]`},

		{`{"a":"foo"}`, `null`, `null`},

		{`{"a":"foo"}`, `"bar"`, `"bar"`},

		{`{"e":null}`, `{"a":1}`, `{"e":null,"a":1}`},

		{`[1,2]`, `{"a":"b","c":null}`, `{"a":"b"}`},

		{`{}`, `{"a":{"bb":{"ccc":null}}}`, `{"a":{"bb":{}}}`},
	}
	for i, c := range testCases {
		result, err := jsonMergePatch(c.target, c.patch)
		if err != nil {
			t.Errorf("test case %d: %v", i, err)
			continue
		}
		var resultObject, caseResultObject any
		if err := json.Unmarshal([]byte(result), &resultObject); err != nil {
			t.Error(err)
			continue
		}
		if err := json.Unmarshal([]byte(c.result), &caseResultObject); err != nil {
			t.Error(err)
			continue
		}
		if !reflect.DeepEqual(resultObject, caseResultObject) {
			t.Errorf("expected %v; got %v", caseResultObject, resultObject)
		}
	}
}

func TestCanonicalizeFieldNames(t *testing.T) {
	var testCases = []struct {
		value     string
		prototype proto.Message
		result    string
	}{
		{
			`{"mount_path": "foo"}`,
			&pps.SecretMount{},
			`{"mountPath": "foo"}`,
		},
		{
			`{
				"contexts": {
					"foo": {
						"pachdAddress": "foo"
					},
					"bar": {
						"pachd_address": "quux"
					}
				}
			}`,
			&config.ConfigV2{},
			`{
				"contexts": {
					"foo": {
						"pachdAddress": "foo"
					},
					"bar": {
						"pachdAddress": "quux"
					}
				}
			}`,
		},
		{
			`{
	"pipeline": {
		"name": "wordcount",
		"project": {
			"name": "projectName"
		}
	},
	"transform": {
		"image": "wordcount-image",
		"cmd": ["/binary", "/pfs/data", "/pfs/out"]
	},
	"input": {
		"pfs": {
			"repo": "data",
			"glob": "/*"
		}
	},
	"s3_out": true
}`,
			&pps.CreatePipelineRequest{},
			`{
	"pipeline": {
		"name": "wordcount",
		"project": {
			"name": "projectName"
		}
	},
	"transform": {
		"image": "wordcount-image",
		"cmd": ["/binary", "/pfs/data", "/pfs/out"]
	},
	"input": {
		"pfs": {
			"repo": "data",
			"glob": "/*"
		}
	},
	"s3Out": true
}`,
		},
	}

	for i, c := range testCases {
		var v, r map[string]any
		if err := json.Unmarshal([]byte(c.value), &v); err != nil {
			t.Errorf("test case %d: bad value: %v", i, err)
			continue
		}
		if err := json.Unmarshal([]byte(c.result), &r); err != nil {
			t.Errorf("test case %d: bad value: %v", i, err)
			continue
		}
		result, err := canonicalizeFieldNames(v, c.prototype.ProtoReflect().Descriptor())
		if err != nil {
			t.Errorf("test case %d: %v", i, err)
			continue
		}
		if !reflect.DeepEqual(result, r) {
			t.Errorf("test case %d: expected %v; got %v", i, r, result)
		}
	}
}
