package registry

import (
	"net/http/httptest"
	"testing"
)

func TestMetricsFormatString(t *testing.T) {
	tests := []struct {
		format   MetricsFormat
		expected string
	}{
		{Prometheus_0_0_4, "prometheus-0.0.4"},
		{OpenMetrics_1_0_0, "openmetrics-1.0.0"},
		{MetricsFormat(999), "unknown"},
	}
	for _, test := range tests {
		result := test.format.String()
		if result != test.expected {
			t.Errorf("String() for %d: got %q, want %q", test.format, result, test.expected)
		}
	}
}

func TestMetricsFormatContentType(t *testing.T) {
	tests := []struct {
		format   MetricsFormat
		expected string
	}{
		{Prometheus_0_0_4, "text/plain; version=0.0.4; charset=utf-8; escaping=underscores"},
		{OpenMetrics_1_0_0, "application/openmetrics-text; version=1.0.0; charset=utf-8"},
		{MetricsFormat(999), ""},
	}
	for _, test := range tests {
		result := test.format.ContentType()
		if result != test.expected {
			t.Errorf("ContentType() for %d: got %q, want %q", test.format, result, test.expected)
		}
	}
}

func TestParseMetricsFormat(t *testing.T) {
	tests := []struct {
		input       string
		expected    MetricsFormat
		expectError bool
	}{
		{"prometheus-0.0.4", Prometheus_0_0_4, false},
		{"PROMETHEUS-0.0.4", Prometheus_0_0_4, false},
		{"  openmetrics-1.0.0  ", OpenMetrics_1_0_0, false},
		{"OpenMetrics-1.0.0", OpenMetrics_1_0_0, false},
		{"nonsense", 0, true},
		{"", 0, true},
	}
	for _, test := range tests {
		result, err := ParseMetricsFormat(test.input)
		if test.expectError {
			if err == nil {
				t.Errorf("ParseMetricsFormat(%q): expected error, got nil", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("ParseMetricsFormat(%q): unexpected error: %v", test.input, err)
			}
			if result != test.expected {
				t.Errorf("ParseMetricsFormat(%q): got %v, want %v", test.input, result, test.expected)
			}
		}
	}
}

func TestNegotiateFormat(t *testing.T) {
	tests := []struct {
		name        string
		accept      string
		expected    MetricsFormat
		expectError bool
	}{
		{"empty accept defaults to prometheus", "", Prometheus_0_0_4, false},
		{"explicit prometheus", "text/plain; version=0.0.4", Prometheus_0_0_4, false},
		{"explicit openmetrics", "application/openmetrics-text; version=1.0.0", OpenMetrics_1_0_0, false},
		{"wildcard falls back to prometheus", "*/*", Prometheus_0_0_4, false},
		{"prefer higher quality", "text/plain; version=0.0.4; q=0.5, application/openmetrics-text; version=1.0.0; q=0.9", OpenMetrics_1_0_0, false},
		{"unsupported type errors", "application/json", 0, true},
		{"zero quality skipped", "text/plain; version=0.0.4; q=0", 0, true},
		{"wildcard with unsupported", "application/json, */*", Prometheus_0_0_4, false},
		{"empty part in accept", "text/plain; version=0.0.4,,", Prometheus_0_0_4, false},
		{"malformed media type skipped", "///bad, text/plain; version=0.0.4", Prometheus_0_0_4, false},
		{"invalid quality value skipped", "text/plain; version=0.0.4; q=abc, application/openmetrics-text; version=1.0.0", OpenMetrics_1_0_0, false},
	}
	for _, test := range tests {
		req := httptest.NewRequest("GET", "/metrics", nil)
		if test.accept != "" {
			req.Header.Set("Accept", test.accept)
		}
		result, err := NegotiateFormat(req)
		if test.expectError {
			if err == nil {
				t.Errorf("%s: expected error, got nil", test.name)
			}
		} else {
			if err != nil {
				t.Errorf("%s: unexpected error: %v", test.name, err)
			}
			if result != test.expected {
				t.Errorf("%s: got %v, want %v", test.name, result, test.expected)
			}
		}
	}
}
