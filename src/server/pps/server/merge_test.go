package server

import (
	"encoding/json"
	"reflect"
	"testing"
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
