package testhelpers

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// mockT is a mock testing.T to capture test failures
type mockT struct {
	errors []string
	fatals []string
}

func (m *mockT) Helper() {}

func (m *mockT) Errorf(format string, args ...interface{}) {
	m.errors = append(m.errors, "ERROR: "+fmt.Sprintf(format, args...))
}

func (m *mockT) Fatalf(format string, args ...interface{}) {
	m.fatals = append(m.fatals, "FATAL: "+fmt.Sprintf(format, args...))
}

func (m *mockT) hasError() bool {
	return len(m.errors) > 0
}

func (m *mockT) hasFatal() bool {
	return len(m.fatals) > 0
}

func (m *mockT) reset() {
	m.errors = nil
	m.fatals = nil
}

func TestAssertEqual(t *testing.T) {
	mock := &mockT{}

	// Test success case
	AssertEqual(mock, "hello", "hello")
	if mock.hasError() {
		t.Errorf("AssertEqual should not fail when values are equal")
	}

	// Test failure case
	mock.reset()
	AssertEqual(mock, "hello", "world")
	if !mock.hasError() {
		t.Errorf("AssertEqual should fail when values are different")
	}

	// Test with message
	mock.reset()
	AssertEqual(mock, 1, 2, "custom message")
	if !mock.hasError() {
		t.Errorf("AssertEqual should fail when values are different")
	}
	if len(mock.errors) > 0 && !strings.Contains(mock.errors[0], "Message: custom message") {
		t.Errorf("AssertEqual should include custom message in error. Got: %s", mock.errors[0])
	}
}

func TestAssertNotEqual(t *testing.T) {
	mock := &mockT{}

	// Test success case
	AssertNotEqual(mock, "hello", "world")
	if mock.hasError() {
		t.Errorf("AssertNotEqual should not fail when values are different")
	}

	// Test failure case
	mock.reset()
	AssertNotEqual(mock, "hello", "hello")
	if !mock.hasError() {
		t.Errorf("AssertNotEqual should fail when values are equal")
	}
}

func TestAssertNotNil(t *testing.T) {
	mock := &mockT{}

	// Test success case
	AssertNotNil(mock, "not nil")
	if mock.hasError() {
		t.Errorf("AssertNotNil should not fail when value is not nil")
	}

	// Test failure case
	mock.reset()
	AssertNotNil(mock, nil)
	if !mock.hasError() {
		t.Errorf("AssertNotNil should fail when value is nil")
	}
}

func TestAssertNil(t *testing.T) {
	mock := &mockT{}

	// Test success case
	AssertNil(mock, nil)
	if mock.hasError() {
		t.Errorf("AssertNil should not fail when value is nil")
	}

	// Test failure case
	mock.reset()
	AssertNil(mock, "not nil")
	if !mock.hasError() {
		t.Errorf("AssertNil should fail when value is not nil")
	}
}

func TestAssertNoError(t *testing.T) {
	mock := &mockT{}

	// Test success case
	AssertNoError(mock, nil)
	if mock.hasError() {
		t.Errorf("AssertNoError should not fail when error is nil")
	}

	// Test failure case
	mock.reset()
	AssertNoError(mock, errors.New("test error"))
	if !mock.hasError() {
		t.Errorf("AssertNoError should fail when error is not nil")
	}
}

func TestAssertError(t *testing.T) {
	mock := &mockT{}

	// Test success case
	AssertError(mock, errors.New("test error"))
	if mock.hasError() {
		t.Errorf("AssertError should not fail when error is not nil")
	}

	// Test failure case
	mock.reset()
	AssertError(mock, nil)
	if !mock.hasError() {
		t.Errorf("AssertError should fail when error is nil")
	}
}

func TestAssertContains(t *testing.T) {
	mock := &mockT{}

	// Test success case
	AssertContains(mock, "hello world", "world")
	if mock.hasError() {
		t.Errorf("AssertContains should not fail when string contains substring")
	}

	// Test failure case
	mock.reset()
	AssertContains(mock, "hello world", "xyz")
	if !mock.hasError() {
		t.Errorf("AssertContains should fail when string does not contain substring")
	}
}

func TestAssertNotContains(t *testing.T) {
	mock := &mockT{}

	// Test success case
	AssertNotContains(mock, "hello world", "xyz")
	if mock.hasError() {
		t.Errorf("AssertNotContains should not fail when string does not contain substring")
	}

	// Test failure case
	mock.reset()
	AssertNotContains(mock, "hello world", "world")
	if !mock.hasError() {
		t.Errorf("AssertNotContains should fail when string contains substring")
	}
}

