package main

import (
	"os"
	"runtime/debug"
	"testing"

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
	expectedMetrics := `
dummy_metric{foo="bar"} 42`
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
