package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"regexp"
	"runtime/debug"
	"testing"
	"time"

	"github.com/newrushbolt/go-ethtool-exporter/registry"

	"github.com/stretchr/testify/assert"
)

func setBuildInfo(mainVersion, vcsRevision, vcsTime string) func() (*debug.BuildInfo, bool) {
	return func() (*debug.BuildInfo, bool) {
		settings := []debug.BuildSetting{}
		if vcsRevision != "" {
			settings = append(settings, debug.BuildSetting{Key: "vcs.revision", Value: vcsRevision})
		}
		if vcsTime != "" {
			settings = append(settings, debug.BuildSetting{Key: "vcs.time", Value: vcsTime})
		}
		return &debug.BuildInfo{
			Main:     debug.Module{Version: mainVersion},
			Settings: settings,
		}, true
	}
}

func TestParseAllowedInterfaceTypes(t *testing.T) {
	typesNormal := parseAllowedInterfaceTypes("1,2,3")
	assert.Equal(t, []int{1, 2, 3}, typesNormal)

	typesSpaced := parseAllowedInterfaceTypes(" 1 , 2 , 3 ")
	assert.Equal(t, []int{1, 2, 3}, typesSpaced)

	typesEmpty := parseAllowedInterfaceTypes("")
	assert.Equal(t, []int{}, typesEmpty)

	typesEmptyWithCommas := parseAllowedInterfaceTypes(",,,")
	assert.Equal(t, []int{}, typesEmptyWithCommas)

	typesWithWord := parseAllowedInterfaceTypes("1,fuu,2")
	assert.Equal(t, []int{1, 2}, typesWithWord)
}

func TestExporterVersionAllFields(t *testing.T) {
	expectedVersion := `go-ethtool-exporter version: v1.2.3
vcs.revision: abc123
vcs.time: 2025-07-07T12:34:56Z`
	versionString := getExporterVersion(setBuildInfo("v1.2.3", "abc123", "2025-07-07T12:34:56Z"))
	assert.Equal(t, expectedVersion, versionString)
}

func TestExporterVersionNoFields(t *testing.T) {
	expectedVersion := `go-ethtool-exporter version: unknown`
	versionString := getExporterVersion(setBuildInfo("", "", ""))
	assert.Equal(t, expectedVersion, versionString)
}

func TestExporterVersionBuildInfoError(t *testing.T) {
	expectedVersion := `go-ethtool-exporter version: unknown`
	versionString := getExporterVersion(func() (*debug.BuildInfo, bool) { return nil, false })
	assert.Equal(t, expectedVersion, versionString)
}

func TestExporterWriteAllMetricsToTextfiles(t *testing.T) {
	expectedMetrics := `dummy_metric{foo="bar"} 42`
	dir := t.TempDir()
	textfileDirectory = &dir // override global pointer for test

	eth0Registry := registry.Registry{
		{
			Name:   "dummy_metric",
			Value:  42,
			Labels: map[string]string{"foo": "bar"},
		},
	}
	registries := map[string]registry.Registry{
		"eth0": eth0Registry,
	}

	writeAllMetricsToTextfiles(registries)

	filePath := dir + "/ethtool_exporter.prom"
	metrics, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}

	assert.Equal(t, expectedMetrics, string(metrics))
}

func TestExporterDirectoryMustExist(t *testing.T) {
	existingDir := "testdata/interfaces/"
	assert.NotPanics(t, func() { MustDirectoryExist(&existingDir) })

	nonExistentDir := "testdata/interfaces2/"
	assert.Panics(t, func() { MustDirectoryExist(&nonExistentDir) })

	notDir := "testdata/interfaces/sys/class/net/broken_interface"
	assert.Panics(t, func() { MustDirectoryExist(&notDir) })
}

func ptr[T any](v T) *T { return &v }

