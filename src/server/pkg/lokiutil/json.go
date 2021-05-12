package lokiutil

import (
	"fmt"
	"strconv"
	"time"
	"unsafe"

	json "github.com/json-iterator/go"
	jsoniter "github.com/json-iterator/go"
	"github.com/modern-go/reflect2"
)

func init() {
	jsoniter.RegisterExtension(&jsonExtension{})
}

// ResultType holds the type of the result
type ResultType string

type ResultValue interface {
	Type() ResultType
}

const (
	ResultTypeStream = "streams"
)

// Type implements the promql.Value interface
func (Streams) Type() ResultType { return ResultTypeStream }

// Streams is a slice of Stream
type Streams []Stream

type Stream struct {
	Labels  LabelSet `json:"stream"`
	Entries []Entry  `json:"values"`
}

type LabelSet map[string]string

type Entry struct {
	Timestamp time.Time
	Line      string
}

type QueryResponseData struct {
	ResultType ResultType  `json:"resultType"`
	Result     ResultValue `json:"result"`
	//Statistics interface{} `json:"stats"`
}

type QueryResponse struct {
	Status string            `json:"status"`
	Data   QueryResponseData `json:"data"`
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (q *QueryResponseData) UnmarshalJSON(data []byte) error {
	unmarshal := struct {
		Type   ResultType      `json:"resultType"`
		Result json.RawMessage `json:"result"`
		//Statistics stats.Result    `json:"stats"`
	}{}

	err := json.Unmarshal(data, &unmarshal)
	if err != nil {
		return err
	}

	var value ResultValue

	// unmarshal results
	switch unmarshal.Type {
	case ResultTypeStream:
		var s Streams
		err = json.Unmarshal(unmarshal.Result, &s)
		value = s
	/*case ResultTypeMatrix:
		var m Matrix
		err = json.Unmarshal(unmarshal.Result, &m)
		value = m
	case ResultTypeVector:
		var v Vector
		err = json.Unmarshal(unmarshal.Result, &v)
		value = v
	case ResultTypeScalar:
		var v Scalar
		err = json.Unmarshal(unmarshal.Result, &v)
		value = v*/
	default:
		return fmt.Errorf("unknown type: %s", unmarshal.Type)
	}

	if err != nil {
		return err
	}

	q.ResultType = unmarshal.Type
	q.Result = value
	//q.Statistics = unmarshal.Statistics

	return nil
}

type jsonExtension struct {
	jsoniter.DummyExtension
}

type sliceEntryDecoder struct{}

func (sliceEntryDecoder) Decode(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
	fmt.Println("Decode called")
	*((*[]Entry)(ptr)) = (*((*[]Entry)(ptr)))[:0]
	iter.ReadArrayCB(func(iter *jsoniter.Iterator) bool {
		i := 0
		var ts time.Time
		var line string
		ok := iter.ReadArrayCB(func(iter *jsoniter.Iterator) bool {
			var ok bool
			switch i {
			case 0:
				ts, ok = readTimestamp(iter)
				i++
				return ok
			case 1:
				line = iter.ReadString()
				i++
				if iter.Error != nil {
					return false
				}
				return true
			default:
				iter.ReportError("error reading entry", "array must contains 2 values")
				return false
			}
		})
		if ok {
			*((*[]Entry)(ptr)) = append(*((*[]Entry)(ptr)), Entry{
				Timestamp: ts,
				Line:      line,
			})
			return true
		}
		return false
	})
}

func readTimestamp(iter *jsoniter.Iterator) (time.Time, bool) {
	s := iter.ReadString()
	if iter.Error != nil {
		return time.Time{}, false
	}
	t, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		iter.ReportError("error reading entry timestamp", err.Error())
		return time.Time{}, false

	}
	return time.Unix(0, t), true
}

type entryEncoder struct{}

func (entryEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	// we don't omit-empty with log entries.
	return false
}

func (e *jsonExtension) CreateDecoder(typ reflect2.Type) jsoniter.ValDecoder {
	if typ == reflect2.TypeOf([]Entry{}) {
		return sliceEntryDecoder{}
	}
	return nil
}
