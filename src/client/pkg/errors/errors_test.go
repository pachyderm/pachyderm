package errors

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

// FooErr is an error with a pointer receiver for fulfilling the error interface
type FooPtrErr struct {
	val int
}

func (err *FooPtrErr) Error() string {
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

func TestAsFooPtrErr(t *testing.T) {
	fooptrerr := &FooPtrErr{}
	// doesn't compile - fooptrerr can't be used as an error unless it's a pointer
	// err = FooPtrErr{1}
	// require.True(t, As(err, &FooPtrErr{}))
	// require.False(t, As(err, &OtherErr{}))

	// err = FooPtrErr{2}
	// fmt.Printf("SafeErrorsAs(FooPtrErr{}, fooptrerr): %v\n", SafeErrorsAs(err, fooptrerr))       // True
	// require.True(t, As(err, &fooerr))
	// require.False(t, As(err, &otherErr))
	// require.Equal(t, fooerr, &FooErr{6})

	// err = FooPtrErr{3}
	// fmt.Printf("SafeErrorsAs(FooPtrErr{}, &fooptrerr): %v\n", SafeErrorsAs(err, &fooptrerr))     // False
	// require.True(t, As(err, &fooerr))
	// require.False(t, As(err, &otherErr))
	// require.Equal(t, fooerr, &FooErr{6})

	err = &FooPtrErr{4}
	fmt.Printf("SafeErrorsAs(&FooPtrErr{}, &FooPtrErr{}): %v    - %v (nil)\n", SafeErrorsAs(err, &FooPtrErr{}), nil) // True
	require.True(t, As(err, &fooerr))
	require.False(t, As(err, &otherErr))
	require.Equal(t, fooerr, &FooErr{6})

	err = &FooPtrErr{5}
	fmt.Printf("SafeErrorsAs(&FooPtrErr{}, fooptrerr): %v       - %v (2)\n", SafeErrorsAs(err, fooptrerr), fooptrerr) // True
	require.True(t, As(err, &fooerr))
	require.False(t, As(err, &otherErr))
	require.Equal(t, fooerr, &FooErr{6})

	err = &FooPtrErr{6}
	fmt.Printf("SafeErrorsAs(&FooPtrErr{}, &fooptrerr): %v      - %v (3)\n", SafeErrorsAs(err, &fooptrerr), fooptrerr) // True
	require.True(t, As(err, &fooerr))
	require.False(t, As(err, &otherErr))
	require.Equal(t, fooerr, &FooErr{6})
}

func TestAsConcreteIntErr(t *testing.T) {
	var interr IntErr
	interr = &ConcreteIntErr{}

	// doesn't compile - can't construct an IntErr{} as it's an interface
	// err = ConcreteIntErr{1}
	// require.True(t, As(err, &IntErr{}))
	// require.False(t, As(err, &OtherErr{}))

	err = ConcreteIntErr{2}
	fmt.Printf("SafeErrorsAs(ConcreteIntErr{}, interr): %v      - %v (2)\n", SafeErrorsAs(err, interr), interr) // True

	err = ConcreteIntErr{3}
	fmt.Printf("SafeErrorsAs(ConcreteIntErr{}, &interr): %v     - %v (3)\n", SafeErrorsAs(err, &interr), interr) // True

	// doesn't compile - can't construct an IntErr{} as it's an interface
	// err = &ConcreteIntErr{4}
	// require.True(t, As(err, &IntErr{}))
	// require.False(t, As(err, &OtherErr{}))

	err = &ConcreteIntErr{5}
	fmt.Printf("SafeErrorsAs(&ConcreteIntErr{}, interr): %v     - %v (5)\n", SafeErrorsAs(err, interr), interr) // True

	err = &ConcreteIntErr{6}
	fmt.Printf("SafeErrorsAs(&ConcreteIntErr{}, &interr): %v    - %v (6)\n", SafeErrorsAs(err, &interr), interr) // True
}

func TestAsConcretePtrIntErr(t *testing.T) {
	interr = &ConcretePtrIntErr{}

	// doesn't compile - can't construct an IntErr{} as it's an interface
	/*
		err = ConcretePtrIntErr{1}
		fmt.Printf("SafeErrorsAs(ConcretePtrIntErr{}, &IntErr{}): %v         - %v\n", SafeErrorsAs(err, &IntErr{}), nil) // True
	*/
	// doesn't compile - concrete err can't be used as an error unless it's a pointer
	/*
		err = ConcretePtrIntErr{2}
		fmt.Printf("SafeErrorsAs(ConcretePtrIntErr{}, interr): %v    - %v\n", SafeErrorsAs(err, interr), interr) // True
		err = ConcretePtrIntErr{3}
		fmt.Printf("SafeErrorsAs(ConcretePtrIntErr{}, &interr): %v   - %v\n", SafeErrorsAs(err, &interr), interr) // True
	*/

	// doesn't compile - can't construct an IntErr{} as it's an interface
	/*
		err = &ConcretePtrIntErr{4}
		fmt.Printf("SafeErrorsAs(&ConcretePtrIntErr{}, &IntErr{}): %v - %v\n", SafeErrorsAs(err, &IntErr{}), nil) // False
	*/
	err = &ConcretePtrIntErr{5}
	fmt.Printf("SafeErrorsAs(&ConcretePtrIntErr{}, interr): %v  - %v (5)\n", SafeErrorsAs(err, interr), interr) // True
	err = &ConcretePtrIntErr{6}
	fmt.Printf("SafeErrorsAs(&ConcretePtrIntErr{}, &interr): %v - %v (6)\n", SafeErrorsAs(err, &interr), interr) // True
}
