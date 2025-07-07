package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAllowedInterfaceTypes(t *testing.T) {
	types := parseAllowedInterfaceTypes("1,2,3")
	assert.Equal(t, []int{1, 2, 3}, types)

	types = parseAllowedInterfaceTypes(" 1 , 2 , 3 ")
	assert.Equal(t, []int{1, 2, 3}, types)

	types = parseAllowedInterfaceTypes("")
	assert.Equal(t, types, []int{})

	types = parseAllowedInterfaceTypes(",,,")
	assert.Equal(t, types, []int{})

	types = parseAllowedInterfaceTypes("1,foo,2")
	assert.Equal(t, []int{1, 2}, types)
}

func TestReadEthtoolData(t *testing.T) {
	stubPath := "testdata/ethtool.sh"
	// Make sure the stub is executable
	os.Chmod(stubPath, 0755)

	// No mode
	out := readEthtoolData("eth0", "", stubPath)
	assert.Contains(t, out, "ethtool output for eth0")

	// -i mode
	out = readEthtoolData("eth0", "-i", stubPath)
	assert.Contains(t, out, "driver info for eth0")

	// -m mode
	out = readEthtoolData("eth0", "-m", stubPath)
	assert.Contains(t, out, "module info for eth0")

	// -S mode
	out = readEthtoolData("eth0", "-S", stubPath)
	assert.Contains(t, out, "statistics for eth0")

	// Unknown interface
	out = readEthtoolData("unknown", "", stubPath)
	assert.Equal(t, "", out)
}
