package registry

import (
	"fmt"
	"log/slog"
	"strings"
)

type Registry []MetricRecord

func (metricRegistry *Registry) GetMetricIndex(metricName string) (int, error) {
	var err error
	metricsFoundIndexes := []int{}
	for metricIndex, metricObject := range *metricRegistry {
		if metricObject.Name == metricName {
			metricsFoundIndexes = append(metricsFoundIndexes, metricIndex)
		}
	}

	switch len(metricsFoundIndexes) {
	case 0:
		return -1, err
	case 1:
		return metricsFoundIndexes[0], err
	default:
		metricsFoundObjects := Registry{}
		for _, metricIndex := range metricsFoundIndexes {
			metricsFoundObjects = append(metricsFoundObjects, (*metricRegistry)[metricIndex])
		}
		err = fmt.Errorf("multiple metrics with the same name <%s> found: %+v", metricName, metricsFoundObjects)
		return -1, err
	}
}

func (registry *Registry) FormatTextfileString() string {
	var allMetricLines []string

	for _, metric := range *registry {
		metricString, err := metric.FormatPrometheusLine()
		if err != nil {
			slog.Error("Cannot format metric: ", "metricFormatError", err)
			continue
		}
		allMetricLines = append(allMetricLines, metricString)

	}

	metrics := strings.Join(allMetricLines, "\n")
	return metrics
}
