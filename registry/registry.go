package registry

import (
	"fmt"
	"log/slog"
	"maps"
	"slices"
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

func (registry *Registry) FormatTextfileString(format MetricsFormat) string {
	var allMetricLines []string
	// keep track of seen metric names for OpenMetrics TYPE annotations
	seen := map[string]struct{}{}

	for _, metric := range *registry {
		metricString, err := metric.FormatPrometheusLine()
		if err != nil {
			slog.Error("Cannot format metric: ", "metricFormatError", err)
			continue
		}
		allMetricLines = append(allMetricLines, metricString)
		if format == OpenMetrics_1_0_0 {
			seen[metric.Name] = struct{}{}
		}
	}

	// assemble output based on format
	switch format {
	case OpenMetrics_1_0_0:
		// prepend TYPE annotations in deterministic order
		var types []string
		for name := range seen {
			types = append(types, name)
		}
		slices.Sort(types)
		var outLines []string
		for _, name := range types {
			outLines = append(outLines, "# TYPE "+name+" UNKNOWN")
		}
		outLines = append(outLines, allMetricLines...)
		outLines = append(outLines, "# EOF")
		return strings.Join(outLines, "\n")
	default:
		return strings.Join(allMetricLines, "\n")
	}
}

func (registry *Registry) AddLabelsToSomeMetrics(targetMetricName string, extraLabels map[string]string) {
	for metricIndex, metricObj := range *registry {
		if metricObj.Name == targetMetricName {
			newLabels := map[string]string{}
			maps.Insert(newLabels, maps.All(metricObj.Labels))
			maps.Insert(newLabels, maps.All(extraLabels))
			(*registry)[metricIndex].Labels = newLabels
		}
	}
}
