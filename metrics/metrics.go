package metrics

import (
	"fmt"
	"log/slog"
	"maps"
	"math"
	"reflect"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/newrushbolt/go-ethtool-exporter/registry"

	"github.com/newrushbolt/go-ethtool-metrics/pkg/metrics/statistics"
)

const AbsentMetricDetailedName = "missing_metric_info"
const AbsentMetricTotalName = "missing_metrics_total"
const AbsentMetricSkippedTotalName = "missing_metrics_skipped_total"
const MetricParseErrorTotalName = "metric_parse_error_total"

type AbsentMetricsConfig struct {
	ExposeNan          bool
	ExposeTotalCounter bool
	ExposeDetailedInfo bool
}

func toSnakeCase(inputString string) string {
	var replacePairs []string
	for i := 1; i < len(inputString); i++ {
		firstChar, _ := utf8.DecodeRune([]byte{inputString[i-1]})
		secondChar, _ := utf8.DecodeRune([]byte{inputString[i]})

		if (unicode.IsLower(firstChar) && unicode.IsLetter(firstChar)) && (unicode.IsUpper(secondChar) && unicode.IsLetter(secondChar)) {
			replacePair := fmt.Sprintf("%s%s", string(inputString[i-1]), string(inputString[i]))
			replacePairs = append(replacePairs, replacePair)
		}
	}
	for _, replacePair := range replacePairs {
		replaceIndex := strings.Index(inputString, replacePair)
		stringSlice := strings.Split(inputString, "")
		stringSlice[replaceIndex+1] = strings.ToLower(stringSlice[replaceIndex+1])
		stringSlice = slices.Insert(stringSlice, replaceIndex+1, "_")
		inputString = strings.Join(stringSlice, "")
	}
	return strings.ToLower(inputString)
}

