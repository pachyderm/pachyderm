package pfsdb

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

// jsonMap wraps a map[string]string that is stored as JSONB in the database.  While structs can be
// directly inserted into JSONB columns by the pgx driver, a raw map[string]string cannot be.
type jsonMap struct {
	Data map[string]string
}

var _ sql.Scanner = (*jsonMap)(nil)
var _ driver.Valuer = (*jsonMap)(nil)

// Scan implements database/sql.Scanner.
func (c *jsonMap) Scan(src any) error {
	content, ok := src.([]byte)
	if !ok {
		return errors.Errorf("jsonMap scan source is %T, not []byte", src)
	}
	// Postgres won't let us constrain the type of the JSONB column to map[string]string. (It
	// will let us constrain it to type 'object', so it's guaranteed to always be a map of
	// something).  This means that someone could directly edit the database (perhaps a
	// migration) and add data that won't unmarshal into that type.  Then all getters on the
	// data with this column would hard-fail and be unfixable without manually editing the
	// database.  We would rather not fail hard under these conditions, so we do a two-step
	// unmarshal; get all the key -> json pairs, and then interpet each value as a string.  If
	// the value unmarhsals into a string, then it's a Go string.  If it doesn't, then the Go
	// string is just the JSON bytes that were at that value.  This would let the user at least
	// use Pachyderm APIs to replace the content of the fautly key.
	//
	// This will not let users use the API to store arbitrary JSON data in the column; the
	// marshaling step below still requires a map[string]string and map[string]string{"key":
	// "[1, 2, 3]"} is still string -> string in JSON, not string -> array.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(content, &raw); err != nil {
		return errors.Wrap(err, "unmarshal database JSON passed to a jsonMap into map[string]RawMessage")
	}
	c.Data = make(map[string]string)
	for k, rawValue := range raw {
		var v string
		if err := json.Unmarshal(rawValue, &v); err != nil {
			v = string(rawValue)
		}
		c.Data[k] = v
	}
	return nil
}

// Value implements database/sql/driver.Valuer.
func (c jsonMap) Value() (driver.Value, error) {
	// This has a non-pointer receiver because it's easy to accidentally insert jsonMap{...}
	// instead of &jsonMap{...}.  If you do that, then it will marshal this struct to JSON
	// automatically, instead of marshalling the underlying data.
	if c.Data == nil || len(c.Data) == 0 {
		return []byte(`{}`), nil
	}
	content, err := json.Marshal(c.Data)
	if err != nil {
		return nil, errors.Wrap(err, "marshal jsonMap Data to JSON")
	}
	return content, nil
}
