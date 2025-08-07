package interfaces

import (
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

var defaultNetClassPath = "../testdata/interfaces/sys/class/net"

func TestInterfacesRegexp(t *testing.T) {
	allowedTypes := []int{}
	discoverConfig := PortDiscoveryOptions{
		PortsRegexp:      regexp.MustCompile("b[a-z]+[0-1]"),
		DiscoverAllPorts: true,
	}
	interfaces := GetInterfacesList(defaultNetClassPath, discoverConfig, allowedTypes)

	assert.Equal(t, []string{"bond0", "br0"}, interfaces)
}

func TestInterfacesAllTypes(t *testing.T) {
	allowedTypes := []int{}
	discoverConfig := PortDiscoveryOptions{
		PortsRegexp:      regexp.MustCompile(".+"),
		DiscoverAllPorts: true,
	}
	interfaces := GetInterfacesList(defaultNetClassPath, discoverConfig, allowedTypes)

	assert.Equal(t, []string{"bond0", "br0", "eth0", "eth1", "eth2", "eth3", "eth4", "slave0"}, interfaces)
}

func TestInterfacesAllEthernet(t *testing.T) {
	allowedTypes := []int{1}
	discoverConfig := PortDiscoveryOptions{
		PortsRegexp:      regexp.MustCompile(".+"),
		DiscoverAllPorts: true,
	}
	interfaces := GetInterfacesList(defaultNetClassPath, discoverConfig, allowedTypes)

	assert.Equal(t, []string{"bond0", "br0", "eth0", "eth4", "slave0"}, interfaces)
}

func TestInterfacesBonded(t *testing.T) {
	allowedTypes := []int{1}
	discoverConfig := PortDiscoveryOptions{
		PortsRegexp:        regexp.MustCompile(".+"),
		DiscoverAllPorts:   false,
		DiscoverBondSlaves: true,
	}
	interfaces := GetInterfacesList(defaultNetClassPath, discoverConfig, allowedTypes)

	assert.Equal(t, []string{"eth0"}, interfaces)
}

func TestInterfacesBridged(t *testing.T) {
	allowedTypes := []int{1}
	discoverConfig := PortDiscoveryOptions{
		PortsRegexp:          regexp.MustCompile(".+"),
		DiscoverBridgeSlaves: true,
	}
	interfaces := GetInterfacesList(defaultNetClassPath, discoverConfig, allowedTypes)

	assert.Equal(t, []string{"slave0"}, interfaces)
}

func TestInterfacesBrokenPath(t *testing.T) {
	absentNetClassPath := "../testdata/interfaces/sys/class/net2"
	allowedTypes := []int{1}
	discoverConfig := PortDiscoveryOptions{
		PortsRegexp:      regexp.MustCompile(".+"),
		DiscoverAllPorts: true,
	}

	assert.Panics(t, func() { GetInterfacesList(absentNetClassPath, discoverConfig, allowedTypes) })
}

func TestIsInterfaceBondedPermissionError(t *testing.T) {
	unreadableFile := "../testdata/interfaces/sys/class/net/unreadable_file"
	os.Chmod(unreadableFile, 0000) // Set permissions to unreadable
	defer func(f string) {
		t.Cleanup(func() {
			os.Chmod(f, 0644) // Restore permissions after test
		})
	}(unreadableFile)

	isBonded := isInterfaceBondSlave(unreadableFile)
	assert.False(t, isBonded)
}
