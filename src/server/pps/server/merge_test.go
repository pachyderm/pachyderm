package server

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"google.golang.org/protobuf/encoding/protojson"
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
		result, err := jsonMergePatch(c.target, c.patch, nil)
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
				"cmd": [
					"/binary",
					"/pfs/data",
					"/pfs/out"
				]
				},
				"input": {
				"cross": [
					{
						"pfs": {
							"repo": "data",
							"glob": "/*"
						}
					},
					{
						"pfs": {
							"repo": "data2",
							"glob": "/*"
						}
					}
				]
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
				"cmd": [
					"/binary",
					"/pfs/data",
					"/pfs/out"
				]
				},
				"input": {
				"cross": [
					{
						"pfs": {
							"repo": "data",
							"glob": "/*"
						}
					},
					{
						"pfs": {
							"repo": "data2",
							"glob": "/*"
						}
					}
				]
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

func TestFoo(t *testing.T) {
	var testCases = []struct {
		value     string
		prototype proto.Message
		result    string
	}{
		{
			`{"mount_path": "foo"}`,
			&pps.SecretMount{
				MountPath: "foo",
			},
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
			&config.ConfigV2{
				Contexts: map[string]*config.Context{
					"foo": {PachdAddress: "foo"},
					"bar": {PachdAddress: "quux"},
				},
			},
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
			&pps.CreatePipelineRequest{
				Pipeline: &pps.Pipeline{
					Project: &pfs.Project{
						Name: "projectName",
					},
					Name: "wordcount",
				},
				Transform: &pps.Transform{
					Image: "wordcount-image",
					Cmd:   []string{"/binary", "/pfs/data", "/pfs/out"},
				},
				Input: &pps.Input{
					Pfs: &pps.PFSInput{
						Repo: "data",
						Glob: "/*",
					},
				},
				S3Out: true,
			},
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
				"cmd": [
					"/binary",
					"/pfs/data",
					"/pfs/out"
				]
				},
				"input": {
				"cross": [
					{
						"pfs": {
							"repo": "data",
							"glob": "/*"
						}
					},
					{
						"pfs": {
							"repo": "data2",
							"glob": "/*"
						}
					}
				]
				},
				"s3_out": true
			}`,
			&pps.CreatePipelineRequest{
				Pipeline: &pps.Pipeline{
					Project: &pfs.Project{
						Name: "projectName",
					},
					Name: "wordcount",
				},
				Transform: &pps.Transform{
					Image: "wordcount-image",
					Cmd:   []string{"/binary", "/pfs/data", "/pfs/out"},
				},
				Input: &pps.Input{
					Cross: []*pps.Input{
						{
							Pfs: &pps.PFSInput{
								Repo: "data",
								Glob: "/*",
							},
						},
						{
							Pfs: &pps.PFSInput{
								Repo: "data2",
								Glob: "/*",
							},
						},
					},
				},
				S3Out: true,
			},
			`{
				"pipeline": {
				"name": "wordcount",
				"project": {
					"name": "projectName"
				}
				},
				"transform": {
				"image": "wordcount-image",
				"cmd": [
					"/binary",
					"/pfs/data",
					"/pfs/out"
				]
				},
				"input": {
				"cross": [
					{
						"pfs": {
							"repo": "data",
							"glob": "/*"
						}
					},
					{
						"pfs": {
							"repo": "data2",
							"glob": "/*"
						}
					}
				]
				},
				"s3Out": true
			}`,
		},
		{`{"fileFormat":{"type":"CSV"}}`, &pfs.SQLDatabaseEgress{FileFormat: &pfs.SQLDatabaseEgress_FileFormat{Type: pfs.SQLDatabaseEgress_FileFormat_CSV}}, `{"fileFormat":{"type":"CSV"}}`},
		// should canonicalize enums to strings when possible
		{`{"fileFormat":{"type":1}}`, &pfs.SQLDatabaseEgress{FileFormat: &pfs.SQLDatabaseEgress_FileFormat{Type: pfs.SQLDatabaseEgress_FileFormat_CSV}}, `{"fileFormat":{"type":"CSV"}}`},
		// and to numbers when unknown
		{`{"fileFormat":{"type":16783}}`, &pfs.SQLDatabaseEgress{FileFormat: &pfs.SQLDatabaseEgress_FileFormat{Type: pfs.SQLDatabaseEgress_FileFormat_Type(16783)}}, `{"fileFormat":{"type":16783}}`},
		// int64s that exceed float64’s capacity should still work
		{`{"accept_return_code": [9223372036854775807]}`, &pps.Transform{AcceptReturnCode: []int64{9223372036854775807}}, `{"acceptReturnCode": [9223372036854775807]}`},
	}
	// have to use json.Number for numbers in order to preserve precision
	unmarshal := func(s string, dest any) error {
		d := json.NewDecoder(strings.NewReader(s))
		d.UseNumber()
		return d.Decode(dest) //nolint:wrapcheck
	}
	for i, c := range testCases {
		var obj, canObj, result any
		can, err := makeMessageCanonicalizer(c.prototype.ProtoReflect().Descriptor())
		if err != nil {
			t.Errorf("test case %d: couldn’t make canonicalizer for %T: %v", i, c.prototype, err)
			continue
		}
		if err := unmarshal(c.value, &obj); err != nil {
			t.Errorf("test case %d: couldn’t unmarshal value: %v", i, err)
			continue
		}
		if err := unmarshal(c.result, &result); err != nil {
			t.Errorf("test case %d: couldn’t unmarshal result: %v", i, err)
			continue
		}

		if canObj, err = can(obj); err != nil {
			t.Errorf("test case %d: couldn’t canonicalize object: %v", i, err)
			continue
		}
		if !reflect.DeepEqual(canObj, result) {
			t.Errorf("test case %d: %v ≠ %v", i, canObj, result)
			continue
		}
		canObjJSON, err := json.Marshal(canObj)
		if err != nil {
			t.Errorf("test case %d: could not marshal %v: %v", i, canObj, err)
			continue
		}
		protoObj := proto.Clone(c.prototype)
		if err := protojson.Unmarshal([]byte(canObjJSON), protoObj); err != nil {
			t.Errorf("test case %d: could not unmarshal %s: %v", i, canObjJSON, err)
			continue
		}
		if !proto.Equal(protoObj, c.prototype) {
			t.Errorf("test case %d: %v ≠ prototype %v", i, protoObj, c.prototype)
			continue
		}
	}
}
