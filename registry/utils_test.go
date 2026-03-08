package registry

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeLabelPairDropsInvalidNames(t *testing.T) {
	inputs := []string{
		"iface_speed", // valid
		"",            // empty
		"9port",       // leading digit
		"if-name:1",   // invalid chars
		"éth0",        // unicode
	}

	var lines []string
	for _, in := range inputs {
		name, value, keep := sanitizelabelPair(in, "value")
		lines = append(lines, fmt.Sprintf("%-12s %-6s %t", name, value, keep))
	}
	results := strings.Join(lines, "\n")

	expected := "" +
		"iface_speed  value  true\n" +
		"                    false\n" +
		"                    false\n" +
		"                    false\n" +
		"                    false"

	assert.Equal(t, expected, results)
}

func TestSanitizeLabelPairDropsUnsafeChars(t *testing.T) {
	name, value, keep := sanitizelabelPair("iface", "back\\slash new\nline double\"quote")

	assert.Equal(t, "iface", name)
	assert.Equal(t, "backslash newline doublequote", value)
	assert.Equal(t, true, keep)
}