func MetricListFromStructs(inputStruct any, metricList *registry.Registry, prefixes []string, extraLabels map[string]string, absentMetrics AbsentMetricsConfig, listLabelFormat string) {
	inputStructValue := reflect.ValueOf(inputStruct)
	switch inputStructValue.Kind() {
	// Handle pointers
	case reflect.Ptr:
		// TODO: Handle absent metrics logic due to flags
		if !inputStructValue.IsNil() {
			newPrefixes := slices.Clone(prefixes)
			MetricListFromStructs(inputStructValue.Elem().Interface(), metricList, newPrefixes, extraLabels, absentMetrics, listLabelFormat)
		} else {
			missingMetricName := toSnakeCase(strings.Join(prefixes, "_"))

			inputType := reflect.TypeOf(inputStruct)
			if inputType != reflect.TypeOf((*float64)(nil)) {
				slog.Debug("Skipping nil pointer, keeping nils only supported for float64", "metric_name", missingMetricName, "type", inputType)
				metricList.IncrementCounter(AbsentMetricSkippedTotalName, nil, 1)
				return
			}

			if absentMetrics.ExposeNan {
				slog.Debug("Setting `Nan` value for missing metric", "metric_name", missingMetricName, "labels", extraLabels)
				newPrefixes := slices.Clone(prefixes)
				nanValue := math.NaN()
				MetricListFromStructs(nanValue, metricList, newPrefixes, extraLabels, absentMetrics, listLabelFormat)
			}

			if absentMetrics.ExposeTotalCounter {
				slog.Debug("Incrementing total counter for missing metric", "metric_name", missingMetricName, "labels", extraLabels)
				metricList.IncrementCounter(AbsentMetricTotalName, extraLabels, 1)
			}

			if absentMetrics.ExposeDetailedInfo {
				slog.Debug("Adding detailed info for missing metric", "metric_name", missingMetricName, "labels", extraLabels)
				finalLabels := map[string]string{
					"metric_name": missingMetricName,
				}
				maps.Insert(finalLabels, maps.All(extraLabels))
				metricRecord := registry.MetricRecord{
					Name:   AbsentMetricDetailedName,
					Labels: finalLabels,
					Value:  1,
				}
				*metricList = append(*metricList, metricRecord)
			}
		}
	// Handle structs
	case reflect.Struct:
		for structFieldIndex := range inputStructValue.NumField() {
			field := inputStructValue.Type().Field(structFieldIndex)
			newPrefixes := append(prefixes, []string{field.Name}...)
			MetricListFromStructs(inputStructValue.Field(structFieldIndex).Interface(), metricList, newPrefixes, extraLabels, absentMetrics, listLabelFormat)
		}
	// Handle simple types
	default:
		var metricValue float64
		metricLabels := make(map[string]string)
		switch inputStructValue.Kind() {
		case reflect.Float64:
			metricValue = inputStructValue.Float()
		case reflect.Uint64:
			metricValue = float64(inputStructValue.Uint())
		case reflect.Float32:
			metricValue = inputStructValue.Float()
		case reflect.Bool:
			metricValue = float64(0)
			if inputStructValue.Bool() {
				metricValue = float64(1)
			}
		// Auto-convert slices and strings to single info-metric
		case reflect.String, reflect.Slice:
			perQueueStatistics, isPerQueueStatistics := inputStruct.(statistics.PerQueueStatistics)
			// We tried to keep type abstract as long as possible, but per-queue processing is so special,
			// so we need a special branch for it
			// TODO: find a better way?
			if isPerQueueStatistics {
				for queue := range inputStructValue.Len() {
					queueMetrics := perQueueStatistics[queue]
					newPrefixes := slices.Clone(prefixes)
					labels := map[string]string{
						"queue": fmt.Sprintf("%d", queue),
					}
					MetricListFromStructs(queueMetrics, metricList, newPrefixes, labels, absentMetrics, listLabelFormat)
				}
				// Do not add metric for subspace itself
				return
			}

			labelName := prefixes[len(prefixes)-1]
			prefixes = append(prefixes[:len(prefixes)-1], "info")
			metricName := toSnakeCase(strings.Join(prefixes, "_"))

			if inputStructValue.Kind() == reflect.String {
				metricLabels[labelName] = inputStructValue.String()
			} else {
				var labelValuesToJoin []string
				for elementIndex := range inputStructValue.Len() {
					element := inputStructValue.Index(elementIndex)
					if slices.Contains([]string{"multi-label", "both"}, listLabelFormat) {
						partedLabelName := fmt.Sprintf("%sP%d", labelName, elementIndex)
						metricLabels[partedLabelName] = element.String()
					}
					if slices.Contains([]string{"single-label", "both"}, listLabelFormat) {
						labelValuesToJoin = append(labelValuesToJoin, element.String())
					}
				}
				if len(labelValuesToJoin) > 0 {
					metricLabels[labelName] = strings.Join(labelValuesToJoin, ",")
				}
			}

			metricIndex, err := metricList.GetMetricIndex(metricName)
			if err != nil {
				slog.Error("Error getting metric index", "metricName", metricName, "error", err)
				metricList.IncrementCounter(MetricParseErrorTotalName, nil, 1)
				return
			}

			// If an info metric with this name already exists, merge labels into it
			// instead of creating a duplicate. This consolidates multiple string/slice
			// fields from the same struct into a single info metric with combined labels.
			if metricIndex != -1 {
				maps.Insert((*metricList)[metricIndex].Labels, maps.All(metricLabels))
				return
			}
			metricValue = float64(1)
		default:
			logMetricName := toSnakeCase(strings.Join(prefixes, "_"))
			slog.Debug("Error: cannot format type as metric value", "kind", inputStructValue.Kind(), "metricName", logMetricName)
			metricList.IncrementCounter(MetricParseErrorTotalName, nil, 1)
			return
		}

		metricName := toSnakeCase(strings.Join(prefixes, "_"))
		finalLabels := map[string]string{}
		maps.Insert(finalLabels, maps.All(metricLabels))
		maps.Insert(finalLabels, maps.All(extraLabels))
		metricRecord := registry.MetricRecord{
			Name:   metricName,
			Labels: finalLabels,
			Value:  metricValue,
		}
		*metricList = append(*metricList, metricRecord)
	}
}
