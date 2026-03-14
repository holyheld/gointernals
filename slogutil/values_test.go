package slogutil_test

import (
	"testing"

	slogutil "github.com/holyheld/gointernals/slogutil"
)

func TestStringLike(t *testing.T) {
	t.Parallel()

	const expected = "key=value"

	t.Run("regular string", func(t *testing.T) {
		t.Parallel()

		got := slogutil.StringLike("key", "value")

		if got.String() != expected {
			t.Errorf("StringLike() want %s got %s", expected, got)
		}
	})

	t.Run("string-like (underlying string) type", func(t *testing.T) {
		t.Parallel()

		v := stringLikeValue("value")

		got := slogutil.StringLike("key", v)

		if got.String() != expected {
			t.Errorf("StringLike() want %s got %s", expected, got)
		}
	})

	t.Run("fmt.Stringer value", func(t *testing.T) {
		t.Parallel()

		v := stringerValue{}

		got := slogutil.StringLike("key", v)

		if got.String() != expected {
			t.Errorf("StringLike() want %s got %s", expected, got)
		}
	})

	t.Run("fmt.Stringer ptr value", func(t *testing.T) {
		t.Parallel()

		v := &stringerPtrValue{}

		got := slogutil.StringLike("key", v)

		if got.String() != expected {
			t.Errorf("StringLike() want %s got %s", expected, got)
		}
	})

	t.Run("fmt.GoStringer value", func(t *testing.T) {
		t.Parallel()

		v := goStringerValue{}

		got := slogutil.StringLike("key", v)

		if got.String() != expected {
			t.Errorf("StringLike() want %s got %s", expected, got)
		}
	})

	t.Run("fmt.GoStringer ptr value", func(t *testing.T) {
		t.Parallel()

		v := &goStringerPtrValue{}

		got := slogutil.StringLike("key", v)

		if got.String() != expected {
			t.Errorf("StringLike() want %s got %s", expected, got)
		}
	})
}

const returnValue = "value"

type stringLikeValue string

type stringerValue struct{}

func (v stringerValue) String() string {
	return returnValue
}

type goStringerValue struct{}

func (v goStringerValue) GoString() string {
	return returnValue
}

type stringerPtrValue struct{}

func (v *stringerPtrValue) String() string {
	return returnValue
}

type goStringerPtrValue struct{}

func (v *goStringerPtrValue) GoString() string {
	return returnValue
}
