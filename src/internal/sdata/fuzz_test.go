package sdata

import (
	"database/sql"
	"time"

	fuzz "github.com/google/gofuzz"
)

func addFuzzFuncs(fz *fuzz.Fuzzer) {
	fz.Funcs(
		func(ti *time.Time, co fuzz.Continue) {
			*ti = time.Now()
		},
		func(x *sql.NullInt64, co fuzz.Continue) {
			if co.RandBool() {
				x.Valid = true
				x.Int64 = co.Int63()
			} else {
				x.Valid = false
			}
		},
		func(x *sql.NullString, co fuzz.Continue) {
			n := co.Intn(10)
			if n < 3 {
				x.Valid = true
				x.String = co.RandString()
			} else if n < 6 {
				x.Valid = true
				x.String = ""
			} else {
				x.Valid = false
			}
		},
		fuzz.UnicodeRange{First: '!', Last: '~'}.CustomStringFuzzFunc(),
		func(x *interface{}, co fuzz.Continue) {
			switch co.Intn(6) {
			case 0:
				*x = int16(co.Int())
			case 1:
				*x = int32(co.Int())
			case 2:
				*x = int64(co.Int())
			case 3:
				*x = co.Float32()
			case 4:
				*x = co.Float64()
			case 5:
				*x = co.RandString()
			}
		},
	)
}
