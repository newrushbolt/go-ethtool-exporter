package main

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/newrushbolt/go-ethtool-exporter/interfaces"
	"github.com/newrushbolt/go-ethtool-exporter/registry"
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

func runSingleTextfileCommand(format registry.MetricsFormat) {
	// Single textfile mode
	MustDirectoryExist(textfileDirectory)
	metricRegistries := collectMetrics()
	writeAllMetricsToTextfiles(metricRegistries, format)
}

func runLoopTextfileCommand(format registry.MetricsFormat) {
	// Loop textfile mode
	MustDirectoryExist(textfileDirectory)
	for {
		metricRegistries := collectMetrics()
		writeAllMetricsToTextfiles(metricRegistries, format)
		time.Sleep(*loopTextfileUpdateInterval)
	}
}

// Middleware for logging requests and filtering
func loggingAndFilterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("HTTP request", "method", r.Method, "url", r.URL.String(), "remote", r.RemoteAddr, "accept", r.Header.Get("Accept"))

		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		// ensure the client accepts one of the supported response formats
		if _, err := registry.NegotiateFormat(r); err != nil {
			http.Error(w, "Not Acceptable", http.StatusNotAcceptable)
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
	// Determine the format to render; middleware should have already validated the
	// header, but we repeat the parsing here in case the handler is called
	// directly by tests.
	format, err := registry.NegotiateFormat(r)
	if err != nil {
		http.Error(w, "Not Acceptable", http.StatusNotAcceptable)
		return
	}
	w.Header().Set("Content-Type", format.ContentType())
	allMetrics := metricRegistries.GetAllMetricsText(format)
	_, err = w.Write([]byte(allMetrics))
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
