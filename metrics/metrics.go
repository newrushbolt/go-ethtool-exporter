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

	"newrushbolt/go-ethtool-exporter/registry"

	"github.com/newrushbolt/go-ethtool-metrics/pkg/metrics/statistics"
)

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

func MetricListFromStructs(inputStruct any, metricList *registry.Registry, prefixes []string, extraLabels map[string]string, keepAbsentMetrics bool) {
	// TODO: add counters for all kind of metric events: parsed structures, parsed floats, parsed nils, parse errors, etc
	inputStructValue := reflect.ValueOf(inputStruct)
	switch inputStructValue.Kind() {
	// Handle pointers
	case reflect.Ptr:
		// TODO: Handle absent metrics logic due to flags
		if !inputStructValue.IsNil() {
			newPrefixes := slices.Clone(prefixes)
			MetricListFromStructs(inputStructValue.Elem().Interface(), metricList, newPrefixes, extraLabels, keepAbsentMetrics)
		} else {
			if !keepAbsentMetrics {
				slog.Debug("Skipping nil pointer because keepAbsentMetrics=false", "prefixes", prefixes)
				return
			}
			inputType := reflect.TypeOf(inputStruct)
			if inputType == reflect.TypeOf((*float64)(nil)) {
				if keepAbsentMetrics {
					slog.Debug("Adding `Nan` for missing float64 metric", "prefixes", prefixes)
					newPrefixes := slices.Clone(prefixes)
					nanValue := math.NaN()
					MetricListFromStructs(nanValue, metricList, newPrefixes, extraLabels, keepAbsentMetrics)
				}
			} else {
				slog.Debug("Skipping nil pointer, keeping nils only supporter for float64", "prefixes", prefixes, "type", inputType)
			}
		}
	// Handle structs
	case reflect.Struct:

		for structFieldIndex := range inputStructValue.NumField() {
			field := inputStructValue.Type().Field(structFieldIndex)
			newPrefixes := append(prefixes, []string{field.Name}...)
			MetricListFromStructs(inputStructValue.Field(structFieldIndex).Interface(), metricList, newPrefixes, extraLabels, keepAbsentMetrics)
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
					MetricListFromStructs(queueMetrics, metricList, newPrefixes, labels, keepAbsentMetrics)
				}
				// Do not add metric for subspace itself
				return
			} else {
				labelName := prefixes[len(prefixes)-1]
				prefixes = append(prefixes[:len(prefixes)-1], "info")
				metricName := toSnakeCase(strings.Join(prefixes, "_"))
				var labelValues []string
				if inputStructValue.Kind() == reflect.String {
					labelValues = append(labelValues, inputStructValue.String())
				} else {
					for elementIndex := range inputStructValue.Len() {
						element := inputStructValue.Index(elementIndex)
						labelValues = append(labelValues, element.String())
					}
				}
				labelValuesString := strings.Join(labelValues, ",")
				metricLabels[labelName] = labelValuesString
				metricIndex, err := metricList.GetMetricIndex(metricName)
				if err != nil {
					slog.Error("Error getting metric index", "metricName", metricName, "error", err)
					return
				}
				// TODO: explain this branch
				if metricIndex != -1 {
					maps.Insert((*metricList)[metricIndex].Labels, maps.All(metricLabels))
					return
				}
				metricValue = float64(1)
			}
		default:
			logMetricName := toSnakeCase(strings.Join(prefixes, "_"))
			slog.Debug("Error: cannot format type as metric value", "kind", inputStructValue.Kind(), "metricName", logMetricName)
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
