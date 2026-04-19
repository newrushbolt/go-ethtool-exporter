package registry

import (
	"fmt"
	"mime"
	"net/http"
	"strconv"
	"strings"
)

type MetricsFormat int

const (
	Prometheus_0_0_4 MetricsFormat = iota
	OpenMetrics_1_0_0
)

type formatInfo struct {
	name        string
	contentType string
	mediaType   string
	version     string
}

var formatRegistry = map[MetricsFormat]formatInfo{
	Prometheus_0_0_4: {
		name:        "prometheus-0.0.4",
		contentType: "text/plain; version=0.0.4; charset=utf-8; escaping=underscores",
		mediaType:   "text/plain",
		version:     "0.0.4",
	},
	OpenMetrics_1_0_0: {
		name:        "openmetrics-1.0.0",
		contentType: "application/openmetrics-text; version=1.0.0; charset=utf-8",
		mediaType:   "application/openmetrics-text",
		version:     "1.0.0",
	},
}

func (format MetricsFormat) String() string {
	if info, ok := formatRegistry[format]; ok {
		return info.name
	}
	return "unknown"
}

func (format MetricsFormat) ContentType() string {
	if info, ok := formatRegistry[format]; ok {
		return info.contentType
	}
	return ""
}

func ParseMetricsFormat(str string) (MetricsFormat, error) {
	lower := strings.ToLower(strings.TrimSpace(str))
	for format, info := range formatRegistry {
		if info.name == lower {
			return format, nil
		}
	}
	return 0, fmt.Errorf("unknown metrics format: %s", str)
}

// resolveMediaType maps a parsed media type + version to a MetricsFormat.
func resolveMediaType(mediaType string, version string) (MetricsFormat, bool) {
	for format, info := range formatRegistry {
		if info.mediaType == mediaType && (version == "" || version == info.version) {
			return format, true
		}
	}
	return 0, false
}

// NegotiateFormat inspects the Accept header and returns the best supported
// response format.
//
// If the Accept header is missing, Prometheus 0.0.4 is used as a fallback.
// If Accept is present but does not include any supported type (and no */*
// wildcard), an error is returned.
func NegotiateFormat(req *http.Request) (MetricsFormat, error) {
	accept := strings.TrimSpace(req.Header.Get("Accept"))
	if accept == "" {
		return Prometheus_0_0_4, nil
	}

	type formatCandidate struct {
		format  MetricsFormat
		quality float64
	}

	var best *formatCandidate
	allowAny := false

	for _, part := range strings.Split(accept, ",") {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}

		mediaType, params, err := mime.ParseMediaType(item)
		if err != nil {
			continue
		}

		quality := 1.0
		if raw, ok := params["q"]; ok {
			parsed, err := strconv.ParseFloat(raw, 64)
			if err != nil {
				continue
			}
			quality = parsed
		}
		if quality <= 0 {
			continue
		}

		if mediaType == "*/*" {
			allowAny = true
			continue
		}

		format, ok := resolveMediaType(mediaType, params["version"])
		if !ok {
			continue
		}

		candidate := formatCandidate{format: format, quality: quality}
		if best == nil || candidate.quality > best.quality {
			best = &candidate
		}
	}

	if best != nil {
		return best.format, nil
	}
	if allowAny {
		return Prometheus_0_0_4, nil
	}

	return 0, fmt.Errorf("unsupported Accept header: %s", accept)
}
