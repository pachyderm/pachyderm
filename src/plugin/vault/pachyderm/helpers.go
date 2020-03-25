package pachyderm

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
)

// errMissingField returns a logical response error that prints a consistent
// error message for when a required field is missing.
func errMissingField(field string) *logical.Response {
	return logical.ErrorResponse(fmt.Sprintf("missing required field '%s'", field))
}

// validateFields verifies that no bad arguments were given to the request.
func validateFields(req *logical.Request, data *framework.FieldData) error {
	var unknownFields []string
	for k := range req.Data {
		if _, ok := data.Schema[k]; !ok {
			unknownFields = append(unknownFields, k)
		}
	}

	if len(unknownFields) > 0 {
		return errors.Errorf("unknown fields: %q", unknownFields)
	}

	return nil
}

// putConfig parses and returns the configuration data from the storage backend.
func putConfig(ctx context.Context, s logical.Storage, cfg *config) error {
	entry, err := logical.StorageEntryJSON("config", cfg)
	if err != nil {
		return errors.Wrapf(err, "failed to generate storage entry")
	}
	if err := s.Put(ctx, entry); err != nil {
		return errors.Wrapf(err, "failed to write configuration to storage")
	}
	return nil
}

// getConfig parses and returns the configuration data from the storage backend.
func getConfig(ctx context.Context, s logical.Storage) (*config, error) {
	entry, err := s.Get(ctx, "config")
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get config from storage")
	}
	if entry == nil || len(entry.Value) == 0 {
		return nil, errors.New("no configuration in storage")
	}

	var result config
	if err := entry.DecodeJSON(&result); err != nil {
		return nil, errors.Wrapf(err, "failed to decode configuration")
	}

	return &result, nil
}

// getStringField extracts 'key' from 'req', and either returns the value as a
// string or an error response (vault path handlers seem to return
// ErrorResponse rather than actual errors for malformed requests)
func getStringField(data *framework.FieldData, key string) (string, *logical.Response) {
	valueIface, ok, err := data.GetOkErr(key)
	if err != nil {
		return "", logical.ErrorResponse(fmt.Sprintf("%v: could not extract '%s' from request", err, key))
	}
	// the convention (in e.g. https://github.com/hashicorp/vault/blob/8142b42d951119a73ce46daa3331921b5e21cdee/builtin/logical/aws/path_config_lease.go
	// seems to be to return logical.ErrorResponse for invalid requests, and error for internal errors)
	if !ok {
		return "", errMissingField(key)
	}
	value, ok := valueIface.(string)
	if !ok {
		return "", logical.ErrorResponse(fmt.Sprintf("invalid type for param '%s' (expected string but got %T)", key, valueIface))
	}
	return value, nil
}
