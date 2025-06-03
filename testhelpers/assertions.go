package testhelpers

import (
	"strings"
)

// TestingT is an interface that both testing.T and our mock can implement
type TestingT interface {
	Helper()
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
}

// Helper functions to replace testify functionality with better error messages

func AssertEqual(t TestingT, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			if len(msgAndArgs) > 0 {
				t.Errorf("AssertEqual panicked: %v. Message: %v", r, msgAndArgs[0])
			} else {
				t.Errorf("AssertEqual panicked: %v", r)
			}
		}
	}()
	
	if expected != actual {
		if len(msgAndArgs) > 0 {
			t.Errorf("Expected %v, got %v. Message: %v", expected, actual, msgAndArgs[0])
		} else {
			t.Errorf("Expected %v, got %v", expected, actual)
		}
	}
}

func AssertNotEqual(t TestingT, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			if len(msgAndArgs) > 0 {
				t.Errorf("AssertNotEqual panicked: %v. Message: %v", r, msgAndArgs[0])
			} else {
				t.Errorf("AssertNotEqual panicked: %v", r)
			}
		}
	}()
	
	if expected == actual {
		if len(msgAndArgs) > 0 {
			t.Errorf("Expected values to be different, but both were %v. %v", expected, msgAndArgs[0])
		} else {
			t.Errorf("Expected values to be different, but both were %v", expected)
		}
	}
}

func AssertNotNil(t TestingT, object interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			if len(msgAndArgs) > 0 {
				t.Errorf("AssertNotNil panicked: %v. Message: %v", r, msgAndArgs[0])
			} else {
				t.Errorf("AssertNotNil panicked: %v", r)
			}
		}
	}()
	
	if object == nil {
		if len(msgAndArgs) > 0 {
			t.Errorf("Expected non-nil value. %v", msgAndArgs[0])
		} else {
			t.Errorf("Expected non-nil value")
		}
	}
}

func AssertNil(t TestingT, object interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if object != nil {
		if len(msgAndArgs) > 0 {
			t.Errorf("Expected nil value, got %v. %v", object, msgAndArgs[0])
		} else {
			t.Errorf("Expected nil value, got %v", object)
		}
	}
}

func AssertNoError(t TestingT, err error, msgAndArgs ...interface{}) {
	t.Helper()
	if err != nil {
		if len(msgAndArgs) > 0 {
			t.Errorf("Expected no error, got %v. %v", err, msgAndArgs[0])
		} else {
			t.Errorf("Expected no error, got %v", err)
		}
	}
}

func AssertError(t TestingT, err error, msgAndArgs ...interface{}) {
	t.Helper()
	if err == nil {
		if len(msgAndArgs) > 0 {
			t.Errorf("Expected error, got nil. %v", msgAndArgs[0])
		} else {
			t.Errorf("Expected error, got nil")
		}
	}
}

func AssertContains(t TestingT, s, substr string, msgAndArgs ...interface{}) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			if len(msgAndArgs) > 0 {
				t.Errorf("AssertContains panicked: %v. Message: %v", r, msgAndArgs[0])
			} else {
				t.Errorf("AssertContains panicked: %v", r)
			}
		}
	}()
	
	if !strings.Contains(s, substr) {
		if len(msgAndArgs) > 0 {
			t.Errorf("Expected string to contain %q, but it didn't. String: %q. %v", substr, s, msgAndArgs[0])
		} else {
			t.Errorf("Expected string to contain %q, but it didn't. String: %q", substr, s)
		}
	}
}

func AssertNotContains(t TestingT, s, substr string, msgAndArgs ...interface{}) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			if len(msgAndArgs) > 0 {
				t.Errorf("AssertNotContains panicked: %v. Message: %v", r, msgAndArgs[0])
			} else {
				t.Errorf("AssertNotContains panicked: %v", r)
			}
		}
	}()
	
	if strings.Contains(s, substr) {
		if len(msgAndArgs) > 0 {
			t.Errorf("Expected string to not contain %q, but it did. String: %q. %v", substr, s, msgAndArgs[0])
		} else {
			t.Errorf("Expected string to not contain %q, but it did. String: %q", substr, s)
		}
	}
}

func RequireNoError(t TestingT, err error, msgAndArgs ...interface{}) {
	t.Helper()
	if err != nil {
		if len(msgAndArgs) > 0 {
			t.Fatalf("Expected no error, got %v. %v", err, msgAndArgs[0])
		} else {
			t.Fatalf("Expected no error, got %v", err)
		}
	}
}

func RequireError(t TestingT, err error, msgAndArgs ...interface{}) {
	t.Helper()
	if err == nil {
		if len(msgAndArgs) > 0 {
			t.Fatalf("Expected error, got nil. %v", msgAndArgs[0])
		} else {
			t.Fatalf("Expected error, got nil")
		}
	}
}

func AssertTrue(t TestingT, value bool, msgAndArgs ...interface{}) {
	t.Helper()
	if !value {
		if len(msgAndArgs) > 0 {
			t.Errorf("Expected true, got false. %v", msgAndArgs[0])
		} else {
			t.Errorf("Expected true, got false")
		}
	}
}

func AssertFalse(t TestingT, value bool, msgAndArgs ...interface{}) {
	t.Helper()
	if value {
		if len(msgAndArgs) > 0 {
			t.Errorf("Expected false, got true. %v", msgAndArgs[0])
		} else {
			t.Errorf("Expected false, got true")
		}
	}
}

func AssertLen(t TestingT, object interface{}, expectedLen int, msgAndArgs ...interface{}) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			if len(msgAndArgs) > 0 {
				t.Errorf("AssertLen panicked: %v. Message: %v", r, msgAndArgs[0])
			} else {
				t.Errorf("AssertLen panicked: %v", r)
			}
		}
	}()
	
	var actualLen int
	
	switch v := object.(type) {
	case string:
		actualLen = len(v)
	case []interface{}:
		actualLen = len(v)
	case []string:
		actualLen = len(v)
	case []int:
		actualLen = len(v)
	case map[string]interface{}:
		actualLen = len(v)
	default:
		if len(msgAndArgs) > 0 {
			t.Errorf("Cannot determine length of type %T. %v", object, msgAndArgs[0])
		} else {
			t.Errorf("Cannot determine length of type %T", object)
		}
		return
	}
	
	if actualLen != expectedLen {
		if len(msgAndArgs) > 0 {
			t.Errorf("Expected length %d, got %d. %v", expectedLen, actualLen, msgAndArgs[0])
		} else {
			t.Errorf("Expected length %d, got %d", expectedLen, actualLen)
		}
	}
}

func AssertEmpty(t TestingT, object interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	AssertLen(t, object, 0, msgAndArgs...)
}

func AssertNotEmpty(t TestingT, object interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	var actualLen int
	
	switch v := object.(type) {
	case string:
		actualLen = len(v)
	case []interface{}:
		actualLen = len(v)
	case []string:
		actualLen = len(v)
	case []int:
		actualLen = len(v)
	case map[string]interface{}:
		actualLen = len(v)
	default:
		if len(msgAndArgs) > 0 {
			t.Errorf("Cannot determine length of type %T. %v", object, msgAndArgs[0])
		} else {
			t.Errorf("Cannot determine length of type %T", object)
		}
		return
	}
	
	if actualLen == 0 {
		if len(msgAndArgs) > 0 {
			t.Errorf("Expected non-empty, got empty. %v", msgAndArgs[0])
		} else {
			t.Errorf("Expected non-empty, got empty")
		}
	}
}