package plugins

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type RequestLoggerTestSuite struct {
	suite.Suite
}

func TestRequestLoggerTestSuite(t *testing.T) {
	suite.Run(t, new(RequestLoggerTestSuite))
}

// newTestRequest is a helper to create a new HTTP request for testing.
func newTestRequest(method, urlStr string, body io.Reader) *http.Request {
	req, _ := http.NewRequest(method, urlStr, body)
	return req
}

// extractDuration extracts the duration in milliseconds from a log string.
// Example log: " ... [50ms] ..."
func extractDuration(log string) (int64, error) {
	r := regexp.MustCompile(`\[(\d+)ms\]`)
	matches := r.FindStringSubmatch(log)
	if len(matches) < 2 {
		return 0, fmt.Errorf("duration not found in log: %s", log)
	}
	duration, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("could not parse duration '%s': %w", matches[1], err)
	}
	return duration, nil
}

func (s *RequestLoggerTestSuite) TestNewRequestLogger() {
	s.Run("nil_outputs_default_to_os_streams", func() {
		plugin := NewRequestLogger(nil, nil)
		s.Require().NotNil(plugin, "Plugin should not be nil")

		rl, ok := plugin.(*requestLogger)
		s.Require().True(ok, "Plugin should be of type *requestLogger")
		assert.Equal(s.T(), os.Stdout, rl.out, "Default out writer should be os.Stdout")
		assert.Equal(s.T(), os.Stderr, rl.errOut, "Default errOut writer should be os.Stderr")
	})

	s.Run("provided_outputs_are_used", func() {
		var myOut bytes.Buffer
		var myErrOut bytes.Buffer
		plugin := NewRequestLogger(&myOut, &myErrOut)
		s.Require().NotNil(plugin, "Plugin should not be nil")

		rl, ok := plugin.(*requestLogger)
		s.Require().True(ok, "Plugin should be of type *requestLogger")
		assert.Equal(s.T(), &myOut, rl.out, "Provided out writer was not used")
		assert.Equal(s.T(), &myErrOut, rl.errOut, "Provided errOut writer was not used")
	})
}

func (s *RequestLoggerTestSuite) TestRequestLogger_OnRequestStart() {
	logger := NewRequestLogger(nil, nil)
	req := newTestRequest(http.MethodGet, "http://example.com", nil)

	beforeTime := time.Now()
	logger.OnRequestStart(req)
	afterTime := time.Now()

	ctxValue := req.Context().Value(reqTime)
	s.Require().NotNil(ctxValue, "reqTime should be set in context")

	startTime, ok := ctxValue.(time.Time)
	s.Require().True(ok, "reqTime should be of type time.Time")

	assert.True(s.T(), !startTime.Before(beforeTime) || startTime.Equal(beforeTime), "startTime should be >= beforeTime")
	assert.True(s.T(), !startTime.After(afterTime) || startTime.Equal(afterTime), "startTime should be <= afterTime")
}

func (s *RequestLoggerTestSuite) TestRequestLogger_OnRequestEnd() {
	var outBuf bytes.Buffer
	loggerPlugin := NewRequestLogger(&outBuf, nil)
	// We need the concrete type to call methods directly if the interface doesn't expose them,
	// but heimdall.Plugin interface methods are what we test.

	req := newTestRequest(http.MethodGet, "http://example.com/test", nil)
	res := httptest.NewRecorder() // Using httptest.ResponseRecorder as a simple http.Response
	res.Code = http.StatusOK

	s.Run("logs_with_duration_when_time_in_context", func() {
		outBuf.Reset()
		loggerPlugin.OnRequestStart(req)
		// Ensure some time passes for duration calculation
		time.Sleep(50 * time.Millisecond)
		loggerPlugin.OnRequestEnd(req, res.Result())

		logOutput := outBuf.String()
		s.T().Logf("OnRequestEnd output: %s", logOutput)
		assert.Contains(s.T(), logOutput, http.MethodGet, "Log should contain HTTP method")
		assert.Contains(s.T(), logOutput, req.URL.String(), "Log should contain URL")
		assert.Contains(s.T(), logOutput, strconv.Itoa(http.StatusOK), "Log should contain status code")

		durationMs, err := extractDuration(logOutput)
		require.NoError(s.T(), err, "Failed to extract duration from log")
		assert.InDelta(s.T(), 50, durationMs, 25, "Duration should be around 50ms (+/- 25ms margin)")
	})

	s.Run("logs_0ms_duration_when_time_not_in_context", func() {
		outBuf.Reset()
		// Create a new request without OnRequestStart being called on it by our logger
		freshReq := newTestRequest(http.MethodPost, "http://example.com/notime", nil)
		loggerPlugin.OnRequestEnd(freshReq, res.Result())

		logOutput := outBuf.String()
		s.T().Logf("OnRequestEnd (no time context) output: %s", logOutput)
		assert.Contains(s.T(), logOutput, "[0ms]", "Log should show 0ms duration")
	})

	s.Run("logs_0ms_duration_when_time_is_wrong_type", func() {
		outBuf.Reset()
		wrongTypeCtx := context.WithValue(req.Context(), reqTime, "not a time.Time instance")
		reqWithWrongCtx := req.WithContext(wrongTypeCtx)
		loggerPlugin.OnRequestEnd(reqWithWrongCtx, res.Result())

		logOutput := outBuf.String()
		s.T().Logf("OnRequestEnd (wrong time type) output: %s", logOutput)
		assert.Contains(s.T(), logOutput, "[0ms]", "Log should show 0ms duration for wrong type")
	})
}

