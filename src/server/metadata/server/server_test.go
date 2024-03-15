package server

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pachyderm/pachyderm/v2/src/metadata"
)

func TestEditMetadata(t *testing.T) {
	testData := []struct {
		name    string
		md      map[string]string
		edit    *metadata.Edit
		want    map[string]string
		wantErr bool
	}{
		{
			name: "no op",
		},
		{
			name: "add empty",
			edit: &metadata.Edit{
				Op: &metadata.Edit_AddKey_{
					AddKey: &metadata.Edit_AddKey{
						Key:   "key",
						Value: "value",
					},
				},
			},
			want: map[string]string{"key": "value"},
		},
		{
			name: "add duplicate",
			md:   map[string]string{"key": "value"},
			edit: &metadata.Edit{
				Op: &metadata.Edit_AddKey_{
					AddKey: &metadata.Edit_AddKey{
						Key:   "key",
						Value: "value2",
					},
				},
			},
			want:    map[string]string{"key": "value"},
			wantErr: true,
		},
		{
			name: "add",
			md:   map[string]string{"key": "value"},
			edit: &metadata.Edit{
				Op: &metadata.Edit_AddKey_{
					AddKey: &metadata.Edit_AddKey{
						Key:   "key2",
						Value: "value2",
					},
				},
			},
			want: map[string]string{"key": "value", "key2": "value2"},
		},
		{
			name: "edit empty",
			edit: &metadata.Edit{
				Op: &metadata.Edit_EditKey_{
					EditKey: &metadata.Edit_EditKey{
						Key:   "key",
						Value: "value",
					},
				},
			},
			want: map[string]string{"key": "value"},
		},
		{
			name: "edit existing",
			md:   map[string]string{"key": "value"},
			edit: &metadata.Edit{
				Op: &metadata.Edit_EditKey_{
					EditKey: &metadata.Edit_EditKey{
						Key:   "key",
						Value: "value2",
					},
				},
			},
			want: map[string]string{"key": "value2"},
		},
		{
			name: "edit new",
			md:   map[string]string{"key": "value"},
			edit: &metadata.Edit{
				Op: &metadata.Edit_EditKey_{
					EditKey: &metadata.Edit_EditKey{
						Key:   "key2",
						Value: "value2",
					},
				},
			},
			want: map[string]string{"key": "value", "key2": "value2"},
		},
		{
			name: "delete empty",
			edit: &metadata.Edit{
				Op: &metadata.Edit_DeleteKey_{
					DeleteKey: &metadata.Edit_DeleteKey{
						Key: "key",
					},
				},
			},
		},
		{
			name: "delete",
			md:   map[string]string{"key": "value"},
			edit: &metadata.Edit{
				Op: &metadata.Edit_DeleteKey_{
					DeleteKey: &metadata.Edit_DeleteKey{
						Key: "key",
					},
				},
			},
		},
		{
			name: "replace empty",
			edit: &metadata.Edit{
				Op: &metadata.Edit_Replace_{
					Replace: &metadata.Edit_Replace{
						Replacement: map[string]string{"key": "value", "key2": "value2"},
					},
				},
			},
			want: map[string]string{"key": "value", "key2": "value2"},
		},
		{
			name: "replace empty",
			md:   map[string]string{"key": "original", "key2": "original2"},
			edit: &metadata.Edit{
				Op: &metadata.Edit_Replace_{
					Replace: &metadata.Edit_Replace{
						Replacement: map[string]string{"key": "value"},
					},
				},
			},
			want: map[string]string{"key": "value"},
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			var got map[string]string
			if test.md != nil {
				got = make(map[string]string)
				for k, v := range test.md {
					got[k] = v
				}
			}
			err := editMetadata(test.edit, &got)
			if diff := cmp.Diff(test.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("metadata (-want +got):\n%s", diff)
			}
			if err == nil && test.wantErr {
				t.Error("expected error, but got success")
				return
			}
			if err != nil && !test.wantErr {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
