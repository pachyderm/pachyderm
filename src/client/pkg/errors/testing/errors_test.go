package testing

import (
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

// FooErr is an error with a normal receiver for fulfilling the error interface
type FooErr struct {
	val int
}

func (err FooErr) Error() string {
	return fmt.Sprintf("foo-%d", err.val)
}

// PtrFooErr is an error with a pointer receiver for fulfilling the error interface
type PtrFooErr struct {
	val int
}

func (err *PtrFooErr) Error() string {
	return fmt.Sprintf("fooptr-%d", err.val)
}

// IntErr is an interface that also fulfills the error interface
type IntErr interface {
	Int() int
	Error() string
}

// ConcreteIntErr is an implementation of the IntErr interface using a normal receiver
type ConcreteIntErr struct {
	val int
}

func (err ConcreteIntErr) Int() int {
	return err.val
}

func (err ConcreteIntErr) Error() string {
	return fmt.Sprintf("int-%d", err.Int())
}

// ConcreteIntErr is an implementation of the IntErr interface using a pointer receiver
type ConcretePtrIntErr struct {
	val int
}

func (err *ConcretePtrIntErr) Int() int {
	return err.val
}

func (err *ConcretePtrIntErr) Error() string {
	return fmt.Sprintf("intptr-%d", err.Int())
}

// OtherErr is a different error type to be used for failed casting
type OtherErr struct{}

func (err OtherErr) Error() string {
	return "other"
}

func TestAsFooErr(t *testing.T) {
	var err error
	fooerr := &FooErr{}
	otherErr := &OtherErr{}

	err = FooErr{1}
	require.True(t, As(err, &FooErr{}))
	require.False(t, As(err, &OtherErr{}))

	err = FooErr{2}
	require.True(t, As(err, fooerr))
	require.False(t, As(err, otherErr))
	require.Equal(t, fooerr, &FooErr{2})

	err = FooErr{3}
	require.True(t, As(err, &fooerr))
	require.False(t, As(err, &otherErr))
	require.Equal(t, fooerr, &FooErr{3})

	err = &FooErr{4}
	require.True(t, As(err, &FooErr{}))
	require.False(t, As(err, &OtherErr{}))

	err = &FooErr{5}
	require.True(t, As(err, fooerr))
	require.False(t, As(err, otherErr))
	require.Equal(t, fooerr, &FooErr{5})

	err = &FooErr{6}
	require.True(t, As(err, &fooerr))
	require.False(t, As(err, &otherErr))
	require.Equal(t, fooerr, &FooErr{6})
}

func TestAsPtrFooErr(t *testing.T) {
	var err error
	fooptrerr := &PtrFooErr{}
	otherErr := &OtherErr{}

	// these don't compile - fooptrerr can't be used as an error unless it's a pointer
	// err = PtrFooErr{1}
	// require.True(t, As(err, &PtrFooErr{}))
	// require.False(t, As(err, &OtherErr{}))

	// err = PtrFooErr{2}
	// require.True(t, As(err, fooptrerr))
	// require.False(t, As(err, otherErr))
	// require.Equal(t, fooptrerr, &PtrFooErr{2})

	// err = PtrFooErr{3}
	// require.True(t, As(err, &fooptrerr))
	// require.False(t, As(err, &otherErr))
	// require.Equal(t, fooptrerr, &PtrFooErr{3})

	err = &PtrFooErr{4}
	require.True(t, As(err, &PtrFooErr{}))
	require.False(t, As(err, &OtherErr))

	err = &PtrFooErr{5}
	require.True(t, As(err, fooptrerr))
	require.False(t, As(err, otherErr))
	require.Equal(t, fooptrerr, &PtrFooErr{5})

	err = &PtrFooErr{6}
	require.True(t, As(err, &fooptrerr))
	require.False(t, As(err, &otherErr))
	require.Equal(t, fooptrerr, &PtrFooErr{6})
}

func TestAsConcreteIntErr(t *testing.T) {
	var err error
	var interr IntErr
	interr = &ConcreteIntErr{}
	otherErr := &OtherErr{}

	err = ConcreteIntErr{1}
	// this doesn't compile - can't construct an IntErr{} as it's an interface
	// require.True(t, As(err, &IntErr{}))
	require.False(t, As(err, &OtherErr{}))

	err = ConcreteIntErr{2}
	require.True(t, As(err, interr))
	require.False(t, As(err, otherErr))
	require.Equal(t, interr, &ConcreteIntErr{2})

	err = ConcreteIntErr{3}
	require.True(t, As(err, &interr))
	require.False(t, As(err, &otherErr))
	require.Equal(t, interr, &ConcreteIntErr{3})

	err = &ConcreteIntErr{4}
	// this doesn't compile - can't construct an IntErr{} as it's an interface
	// require.True(t, As(err, &IntErr{}))
	require.False(t, As(err, &OtherErr{}))

	err = &ConcreteIntErr{5}
	require.True(t, As(err, interr))
	require.False(t, As(err, otherErr))
	require.Equal(t, interr, &ConcreteIntErr{5})

	err = &ConcreteIntErr{6}
	require.True(t, As(err, &interr))
	require.False(t, As(err, &otherErr))
	require.Equal(t, interr, &ConcreteIntErr{6})
}

func TestAsConcretePtrIntErr(t *testing.T) {
	var err error
	var interr IntErr
	interr = &ConcretePtrIntErr{}
	otherErr := &OtherErr{}

	// these don't compile - ConcretePtrIntErr can't be used as an error unless it's a pointer
	// err = ConcretePtrIntErr{1}
	// require.True(t, As(err, &IntErr{}))
	// require.False(t, As(err, &OtherErr{}))

	// err = ConcretePtrIntErr{2}
	// require.True(t, As(err, interr))
	// require.False(t, As(err, otherErr))
	// require.Equal(t, interr, &ConcretePtrIntErr{2})

	// err = ConcretePtrIntErr{3}
	// require.True(t, As(err, &interr))
	// require.False(t, As(err, &otherErr))
	// require.Equal(t, interr, &ConcretePtrIntErr{3})

	err = &ConcretePtrIntErr{4}
	// this doesn't compile - can't construct an IntErr{} as it's an interface
	// require.True(t, As(err, &IntErr{}))
	require.False(t, As(err, &OtherErr{}))

	err = &ConcretePtrIntErr{5}
	require.True(t, As(err, interr))
	require.False(t, As(err, otherErr))
	require.Equal(t, interr, &ConcretePtrIntErr{5})

	err = &ConcretePtrIntErr{6}
	require.True(t, As(err, &interr))
	require.False(t, As(err, &otherErr))
	require.Equal(t, interr, &ConcretePtrIntErr{6})
}
