package adsbexchange

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// newClient is a test helper that creates a Client, failing the test on error.
func newClient(t *testing.T, host, key string) *Client {
	t.Helper()
	client, err := NewClient(zap.NewNop(), &Config{Host: host, Key: key})
	require.NoError(t, err)
	return client
}

// newMockClient starts a test HTTP server that calls handler, points the client
// at it, and registers cleanup. Returns the ready-to-use client.
func newMockClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	c := newClient(t, "api.example.com", "test-key-123")
	c.URL = strings.TrimSuffix(srv.URL, "/") + "/v2"
	return c
}

// jsonHandler returns an http.HandlerFunc that encodes v as JSON.
func jsonHandler(v any) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(v)
	}
}

// ---------------------------------------------------------------------------
// NewClient
// ---------------------------------------------------------------------------

func TestNewClient(t *testing.T) {
	client, err := NewClient(zap.NewNop(), &Config{Host: "api.example.com", Key: "test-key-123"})

	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "api.example.com", client.Host)
	assert.Equal(t, "test-key-123", client.Key)
	assert.Equal(t, "https://api.example.com/v2", client.URL)
}

func TestNewClientMissingHost(t *testing.T) {
	client, err := NewClient(zap.NewNop(), &Config{Host: "", Key: "test-key"})

	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Equal(t, "adsbexchange host and key must be set", err.Error())
}

func TestNewClientMissingKey(t *testing.T) {
	client, err := NewClient(zap.NewNop(), &Config{Host: "api.example.com", Key: ""})

	assert.Error(t, err)
	assert.Nil(t, client)
}

// ---------------------------------------------------------------------------
// trimWhitespace
// ---------------------------------------------------------------------------

func TestTrimWhitespace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"trim leading and trailing spaces", "  hello world  ", "hello world"},
		{"no trimming needed", "hello", "hello"},
		{"empty string", "", ""},
		{"only spaces", "   ", ""},
		{"trim tabs and newlines", "\t\nhello\n\t", "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, trimWhitespace(tt.input))
		})
	}
}

// ---------------------------------------------------------------------------
// filterADSBICAO
// ---------------------------------------------------------------------------

func TestFilterADSBICAO(t *testing.T) {
	client := newClient(t, "api.example.com", "test-key-123")

	aircraft := []*Aircraft{
		{RecordType: "adsb_icao", Flight: "  UA123  "},
		{RecordType: "adsb_other", Flight: "AA456"},
		{RecordType: "adsb_icao", Flight: "  DL789  "},
		{RecordType: "adsb_other", Flight: "SW321"},
	}

	filtered := client.filterADSBICAO(aircraft)

	assert.Len(t, filtered, 2)
	assert.Equal(t, "UA123", filtered[0].Flight)
	assert.Equal(t, "DL789", filtered[1].Flight)
}

// TestFilterADSBICAOEmpty tests filterADSBICAO with empty aircraft list
func TestFilterADSBICAOEmpty(t *testing.T) {
	client := newClient(t, "api.example.com", "test-key-123")
	assert.Empty(t, client.filterADSBICAO([]*Aircraft{}))
}

// TestFilterADSBICAONoMatches tests filterADSBICAO with no matching records
func TestFilterADSBICAONoMatches(t *testing.T) {
	client := newClient(t, "api.example.com", "test-key-123")

	aircraft := []*Aircraft{
		{RecordType: "adsb_other", Flight: "AA123"},
		{RecordType: "adsb_other", Flight: "UA456"},
	}

	assert.Empty(t, client.filterADSBICAO(aircraft))
}

// ---------------------------------------------------------------------------
// getRequest
// ---------------------------------------------------------------------------

func TestGetRequest(t *testing.T) {
	client := newClient(t, "api.example.com", "test-key-123")

	req, err := client.getRequest("/lat40.00000/lon/120.00000/dist/50/")

	assert.NoError(t, err)
	assert.NotNil(t, req)
	assert.Equal(t, "GET", req.Method)
	assert.Equal(t, "https://api.example.com/v2/lat40.00000/lon/120.00000/dist/50/", req.URL.String())
	assert.Equal(t, "test-key-123", req.Header.Get(headerRapidAPIKey))
	assert.Equal(t, "api.example.com", req.Header.Get(headerRapidAPIHost))
}