func (s *RequestLoggerTestSuite) TestRequestLogger_OnError() {
	var errBuf bytes.Buffer
	loggerPlugin := NewRequestLogger(nil, &errBuf)
	req := newTestRequest(http.MethodPut, "http://example.com/error", nil)
	testError := errors.New("simulated network error")

	s.Run("logs_error_with_duration_when_time_in_context", func() {
		errBuf.Reset()
		loggerPlugin.OnRequestStart(req)
		time.Sleep(60 * time.Millisecond) // Simulate work/delay
		loggerPlugin.OnError(req, testError)

		logOutput := errBuf.String()
		s.T().Logf("OnError output: %s", logOutput)
		assert.Contains(s.T(), logOutput, http.MethodPut, "Log should contain HTTP method")
		assert.Contains(s.T(), logOutput, req.URL.String(), "Log should contain URL")
		assert.Contains(s.T(), logOutput, "ERROR: "+testError.Error(), "Log should contain error message")

		durationMs, err := extractDuration(logOutput)
		require.NoError(s.T(), err, "Failed to extract duration from log")
		assert.InDelta(s.T(), 60, durationMs, 25, "Duration should be around 60ms (+/- 25ms margin)")
	})

	s.Run("logs_0ms_duration_when_time_not_in_context", func() {
		errBuf.Reset()
		freshReq := newTestRequest(http.MethodDelete, "http://example.com/errortime", nil)
		loggerPlugin.OnError(freshReq, testError)

		logOutput := errBuf.String()
		s.T().Logf("OnError (no time context) output: %s", logOutput)
		assert.Contains(s.T(), logOutput, "[0ms]", "Log should show 0ms duration")
	})

	s.Run("logs_0ms_duration_when_time_is_wrong_type", func() {
		errBuf.Reset()
		wrongTypeCtx := context.WithValue(req.Context(), reqTime, 12345) // int, not time.Time
		reqWithWrongCtx := req.WithContext(wrongTypeCtx)
		loggerPlugin.OnError(reqWithWrongCtx, testError)

		logOutput := errBuf.String()
		s.T().Logf("OnError (wrong time type) output: %s", logOutput)
		assert.Contains(s.T(), logOutput, "[0ms]", "Log should show 0ms duration for wrong type")
	})
}

func (s *RequestLoggerTestSuite) TestGetRequestDuration() {
	s.Run("valid_time_in_context", func() {
		startTime := time.Now().Add(-100 * time.Millisecond) // Known start time
		ctx := context.WithValue(context.Background(), reqTime, startTime)

		// To make this test deterministic for getRequestDuration's internal time.Now() call,
		// we'd ideally mock time.Now(). Since we can't easily do that without a library
		// or interface injection for time.Now(), we test that the subtraction is correct.
		// The duration will be time.Now() (when getRequestDuration is called) - startTime.
		// We expect it to be slightly more than 100ms due to execution time.

		calculatedDuration := getRequestDuration(ctx)

		// Assert that the duration is at least 100ms and not excessively large.
		assert.GreaterOrEqual(s.T(), calculatedDuration, 100*time.Millisecond)
		assert.LessOrEqual(s.T(), calculatedDuration, 150*time.Millisecond, "Duration calculation seems too high, check test logic or system load")
	})

	s.Run("no_time_in_context", func() {
		ctx := context.Background()
		duration := getRequestDuration(ctx)
		assert.Equal(s.T(), time.Duration(0), duration, "Duration should be 0 if reqTime not in context")
	})

	s.Run("wrong_type_in_context", func() {
		ctx := context.WithValue(context.Background(), reqTime, "not a time.Time")
		duration := getRequestDuration(ctx)
		assert.Equal(s.T(), time.Duration(0), duration, "Duration should be 0 if reqTime is of wrong type")
	})
}
