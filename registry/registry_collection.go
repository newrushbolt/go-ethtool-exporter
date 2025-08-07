package registry

import "strings"

type RegistryCollection map[string]Registry

func (collection *RegistryCollection) GetAllMetricsText() string {
	var allMetrics []string
	for _, registry := range *collection {
		allMetrics = append(allMetrics, registry.FormatTextfileString())
	}
	return strings.Join(allMetrics, "\n")
}