func TestRequireNoError(t *testing.T) {
	mock := &mockT{}

	// Test success case
	RequireNoError(mock, nil)
	if mock.hasFatal() {
		t.Errorf("RequireNoError should not fatal when error is nil")
	}

	// Test failure case
	mock.reset()
	RequireNoError(mock, errors.New("test error"))
	if !mock.hasFatal() {
		t.Errorf("RequireNoError should fatal when error is not nil")
	}
}

func TestRequireError(t *testing.T) {
	mock := &mockT{}

	// Test success case
	RequireError(mock, errors.New("test error"))
	if mock.hasFatal() {
		t.Errorf("RequireError should not fatal when error is not nil")
	}

	// Test failure case
	mock.reset()
	RequireError(mock, nil)
	if !mock.hasFatal() {
		t.Errorf("RequireError should fatal when error is nil")
	}
}

func TestAssertTrue(t *testing.T) {
	mock := &mockT{}

	// Test success case
	AssertTrue(mock, true)
	if mock.hasError() {
		t.Errorf("AssertTrue should not fail when value is true")
	}

	// Test failure case
	mock.reset()
	AssertTrue(mock, false)
	if !mock.hasError() {
		t.Errorf("AssertTrue should fail when value is false")
	}
}

func TestAssertFalse(t *testing.T) {
	mock := &mockT{}

	// Test success case
	AssertFalse(mock, false)
	if mock.hasError() {
		t.Errorf("AssertFalse should not fail when value is false")
	}

	// Test failure case
	mock.reset()
	AssertFalse(mock, true)
	if !mock.hasError() {
		t.Errorf("AssertFalse should fail when value is true")
	}
}

func TestAssertLen(t *testing.T) {
	mock := &mockT{}

	// Test success cases
	AssertLen(mock, "hello", 5)
	if mock.hasError() {
		t.Errorf("AssertLen should not fail for correct string length")
	}

	mock.reset()
	AssertLen(mock, []string{"a", "b", "c"}, 3)
	if mock.hasError() {
		t.Errorf("AssertLen should not fail for correct slice length")
	}

	mock.reset()
	AssertLen(mock, map[string]interface{}{"a": 1, "b": 2}, 2)
	if mock.hasError() {
		t.Errorf("AssertLen should not fail for correct map length")
	}

	// Test failure case
	mock.reset()
	AssertLen(mock, "hello", 3)
	if !mock.hasError() {
		t.Errorf("AssertLen should fail for incorrect length")
	}

	// Test unsupported type
	mock.reset()
	AssertLen(mock, 123, 3)
	if !mock.hasError() {
		t.Errorf("AssertLen should fail for unsupported type")
	}
}

func TestAssertEmpty(t *testing.T) {
	mock := &mockT{}

	// Test success cases
	AssertEmpty(mock, "")
	if mock.hasError() {
		t.Errorf("AssertEmpty should not fail for empty string")
	}

	mock.reset()
	AssertEmpty(mock, []string{})
	if mock.hasError() {
		t.Errorf("AssertEmpty should not fail for empty slice")
	}

	// Test failure case
	mock.reset()
	AssertEmpty(mock, "hello")
	if !mock.hasError() {
		t.Errorf("AssertEmpty should fail for non-empty string")
	}
}

func TestAssertNotEmpty(t *testing.T) {
	mock := &mockT{}

	// Test success case
	AssertNotEmpty(mock, "hello")
	if mock.hasError() {
		t.Errorf("AssertNotEmpty should not fail for non-empty string")
	}

	mock.reset()
	AssertNotEmpty(mock, []string{"a"})
	if mock.hasError() {
		t.Errorf("AssertNotEmpty should not fail for non-empty slice")
	}

	// Test failure case
	mock.reset()
	AssertNotEmpty(mock, "")
	if !mock.hasError() {
		t.Errorf("AssertNotEmpty should fail for empty string")
	}

	mock.reset()
	AssertNotEmpty(mock, []string{})
	if !mock.hasError() {
		t.Errorf("AssertNotEmpty should fail for empty slice")
	}
}

func TestAssertLenWithDifferentTypes(t *testing.T) {
	mock := &mockT{}

	// Test different slice types
	AssertLen(mock, []int{1, 2, 3}, 3)
	if mock.hasError() {
		t.Errorf("AssertLen should work with []int")
	}

	mock.reset()
	AssertLen(mock, []interface{}{1, "a", true}, 3)
	if mock.hasError() {
		t.Errorf("AssertLen should work with []interface{}")
	}
}