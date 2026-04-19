package registry

import "strings"

type RegistryCollection map[string]Registry

func (collection *RegistryCollection) GetAllMetricsText(format MetricsFormat) string {
	var allMetrics []string
	for _, registry := range *collection {
		allMetrics = append(allMetrics, registry.FormatTextfileString(format))
	}
	return strings.Join(allMetrics, "\n")
}
