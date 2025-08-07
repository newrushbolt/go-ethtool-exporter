package main

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/newrushbolt/go-ethtool-exporter/interfaces"
	"golang.org/x/net/netutil"
)

func runDiscoverPortsCommand() {
	// Discover ports mode
	allowedTypes := parseAllowedInterfaceTypes(*discoverAllowedPortTypes)
	discoverConfig := createDiscoveryConfig()
	interfaces := interfaces.GetInterfacesList(*linuxNetClassPath, discoverConfig, allowedTypes)
	if len(interfaces) == 0 {
		fmt.Println("No ports discovered, re-run with `GO_ETHTOOL_EXPORTER_LOG_LEVEL=DEBUG` to check the discovery logic")
	} else {
		fmt.Println("Discovered following ports:")
		interfacesString := fmt.Sprintf("  - %s", strings.Join(interfaces, "\n  - "))
		fmt.Println(interfacesString)
	}
}

func runSingleTextfileCommand() {
	// Single textfile mode
	MustDirectoryExist(textfileDirectory)
	metricRegistries := collectMetrics()
	writeAllMetricsToTextfiles(metricRegistries)
}

func runLoopTextfileCommand() {
	// Loop textfile mode
	MustDirectoryExist(textfileDirectory)
	for {
		metricRegistries := collectMetrics()
		writeAllMetricsToTextfiles(metricRegistries)
		time.Sleep(*loopTextfileUpdateInterval)
	}
}

// Middleware for logging requests and filtering
func loggingAndFilterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("HTTP request", "method", r.Method, "url", r.URL.String(), "remote", r.RemoteAddr, "content-type", r.Header.Get("Content-Type"))

		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		contentTypeHeader := r.Header.Get("Content-Type")
		allowedContentTypes := []string{
			"",
			"text/plain",
			"text/plain; version=0.0.4",
		}
		if !slices.Contains(allowedContentTypes, contentTypeHeader) {
			http.Error(w, "Unsupported Content-Type", http.StatusUnsupportedMediaType)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			slog.Error("Panic in metricsHandler", "panic", rec)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}()
	metricRegistries := collectMetrics()
	// The same as in node_exporter :shrug:
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8; escaping=underscores")
	allMetrics := metricRegistries.GetAllMetricsText()
	_, err := w.Write([]byte(allMetrics))
	if err != nil {
		slog.Error("Failed to write response", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func runHttpServerCommand() {
	slog.Info("Starting HTTP server", "address", *httpListenAddress)
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", metricsHandler)
	wrappedMux := loggingAndFilterMiddleware(mux)

	rawListener, err := net.Listen("tcp", *httpListenAddress)
	if err != nil {
		slog.Error("Failed to start listening", "listenAddr", *httpListenAddress, "error", err)
		os.Exit(1)
	}
	listener := netutil.LimitListener(rawListener, *httpMaxRequests)
	// TODO: make use of prometheus standart library for tls config
	// https://github.com/prometheus/exporter-toolkit/blob/master/web/tls_config.go
	server := &http.Server{
		Handler: wrappedMux,
		// Hardcoded limits are better than no limits :shrug:
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	// It looks like we don't really need keepalive in exporter
	server.SetKeepAlivesEnabled(false)

	err = server.Serve(listener)
	if err != nil {
		slog.Error("Failed to start HTTP server", "error", err)
		os.Exit(1)
	}
}
