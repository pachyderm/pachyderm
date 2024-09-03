package log

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type obj struct{ x string }

func (o obj) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString(o.x, "exists") // confusing ordering so that obj{"foo"} results in obj:{"foo":"exists"}
	return nil
}

type arr []int

func (a arr) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	for _, x := range a {
		enc.AppendInt(x)
	}
	return nil
}

type stringer string

func (s stringer) String() string {
	return "stringer<" + string(s) + ">"
}

func TestSpan(t *testing.T) {
	testData := []struct {
		name     string
		f        func(ctx context.Context) error
		want     []string
		wantKeys []string
	}{
		{
			name: "simple success",
			f: func(ctx context.Context) error {
				defer Span(ctx, "x")()
				Debug(ctx, "running")
				return nil
			},
			want:     []string{"x: debug: x: span start", "debug: running", "x: debug: x: span finished ok"},
			wantKeys: []string{"caller", "caller", "caller,spanDuration"},
		},
		{
			name: "leveled success",
			f: func(ctx context.Context) error {
				defer SpanL(ctx, "x", ErrorLevel)()
				Info(ctx, "running")
				return nil
			},
			want:     []string{"x: error: x: span start", "info: running", "x: error: x: span finished ok"},
			wantKeys: []string{"caller", "caller", "caller,spanDuration"},
		},
		{
			name: "span nesting",
			f: func(rctx context.Context) error {
				Debug(rctx, "this is a test")
				ctxA, endA := SpanContextL(rctx, "a", InfoLevel, zap.String("ctx", "A"))
				Debug(ctxA, "this is part of a")
				ctxB, endB := SpanContextL(ctxA, "b", ErrorLevel, zap.String("ctx", "B"))
				Debug(ctxB, "this is part of b")
				ctxC, endC := SpanContext(ctxB, "c")
				Error(ctxC, "this is part of c")
				endB()
				endC() // out of order on purpose
				endA()
				return nil
			},
			want: []string{
				"debug: this is a test",
				"a: info: a: span start",
				"a: debug: this is part of a",
				"a.b: error: b: span start",
				"a.b: debug: this is part of b",
				"a.b.c: debug: c: span start",
				"a.b.c: error: this is part of c",
				"a.b: error: b: span finished ok",
				"a.b.c: debug: c: span finished ok",
				"a: info: a: span finished ok",
			},
			wantKeys: []string{"caller", "caller,ctx", "caller,ctx", "caller,ctx", "caller,ctx", "caller,ctx", "caller,ctx", "caller,ctx,spanDuration", "caller,ctx,spanDuration", "caller,ctx,spanDuration"},
		},
		{
			name: "simple error",
			f: func(ctx context.Context) error {
				end := Span(ctx, "x")
				end(zap.Error(errors.New("hi")))
				return nil
			},
			want: []string{
				"x: debug: x: span start",
				"x: debug: x: span failed",
			},
			wantKeys: []string{"caller", "caller,error,spanDuration"},
		},
		{
			name: "simple nil error",
			f: func(ctx context.Context) error {
				end := Span(ctx, "x")
				end(zap.Error(nil))
				return nil
			},
			want: []string{
				"x: debug: x: span start",
				"x: debug: x: span finished ok",
			},
			wantKeys: []string{"caller", "caller,spanDuration"},
		},
		{
			name: "simple namederror",
			f: func(ctx context.Context) error {
				end := Span(ctx, "x")
				end(zap.NamedError("totally_not_a_failure", errors.New("hi")))
				return nil
			},
			want: []string{
				"x: debug: x: span start",
				"x: debug: x: span failed",
			},
			wantKeys: []string{"caller", "caller,spanDuration,totally_not_a_failure"},
		},
		{
			name: "leveled error",
			f: func(ctx context.Context) error {
				end := Span(ctx, "x")
				end(ErrorL(errors.New("hi"), ErrorLevel))
				return nil
			},
			want: []string{
				"x: debug: x: span start",
				"x: error: x: span failed",
			},
			wantKeys: []string{"caller", "caller,error,spanDuration"},
		},
		{
			name: "leveled nil error",
			f: func(ctx context.Context) error {
				end := Span(ctx, "x")
				end(ErrorL(nil, ErrorLevel))
				return nil
			},
			want: []string{
				"x: debug: x: span start",
				"x: debug: x: span finished ok",
			},
			wantKeys: []string{"caller", "caller,spanDuration"},
		},
		// { // This one causes zap to panic, so it's safe for us to not handle.
		// 	name: "leveled nil error, injected sneakily",
		// 	f: func(ctx context.Context) error {
		// 		end := Span(ctx, "x")
		// 		end(zapcore.Field{
		// 			Key:       "error",
		// 			Type:      zapcore.ErrorType,
		// 			Interface: nil,
		// 		})
		// 		return nil
		// 	},
		// 	want: []string{
		// 		"x: debug: span start",
		// 		"x: debug: span finished ok",
		// 	},
		// },
		{
			name: "simple errorp; success",
			f: func(ctx context.Context) (err error) {
				defer Span(ctx, "x")(Errorp(&err))
				Debug(ctx, "hi")
				return nil
			},
			want: []string{
				"x: debug: x: span start",
				"debug: hi",
				"x: debug: x: span finished ok",
			},
			wantKeys: []string{"caller", "caller", "caller,spanDuration"},
		},
		{
			name: "simple errorp; failure",
			f: func(ctx context.Context) (err error) {
				defer Span(ctx, "x")(Errorp(&err))
				Debug(ctx, "hi")
				return errors.New("failed")
			},
			want: []string{
				"x: debug: x: span start",
				"debug: hi",
				"x: debug: x: span failed",
			},
			wantKeys: []string{"caller", "caller", "caller,error,spanDuration"},
		},
		{
			name: "leveled errorp; failure",
			f: func(ctx context.Context) (err error) {
				defer Span(ctx, "x")(ErrorpL(&err, ErrorLevel))
				Debug(ctx, "hi")
				return errors.New("failed")
			},
			want: []string{
				"x: debug: x: span start",
				"debug: hi",
				"x: error: x: span failed",
			},
			wantKeys: []string{"caller", "caller", "caller,error,spanDuration"},
		},
		{
			name: "ensure nothing gets filtered",
			f: func(ctx context.Context) error {
				b := true
				var c128 complex128 = complex(12, 8)
				var c64 complex64 = complex(6, 4)
				d := time.Minute
				err := errors.New("error")
				f32 := float32(32)
				f64 := float64(64)
				i := int(42)
				i8 := int8(8)
				i16 := int16(16)
				i32 := int32(32)
				i64 := int64(64)
				u := uint(42)
				u8 := uint8(8)
				u16 := uint16(16)
				u32 := uint32(32)
				u64 := uint64(64)
				str := "string"
				t := time.Now()
				up := uintptr(1)

				defer Span(ctx, "x")(
					zap.Any("any", obj{"any"}),
					zap.Array("array", arr{1, 2, 3}),
					zap.Binary("binary", []byte{'p', 'a', 'c', 'h'}),
					zap.Bool("bool", b),
					zap.Boolp("boolp", &b),
					zap.Bools("bools", []bool{true, false, true}),
					zap.ByteString("bytestring", []byte("hi")),
					zap.ByteStrings("bytestrings", [][]byte{[]byte("hello"), []byte("world")}),
					zap.Complex128("complex128", c128),
					zap.Complex128p("complex128p", &c128),
					zap.Complex128s("complex128s", []complex128{c128, c128}),
					zap.Complex64("complex64", c64),
					zap.Complex64p("complex64p", &c64),
					zap.Complex64s("complex64s", []complex64{c64, c64}),
					zap.Duration("duration", d),
					zap.Durationp("durationp", &d),
					zap.Durations("durations", []time.Duration{d, d}),
					zap.Error(err),
					ErrorL(err, ErrorLevel),
					Errorp(&err),
					ErrorpL(&err, ErrorLevel),
					zap.Errors("errors", []error{errors.New("1"), errors.New("2")}),
					zap.Float32("float32", f32),
					zap.Float32p("float32p", &f32),
					zap.Float32s("float32s", []float32{f32, f32}),
					zap.Float64("float64", f64),
					zap.Float64p("float64p", &f64),
					zap.Float64s("float64s", []float64{f64, f64}),
					zap.Inline(obj{"inline"}),
					zap.Int("int", i),
					zap.Intp("intp", &i),
					zap.Ints("ints", []int{i, i}),
					zap.Int16("int16", i16),
					zap.Int16p("int16p", &i16),
					zap.Int16s("int16s", []int16{i16, i16}),
					zap.Int32("int32", i32),
					zap.Int32p("int32p", &i32),
					zap.Int32s("int32s", []int32{i32, i32}),
					zap.Int64("int64", i64),
					zap.Int64p("int64p", &i64),
					zap.Int64s("int64s", []int64{i64, i64}),
					zap.Int8("int8", i8),
					zap.Int8p("int8p", &i8),
					zap.Int8s("int8s", []int8{i8, i8}),
					zap.NamedError("namederror", errors.New("named")),
					// zap.Namespace() is handled out of order
					zap.Object("object", obj{"object"}),
					zap.Reflect("reflect", obj{"reflected"}),
					zap.Skip(),
					zap.Stack("stack"),
					zap.StackSkip("stack2", 2),
					zap.String("string", str),
					zap.Stringer("stringer", stringer("hello")),
					zap.Stringp("stringp", &str),
					zap.Strings("strings", []string{str, str}),
					zap.Time("timeX", t), // timeX because when unmarshaling json this clashes with the time key
					zap.Timep("timep", &t),
					zap.Times("times", []time.Time{t, t}),
					zap.Uint("uint", u),
					zap.Uintp("uintp", &u),
					zap.Uints("uints", []uint{u, u}),
					zap.Uint16("uint16", u16),
					zap.Uint16p("uint16p", &u16),
					zap.Uint16s("uint16s", []uint16{u16, u16}),
					zap.Uint32("uint32", u32),
					zap.Uint32p("uint32p", &u32),
					zap.Uint32s("uint32s", []uint32{u32, u32}),
					zap.Uint64("uint64", u64),
					zap.Uint64p("uint64p", &u64),
					zap.Uint64s("uint64s", []uint64{u64, u64}),
					zap.Uint8("uint8", u8),
					zap.Uint8p("uint8p", &u8),
					zap.Uint8s("uint8s", []uint8{u8, u8}),
					zap.Uintptr("uintptr", up),
					zap.Uintptrp("uintptrp", &up),
					zap.Uintptrs("uintptrs", []uintptr{up, up}),
					zap.Namespace("namespace"),
					zap.String("does-not-appear", "inside namespace"),
				)
				return nil
			},
			want: []string{
				"x: debug: x: span start",
				"x: error: x: span failed",
			},
			wantKeys: []string{
				"caller",
				"any,array,binary,bool,boolp,bools,bytestring,bytestrings,caller,complex128,complex128p,complex128s,complex64,complex64p,complex64s,duration,durationp,durations,error,errors,float32,float32p,float32s,float64,float64p,float64s,inline,int,int16,int16p,int16s,int32,int32p,int32s,int64,int64p,int64s,int8,int8p,int8s,intp,ints,namederror,namespace,object,reflect,spanDuration,stack,stack2,string,stringer,stringp,strings,timeX,timep,times,uint,uint16,uint16p,uint16s,uint32,uint32p,uint32s,uint64,uint64p,uint64s,uint8,uint8p,uint8s,uintp,uintptr,uintptrp,uintptrs,uints",
			},
		},
	}
	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			ctx, h := testWithCaptureParallel(t, zap.Development())
			test.f(ctx) //nolint:errcheck // We don't care if this fails; it's just to make errorp testing easier.
			if diff := cmp.Diff(h.Logs(), test.want, formatLogs(simple)); diff != "" {
				t.Errorf("logs (-got +want):\n%s", diff)
			}
			if diff := cmp.Diff(h.Logs(), test.wantKeys, formatLogs(keys)); diff != "" {
				t.Errorf("log keys (-got +want):\n%s", diff)
			}
			var dump bool
			for i, l := range h.Logs() {
				if !strings.HasPrefix(l.Caller, "log/span_test.go:") {
					t.Errorf("line %d: caller:\n  got:  %v\n want: ^log/span_test.go:", i+1, l.Caller)
					dump = true
				}
			}
			if dump {
				for i, l := range h.Logs() {
					t.Logf("line %d: %s", i+1, l.Orig)
				}
			}
		})
	}
}

func TestDeadlineSpan(t *testing.T) {
	ctx, c := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer c()
	ctx = TestParallel(ctx, t)
	sctx, done := SpanContext(ctx, "Test", zap.String("string", "string"))
	Info(sctx, "before")
	<-ctx.Done()
	Info(sctx, "after")
	done()
}
