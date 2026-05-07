package env

import (
	"testing"

	"github.com/shayanderson/go-project/v2/internal/test"
)

func TestBool(t *testing.T) {
	t.Setenv("TEST_BOOL_TRUE", "true")
	t.Setenv("TEST_BOOL_FALSE", "no")

	test.True(t, Bool("TEST_BOOL_TRUE", false))
	test.False(t, Bool("TEST_BOOL_FALSE", true))
	test.True(t, Bool("TEST_BOOL_MISSING", true))
}

func TestInt(t *testing.T) {
	t.Setenv("TEST_INT_VALID", "42")
	t.Setenv("TEST_INT_INVALID", "abc")

	test.Equal(t, 42, Int("TEST_INT_VALID", 1))
	test.Equal(t, 7, Int("TEST_INT_INVALID", 7))
	test.Equal(t, 9, Int("TEST_INT_MISSING", 9))
}

func TestMustBool(t *testing.T) {
	t.Setenv("TEST_MUST_BOOL_TRUE", "yes")
	test.True(t, MustBool("TEST_MUST_BOOL_TRUE"))

	test.Panics(t, func() { MustBool("TEST_MUST_BOOL_MISSING") })
	t.Setenv("TEST_MUST_BOOL_EMPTY", "")
	test.Panics(t, func() { MustBool("TEST_MUST_BOOL_EMPTY") })
}

func TestMustInt(t *testing.T) {
	t.Setenv("TEST_MUST_INT_VALID", "123")
	test.Equal(t, 123, MustInt("TEST_MUST_INT_VALID"))

	test.Panics(t, func() { MustInt("TEST_MUST_INT_MISSING") })
	t.Setenv("TEST_MUST_INT_EMPTY", "")
	test.Panics(t, func() { MustInt("TEST_MUST_INT_EMPTY") })
	t.Setenv("TEST_MUST_INT_INVALID", "x")
	test.Panics(t, func() { MustInt("TEST_MUST_INT_INVALID") })
}

func TestMustString(t *testing.T) {
	t.Setenv("TEST_MUST_STRING", "value")
	test.Equal(t, "value", MustString("TEST_MUST_STRING"))

	test.Panics(t, func() { MustString("TEST_MUST_STRING_MISSING") })
	t.Setenv("TEST_MUST_STRING_EMPTY", "")
	test.Panics(t, func() { MustString("TEST_MUST_STRING_EMPTY") })
}

func TestMustStrings(t *testing.T) {
	t.Setenv("TEST_MUST_STRINGS", " a, b , ,c ")
	got := MustStrings("TEST_MUST_STRINGS")
	want := []string{"a", "b", "c"}
	test.Equal(t, want, got)

	test.Panics(t, func() { MustStrings("TEST_MUST_STRINGS_MISSING") })
	t.Setenv("TEST_MUST_STRINGS_EMPTY", "")
	test.Panics(t, func() { MustStrings("TEST_MUST_STRINGS_EMPTY") })
}

func TestString(t *testing.T) {
	t.Setenv("TEST_STRING", "hello")
	test.Equal(t, "hello", String("TEST_STRING", "fallback"))
	test.Equal(t, "fallback", String("TEST_STRING_MISSING", "fallback"))
}

func TestStrings(t *testing.T) {
	t.Setenv("TEST_STRINGS", " one, two ,, three ")
	got := Strings("TEST_STRINGS", []string{"fallback"})
	want := []string{"one", "two", "three"}
	test.Equal(t, want, got)

	fallback := []string{"fallback", "values"}
	got = Strings("TEST_STRINGS_MISSING", fallback)
	test.Equal(t, fallback, got)
}

func TestParseBool(t *testing.T) {
	trueCases := []string{"1", "true", "TRUE", "True", "yes", "YES", "on", "ON"}
	for _, v := range trueCases {
		test.True(t, parseBool(v))
	}

	falseCases := []string{"0", "false", "False", "no", "off", "", "random"}
	for _, v := range falseCases {
		test.False(t, parseBool(v))
	}
}

func TestSplitAndTrim(t *testing.T) {
	got := splitAndTrim(" a, b ,, c , ", ",")
	want := []string{"a", "b", "c"}
	test.Equal(t, want, got)

	got = splitAndTrim("", ",")
	test.Equal(t, []string{}, got)
}
