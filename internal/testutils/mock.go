package testutils

import (
	"fmt"
	"sync"
)

// MockCall represents a method call on a mock object
type MockCall struct {
	Method string
	Args   []interface{}
	Return []interface{}
}

// Mock provides a simple mocking framework for tests
type Mock struct {
	mu           sync.RWMutex
	calls        []MockCall
	expectations map[string][]MockCall
	strict       bool
	callIndex    map[string]int
}

// NewMock creates a new mock object
func NewMock() *Mock {
	return &Mock{
		calls:        make([]MockCall, 0),
		expectations: make(map[string][]MockCall),
		callIndex:    make(map[string]int),
	}
}

// NewStrictMock creates a mock that validates call order
func NewStrictMock() *Mock {
	mock := NewMock()
	mock.strict = true
	return mock
}

// Expect sets up an expectation for a method call
func (m *Mock) Expect(method string, args ...interface{}) *MockExpectation {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.expectations[method] == nil {
		m.expectations[method] = make([]MockCall, 0)
	}

	return &MockExpectation{
		mock:   m,
		method: method,
		args:   args,
	}
}

// RecordCall records a method call and returns expected result
func (m *Mock) RecordCall(method string, args ...interface{}) []interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	call := MockCall{
		Method: method,
		Args:   args,
	}
	m.calls = append(m.calls, call)

	// Find matching expectation
	if expectations, exists := m.expectations[method]; exists {
		index := m.callIndex[method]
		if index < len(expectations) {
			expected := expectations[index]
			m.callIndex[method]++
			call.Return = expected.Return
			return expected.Return
		}
	}

	// No expectation found
	return nil
}

// Verify checks that all expectations were met
func (m *Mock) Verify() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for method, expectations := range m.expectations {
		called := m.callIndex[method]
		expected := len(expectations)
		if called != expected {
			return fmt.Errorf("method %s: expected %d calls, got %d", method, expected, called)
		}
	}

	return nil
}

// GetCalls returns all recorded calls
func (m *Mock) GetCalls() []MockCall {
	m.mu.RLock()
	defer m.mu.RUnlock()

	calls := make([]MockCall, len(m.calls))
	copy(calls, m.calls)
	return calls
}

// GetCallsFor returns calls for a specific method
func (m *Mock) GetCallsFor(method string) []MockCall {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var methodCalls []MockCall
	for _, call := range m.calls {
		if call.Method == method {
			methodCalls = append(methodCalls, call)
		}
	}
	return methodCalls
}

// Reset clears all calls and expectations
func (m *Mock) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.calls = m.calls[:0]
	m.expectations = make(map[string][]MockCall)
	m.callIndex = make(map[string]int)
}

// MockExpectation allows method chaining for setting up expectations
type MockExpectation struct {
	mock   *Mock
	method string
	args   []interface{}
}

// Return sets the return values for the expectation
func (e *MockExpectation) Return(values ...interface{}) *MockExpectation {
	e.mock.mu.Lock()
	defer e.mock.mu.Unlock()

	call := MockCall{
		Method: e.method,
		Args:   e.args,
		Return: values,
	}

	e.mock.expectations[e.method] = append(e.mock.expectations[e.method], call)
	return e
}

// Times specifies how many times the method should be called
func (e *MockExpectation) Times(count int) *MockExpectation {
	// For simplicity, just add the expectation multiple times
	for i := 1; i < count; i++ {
		e.Return(e.mock.expectations[e.method][len(e.mock.expectations[e.method])-1].Return...)
	}
	return e
}

// Example mock implementations for common interfaces

// MockHTTPClient provides a mock HTTP client for testing
type MockHTTPClient struct {
	*Mock
}

// NewMockHTTPClient creates a new mock HTTP client
func NewMockHTTPClient() *MockHTTPClient {
	return &MockHTTPClient{
		Mock: NewMock(),
	}
}

// Get simulates an HTTP GET request
func (m *MockHTTPClient) Get(url string) ([]byte, error) {
	results := m.RecordCall("Get", url)
	if len(results) >= 2 {
		if err, ok := results[1].(error); ok {
			return results[0].([]byte), err
		}
		return results[0].([]byte), nil
	}
	return nil, fmt.Errorf("no expectation set for Get(%s)", url)
}

// MockFileSystem provides a mock file system for testing
type MockFileSystem struct {
	*Mock
	files map[string][]byte
}

// NewMockFileSystem creates a new mock file system
func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		Mock:  NewMock(),
		files: make(map[string][]byte),
	}
}

// WriteFile simulates writing a file
func (m *MockFileSystem) WriteFile(path string, data []byte) error {
	results := m.RecordCall("WriteFile", path, data)
	m.files[path] = data
	if len(results) >= 1 {
		if err, ok := results[0].(error); ok {
			return err
		}
	}
	return nil
}

// ReadFile simulates reading a file
func (m *MockFileSystem) ReadFile(path string) ([]byte, error) {
	results := m.RecordCall("ReadFile", path)
	if len(results) >= 2 {
		if err, ok := results[1].(error); ok {
			return results[0].([]byte), err
		}
		return results[0].([]byte), nil
	}
	if data, exists := m.files[path]; exists {
		return data, nil
	}
	return nil, fmt.Errorf("file not found: %s", path)
}

// FileExists simulates checking if a file exists
func (m *MockFileSystem) FileExists(path string) bool {
	results := m.RecordCall("FileExists", path)
	if len(results) >= 1 {
		if exists, ok := results[0].(bool); ok {
			return exists
		}
	}
	_, exists := m.files[path]
	return exists
}
