# Go Project

A zero dependency project starter template with HTTP server for Go.

## Features

- simple project structure
- zero dependency
- HTTP server
  - `net/http` compatible
  - middleware support
  - centralized error handling

## Assertion Helpers

The `service/assert` package provides a set of assertion functions for validation purposes. These functions help ensure that certain conditions are met, and if not, they trigger a panic with message and stack trace.

Example:

```go
var b *App
// panics with message: `assertion failed: value is nil`
assert.NotNil(b)
// or with formatted message, panics with message:
// `assertion failed: value is nil: instance should not be "nil"`
assert.NotNil(b, "instance should not be %q", "nil")
```

- `Empty(s string)` — asserts that the given string is empty
- `Equal[T any](expected, actual T)` — asserts that the expected and actual values are equal (`reflect.DeepEqual`)
- `Error(actual, expected error)` — asserts that the actual error matches the expected error (same instance, wrapped, or same message)
- `False(condition bool)` — asserts that the given condition is `false`
- `Greater[T Ordered](actual, min T)` — asserts that `actual` is greater than `min`
- `GreaterOrEqual[T Ordered](actual, min T)` — asserts that `actual` is greater than or equal to `min`
- `Len(object any, expected int)` — asserts that the given object has the expected length (supports array, slice, map, string, or channel)
- `Less[T Ordered](actual, max T)` — asserts that `actual` is less than `max`
- `LessOrEqual[T Ordered](actual, max T)` — asserts that `actual` is less than or equal to `max`
- `Nil(v any)` — asserts that the given value is `nil` (works for pointers, maps, slices, interfaces, etc.)
- `NoError(err error)` — asserts that the given error is `nil`
- `NotEmpty(s string)` — asserts that the given string is not empty
- `NotEqual[T any](expected, actual T)` — asserts that the expected and actual values are not equal
- `NotNil(v any)` — asserts that the given value is not `nil`
- `True(condition bool)` — asserts that the given condition is `true`
- `Type(a, b any)` — asserts that the types of `a` and `b` are the same

The `service/assert/test` package provides a set of assertion functions for testing purposes. These functions help ensure that certain conditions are met during tests, and if not, they trigger `t.Fatal` with a message and stack trace.

Example:

```go
func TestApp(t *testing.T) {
	var app *App
    // fails the test with message: `assertion failed: value is nil`
	test.NotNil(t, app)
    // or with formatted message, fails the test with message:
    // `assertion failed: value is nil: app should not be "nil"`
	test.NotNil(t, app, "app should not be %q", "nil")
}
```

- `Contains(t TestingT, haystack, needle any)` — asserts that `haystack` contains `needle` (supports strings, slices, arrays, and maps)
- `Empty(t TestingT, s string)` — asserts that the given string is empty
- `Equal[T any](t TestingT, expected, actual T)` — asserts that the expected and actual values are equal (`reflect.DeepEqual`)
- `Error(t TestingT, actual, expected error)` — asserts that the actual error matches the expected error (same instance, wrapped, or same message)
- `False(t TestingT, condition bool)` — asserts that the given condition is `false`
- `Greater[T Ordered](t TestingT, actual, min T)` — asserts that `actual` is greater than `min`
- `GreaterOrEqual[T Ordered](t TestingT, actual, min T)` — asserts that `actual` is greater than or equal to `min`
- `Len(t TestingT, object any, expected int)` — asserts that the given object has the expected length (supports array, slice, map, string, or channel)
- `Less[T Ordered](t TestingT, actual, max T)` — asserts that `actual` is less than `max`
- `LessOrEqual[T Ordered](t TestingT, actual, max T)` — asserts that `actual` is less than or equal to `max`
- `Nil(t TestingT, v any)` — asserts that the given value is `nil` (works for pointers, maps, slices, interfaces, etc.)
- `NoError(t TestingT, err error)` — asserts that the given error is `nil`
- `NotEmpty(t TestingT, s string)` — asserts that the given string is not empty
- `NotEqual[T any](t TestingT, expected, actual T)` — asserts that the expected and actual values are not equal
- `NotNil(t TestingT, v any)` — asserts that the given value is not `nil`
- `Panics(t TestingT, f func())` — asserts that the given function panics when called
- `True(t TestingT, condition bool)` — asserts that the given condition is `true`
- `Type(t TestingT, a, b any)` — asserts that the types of `a` and `b` are the same