func setupHttpHandlerFlags(t *testing.T) {
	// Override global params for test
	portsRegexp := regexp.MustCompile("eth4")
	discoverPortsRegexp = &portsRegexp
	ethtoolPath = ptr("testdata/ethtool.sh")
	linuxNetClassPath = ptr("testdata/interfaces/sys/class/net")
	discoverAllowedPortTypes = ptr("1,")
	discoverAllPorts = ptr(true)
	discoverBondSlaves = ptr(false)
	discoverBridgeSlaves = ptr(false)
	collectGenericInfoModes = ptr(true)
	collectGenericInfoSettings = ptr(true)
	collectDriverInfoCommon = ptr(false)
	collectDriverInfoFeatures = ptr(false)
	collectModuleInfoDiagnosticsAlarms = ptr(false)
	collectModuleInfoDiagnosticsWarnings = ptr(false)
	collectModuleInfoDiagnosticsValues = ptr(false)
	collectModuleInfoVendor = ptr(false)
	keepAbsentMetrics = ptr(false)
	listLabelFormat = ptr("single-label")
	textfileDirectory = ptr(t.TempDir())
	ethtoolTimeout = ptr(time.Second * 5)
	loopTextfileUpdateInterval = ptr(time.Second)
}

func TestExporterHttpMetricsHandler(t *testing.T) {
	setupHttpHandlerFlags(t)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	metricsHandler(recorder, req)

	resp := recorder.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "text/plain; version=0.0.4; charset=utf-8; escaping=underscores", resp.Header.Get("Content-Type"))

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	expectedBytes, err := os.ReadFile("testdata/eth4.generic_info.prom")
	if err != nil {
		t.Fatalf("Failed to read expected metrics: %v", err)
	}
	expectedMetricResult := string(expectedBytes)
	assert.Equal(t, expectedMetricResult, string(body))
}

func TestExporterHttpMetricsHandlerFail(t *testing.T) {
	setupHttpHandlerFlags(t)
	linuxNetClassPath = ptr("non_existent_testdata/interfaces/sys/class/net")

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	metricsHandler(recorder, req)

	resp := recorder.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Equal(t, "text/plain; charset=utf-8", resp.Header.Get("Content-Type"))

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	expectedMetricResult := "Internal Server Error\n"
	assert.Equal(t, expectedMetricResult, string(body))
}

type errorWriter struct {
	http.ResponseWriter
}

func (ew *errorWriter) Write(b []byte) (int, error) {
	return 0, io.ErrClosedPipe
}
func (ew *errorWriter) Header() http.Header {
	return ew.ResponseWriter.Header()
}
func (ew *errorWriter) WriteHeader(statusCode int) {
	ew.ResponseWriter.WriteHeader(statusCode)
}

func TestExporterHttpMetricsHandlerWriteError(t *testing.T) {
	setupHttpHandlerFlags(t)

	recorder := httptest.NewRecorder()
	errWriter := &errorWriter{recorder}
	req := httptest.NewRequest("GET", "/metrics", nil)

	metricsHandler(errWriter, req)

	resp := recorder.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "", string(body))
}

func TestExporterSingleTextfile(t *testing.T) {
	setupHttpHandlerFlags(t)
	var err error
	*textfileDirectory, err = os.MkdirTemp(".", ".test-textfiles-*")
	defer os.RemoveAll(*textfileDirectory)
	assert.NoError(t, err)
	runSingleTextfileCommand()

	expectedMetricBytes, err := os.ReadFile("testdata/eth4.generic_info.prom")
	if err != nil {
		t.Fatalf("Failed to read expected metrics: %v", err)
	}
	expectedMetric := string(expectedMetricBytes)

	resultedMetricsBytes, err := os.ReadFile(path.Join(*textfileDirectory, "ethtool_exporter.prom"))
	if err != nil {
		t.Fatalf("Failed to read expected metrics: %v", err)
	}
	resultedMetrics := string(resultedMetricsBytes)

	assert.Equal(t, expectedMetric, resultedMetrics)
}