// ---------------------------------------------------------------------------
// doRequest
// ---------------------------------------------------------------------------

func TestDoRequest(t *testing.T) {
	client := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-key-123", r.Header.Get(headerRapidAPIKey))
		assert.Equal(t, "api.example.com", r.Header.Get(headerRapidAPIHost))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"msg":"ok"}`))
	})

	body, err := client.doRequest("/mil/")

	assert.NoError(t, err)
	assert.JSONEq(t, `{"msg":"ok"}`, string(body))
}

// ---------------------------------------------------------------------------
// fetchResult / fetchAircraft
// ---------------------------------------------------------------------------

func TestFetchResult(t *testing.T) {
	result := Result{Aircraft: []*Aircraft{
		{RecordType: "adsb_icao", Flight: "  UA123  "},
		{RecordType: "adsb_other", Flight: "DL456"},
		{RecordType: "mlat", Flight: "AA789"},
	}}
	client := newMockClient(t, jsonHandler(result))

	aircraft, err := client.fetchResult("/mil/")

	assert.NoError(t, err)
	assert.Len(t, aircraft, 2)
	assert.Equal(t, "UA123", aircraft[0].Flight)
	assert.Equal(t, "AA789", aircraft[1].Flight)
}

func TestFetchAircraft(t *testing.T) {
	ac := Aircraft{RecordType: "adsb_icao", Flight: "UA123"}
	client := newMockClient(t, jsonHandler(ac))

	got, err := client.fetchAircraft("/hex/abc123")

	assert.NoError(t, err)
	assert.Equal(t, "UA123", got.Flight)
}

// ---------------------------------------------------------------------------
// GetAircraftWithinRange
// ---------------------------------------------------------------------------

func TestGetAircraftWithinRange(t *testing.T) {
	result := Result{Aircraft: []*Aircraft{
		{RecordType: "adsb_icao", Flight: "  UA123  "},
		{RecordType: "adsb_icao", Flight: "  DL456  "},
		{RecordType: "adsb_other", Flight: "AA789"},
	}}
	client := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-key-123", r.Header.Get(headerRapidAPIKey))
		assert.Equal(t, "api.example.com", r.Header.Get(headerRapidAPIHost))
		jsonHandler(result)(w, r)
	})

	aircraft, err := client.GetAircraftWithinRange(40.0, 120.0, 50)

	assert.NoError(t, err)
	assert.Len(t, aircraft, 2)
	assert.Equal(t, "UA123", aircraft[0].Flight)
	assert.Equal(t, "DL456", aircraft[1].Flight)
}

func TestGetAircraftWithinRangeFiltersNonADSBICAO(t *testing.T) {
	result := Result{Aircraft: []*Aircraft{
		{RecordType: "adsb_icao", Flight: "UA123"},
		{RecordType: "adsb_other", Flight: "DL456"},
	}}
	client := newMockClient(t, jsonHandler(result))

	aircraft, err := client.GetAircraftWithinRange(40.0, 120.0, 50)

	assert.NoError(t, err)
	assert.Len(t, aircraft, 1)
	assert.Equal(t, "UA123", aircraft[0].Flight)
}

func TestGetAircraftWithinRangeEmptyResponse(t *testing.T) {
	client := newMockClient(t, jsonHandler(Result{Aircraft: []*Aircraft{}}))

	aircraft, err := client.GetAircraftWithinRange(40.0, 120.0, 50)

	assert.NoError(t, err)
	assert.Empty(t, aircraft)
}

// ---------------------------------------------------------------------------
// closeResponseBody
// ---------------------------------------------------------------------------

func TestCloseResponseBody(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	body := io.NopCloser(strings.NewReader("test"))

	assert.NotPanics(t, func() {
		closeResponseBody(body, logger)
	})
}
