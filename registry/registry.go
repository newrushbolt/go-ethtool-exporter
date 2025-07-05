package registry

import (
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type MetricRecord struct {
	Name   string
	Labels map[string]string
	Value  float64
}

type Registry []MetricRecord

func (metricRegistry *Registry) GetMetricIndex(metricName string) int {
	metricsFoundIndexes := []int{}
	// var err error
	for metricIndex, metricObject := range *metricRegistry {
		if metricObject.Name == metricName {
			metricsFoundIndexes = append(metricsFoundIndexes, metricIndex)
		}
	}
	// TODO: WTF is this code? Why we need it?
	// Skipping coverage check for now
	// default:
	// 	metricsFoundObjects := Registry{}
	// 	for _, metricIndex := range metricsFoundIndexes {
	// 		metricsFoundObjects = append(metricsFoundObjects, (*metricRegistry)[metricIndex])
	// 	}
	// 	err = fmt.Errorf("multiple metrics with the same name <%s> found: %+v", metricName, metricsFoundObjects)
	// 	return -1, err
	if len(metricsFoundIndexes) == 0 {
		return -1
	}
	return metricsFoundIndexes[0]
}

// TODO: actually implement function, return proper error
func sanitizelabelPair(labelName, labelValue string) (string, string) {
	// slog.Error("Skipping label for metric", "metric", metricRecord.Name, "labelName", labelName, "labelValue", labelValue, "error", err)
	return labelName, labelValue
}

// May be missing some sanitizing options
func (metricRecord *MetricRecord) FormatPrometheusLine() (string, error) {
	if len(metricRecord.Labels) > 16 {
		errMsg := fmt.Sprintf("Metric <%s> has more than 16 label pairs: %v", metricRecord.Name, metricRecord.Labels)
		return "", errors.New(errMsg)
	}

	var labelStringsList []string
	sortedLabelKeys := slices.Collect(maps.Keys(metricRecord.Labels))
	slices.Sort(sortedLabelKeys)
	for _, labelName := range sortedLabelKeys {
		labelValue := metricRecord.Labels[labelName]
		cleanlabelName, cleanlabelValue := sanitizelabelPair(labelName, labelValue)
		// if err != nil {
		// 	slog.Error("Skipping label for metric", "metric", metricRecord.Name, "error", err)
		// 	continue
		// }
		labelString := fmt.Sprintf("%s=\"%s\"", cleanlabelName, cleanlabelValue)
		labelStringsList = append(labelStringsList, labelString)
	}
	labelStrings := strings.Join(labelStringsList, ",")

	return fmt.Sprintf("%s{%s} %v\n", metricRecord.Name, labelStrings, metricRecord.Value), nil
}

func (registry *Registry) MustWriteTextfile(filePath string) {
	tmpDir := filepath.Dir(filePath)
	tmpFile, err := os.CreateTemp(tmpDir, "ethtool_exporter.prom-*")
	if err != nil {
		panic(err)
	}
	defer os.Remove(tmpFile.Name())

	// `os.CreateTemp` opens file without O_APPEND, so we need to offset write of each line on our own
	writeIndex := int64(0)
	for _, metric := range *registry {
		metricString, err := metric.FormatPrometheusLine()
		if err != nil {
			slog.Error("Cannot format metric: ", "metricFormatError", err)
			continue
		}
		_, err = tmpFile.WriteAt([]byte(metricString), writeIndex)
		if err != nil {
			slog.Error("Cannot write formated metric to file", "metric", metricString, "file", tmpFile.Name(), "error", err)
			continue
		}
		writeIndex += int64(len(metricString))
	}
	if err := tmpFile.Close(); err != nil {
		panic(err)
	}

	// Not sure, we should check how default mod is set
	// if err := os.Chmod(tmpFile.Name(), 0o644); err != nil {
	// 	return err
	// }

	if err := os.Rename(tmpFile.Name(), filePath); err != nil {
		panic(err)
	}
}
