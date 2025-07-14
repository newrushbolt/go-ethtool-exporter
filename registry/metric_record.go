package registry

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
)

type MetricRecord struct {
	Name   string
	Labels map[string]string
	Value  float64
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

	return fmt.Sprintf("%s{%s} %v", metricRecord.Name, labelStrings, metricRecord.Value), nil
}
