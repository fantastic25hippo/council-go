package jules

import (
	"errors"
	"testing"
	"time"
)

type fakeNetError struct {
	timeout   bool
	temporary bool
}

func (e *fakeNetError) Error() string   { return "fake net error" }
func (e *fakeNetError) Timeout() bool   { return e.timeout }
func (e *fakeNetError) Temporary() bool { return e.temporary }

type testDoer struct {
	responses []struct {
		status int
		err    error
	}
	calls int
}

func (t *testDoer) Do(method, path string, body []byte) (int, []byte, error) {
	if t.calls >= len(t.responses) {
		t.calls++
		return 500, nil, errors.New("unexpected call")
	}
	resp := t.responses[t.calls]
	t.calls++
	return resp.status, nil, resp.err
}

func TestWithRetries(t *testing.T) {
	tests := []struct {
		name         string
		responses    []struct{ status int; err error }
		expectedTries int
		expectedStatus int
		expectedErr   bool
	}{
		{
			name: "Success first try",
			responses: []struct{ status int; err error }{
				{200, nil},
			},
			expectedTries: 1,
			expectedStatus: 200,
			expectedErr: false,
		},
		{
			name: "404 retry then success",
			responses: []struct{ status int; err error }{
				{404, nil},
				{404, nil},
				{200, nil},
			},
			expectedTries: 3,
			expectedStatus: 200,
			expectedErr: false,
		},
		{
			name: "404 max retries",
			responses: []struct{ status int; err error }{
				{404, nil},
				{404, nil},
				{404, nil},
				{404, nil},
			},
			expectedTries: 4,
			expectedStatus: 404,
			expectedErr: false,
		},
		{
			name: "Timeout retry then success",
			responses: []struct{ status int; err error }{
				{0, &fakeNetError{timeout: true}},
				{500, nil},
				{200, nil},
			},
			expectedTries: 3,
			expectedStatus: 200,
			expectedErr: false,
		},
		{
			name: "502 max retries",
			responses: []struct{ status int; err error }{
				{502, nil},
				{502, nil},
				{502, nil},
				{502, nil},
			},
			expectedTries: 4,
			expectedStatus: 502,
			expectedErr: false,
		},
		{
			name: "Immediate fail 400",
			responses: []struct{ status int; err error }{
				{400, nil},
			},
			expectedTries: 1,
			expectedStatus: 400,
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inner := &testDoer{responses: tt.responses}
			var sleeps []time.Duration
			sleepFn := func(d time.Duration) { sleeps = append(sleeps, d) }

			doer := WithRetries(inner, sleepFn)
			status, _, err := doer.Do("GET", "/test", nil)

			if inner.calls != tt.expectedTries {
				t.Errorf("expected %d tries, got %d", tt.expectedTries, inner.calls)
			}
			if status != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, status)
			}
			if (err != nil) != tt.expectedErr {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}
