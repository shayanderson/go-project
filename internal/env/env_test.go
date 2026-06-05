package env

import (
	"reflect"
	"testing"
)

func TestBool(t *testing.T) {
	t.Setenv("TEST_BOOL_TRUE", "true")
	t.Setenv("TEST_BOOL_FALSE", "no")

	if !Bool("TEST_BOOL_TRUE", false) {
		t.Error("expected TEST_BOOL_TRUE to be true")
	}
	if Bool("TEST_BOOL_FALSE", true) {
		t.Error("expected TEST_BOOL_FALSE to be false")
	}
	if !Bool("TEST_BOOL_MISSING", true) {
		t.Error("expected TEST_BOOL_MISSING to be true")
	}
}

func TestInt(t *testing.T) {
	t.Setenv("TEST_INT_VALID", "42")
	t.Setenv("TEST_INT_INVALID", "abc")

	if Int("TEST_INT_VALID", 1) != 42 {
		t.Error("expected TEST_INT_VALID to be 42")
	}
	if Int("TEST_INT_INVALID", 7) != 7 {
		t.Error("expected TEST_INT_INVALID to fallback to 7")
	}
	if Int("TEST_INT_MISSING", 9) != 9 {
		t.Error("expected TEST_INT_MISSING to fallback to 9")
	}
}

func TestMustBool(t *testing.T) {
	t.Setenv("TEST_MUST_BOOL_TRUE", "yes")
	if !MustBool("TEST_MUST_BOOL_TRUE") {
		t.Error("expected TEST_MUST_BOOL_TRUE to be true")
	}
}

func TestMustBoolPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for missing bool value")
		}
	}()
	MustBool("TEST_MUST_BOOL_MISSING")
}

func TestMustBoolEmptyPanic(t *testing.T) {
	t.Setenv("TEST_MUST_BOOL_EMPTY", "")
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for empty bool value")
		}
	}()
	MustBool("TEST_MUST_BOOL_EMPTY")
}

func TestMustInt(t *testing.T) {
	t.Setenv("TEST_MUST_INT_VALID", "123")
	if MustInt("TEST_MUST_INT_VALID") != 123 {
		t.Error("expected TEST_MUST_INT_VALID to be 123")
	}
}

func TestMustIntPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for missing int value")
		}
	}()
	MustInt("TEST_MUST_INT_MISSING")
}

func TestMustIntEmptyPanic(t *testing.T) {
	t.Setenv("TEST_MUST_INT_EMPTY", "")
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for empty int value")
		}
	}()
	MustInt("TEST_MUST_INT_EMPTY")
}

func TestMustIntInvalidPanic(t *testing.T) {
	t.Setenv("TEST_MUST_INT_INVALID", "x")
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid int value")
		}
	}()
	MustInt("TEST_MUST_INT_INVALID")
}

func TestMustString(t *testing.T) {
	t.Setenv("TEST_MUST_STRING", "value")
	if MustString("TEST_MUST_STRING") != "value" {
		t.Error("expected TEST_MUST_STRING to be 'value'")
	}
}

func TestMustStringPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for missing string value")
		}
	}()
	MustString("TEST_MUST_STRING_MISSING")
}

func TestMustStringEmptyPanic(t *testing.T) {
	t.Setenv("TEST_MUST_STRING_EMPTY", "")
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for empty string value")
		}
	}()
	MustString("TEST_MUST_STRING_EMPTY")
}

func TestMustStrings(t *testing.T) {
	t.Setenv("TEST_MUST_STRINGS", " a, b , ,c ")
	got := MustStrings("TEST_MUST_STRINGS")
	want := []string{"a", "b", "c"}
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestMustStringsPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for missing strings value")
		}
	}()
	MustStrings("TEST_MUST_STRINGS_MISSING")
}

func TestMustStringsEmptyPanic(t *testing.T) {
	t.Setenv("TEST_MUST_STRINGS_EMPTY", "")
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for empty strings value")
		}
	}()
	MustStrings("TEST_MUST_STRINGS_EMPTY")
}

func TestString(t *testing.T) {
	t.Setenv("TEST_STRING", "hello")
	if String("TEST_STRING", "fallback") != "hello" {
		t.Error("expected TEST_STRING to be 'hello'")
	}
	if String("TEST_STRING_MISSING", "fallback") != "fallback" {
		t.Error("expected TEST_STRING_MISSING to fallback to 'fallback'")
	}
}

func TestStrings(t *testing.T) {
	t.Setenv("TEST_STRINGS", " one, two ,, three ")
	got := Strings("TEST_STRINGS", []string{"fallback"})
	want := []string{"one", "two", "three"}
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("expected %v, got %v", want, got)
	}

	fallback := []string{"fallback", "values"}
	got = Strings("TEST_STRINGS_MISSING", fallback)
	if !reflect.DeepEqual(fallback, got) {
		t.Fatalf("expected %v, got %v", fallback, got)
	}
}

func TestParseBool(t *testing.T) {
	trueCases := []string{"1", "true", "TRUE", "True", "yes", "YES", "on", "ON"}
	for _, v := range trueCases {
		if !parseBool(v) {
			t.Errorf("expected parseBool(%q) to be true", v)
		}
	}

	falseCases := []string{"0", "false", "False", "no", "off", "", "random"}
	for _, v := range falseCases {
		if parseBool(v) {
			t.Errorf("expected parseBool(%q) to be false", v)
		}
	}
}

func TestSplitAndTrim(t *testing.T) {
	got := splitAndTrim(" a, b ,, c , ", ",")
	want := []string{"a", "b", "c"}
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("expected %v, got %v", want, got)
	}

	got = splitAndTrim("", ",")
	if !reflect.DeepEqual([]string{}, got) {
		t.Fatalf("expected %v, got %v", []string{}, got)
	}
}
