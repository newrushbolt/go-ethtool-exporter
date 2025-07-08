package interfaces

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var defaultNetClassPath = "../testdata/interfaces/sys/class/net"

func TestInterfacesAllTypes(t *testing.T) {
	allowedTypes := []int{}
	discoverConfig := PortDiscoveryOptions{
		DiscoverAllPorts:   true,
		DiscoverBondSlaves: true,
	}
	interfaces := GetInterfacesList(defaultNetClassPath, discoverConfig, allowedTypes)

	assert.Equal(t, interfaces, []string{"bond0", "eth0", "eth1", "eth2", "eth3"})
}

func TestInterfacesAllEthernet(t *testing.T) {
	allowedTypes := []int{1}
	discoverConfig := PortDiscoveryOptions{
		DiscoverAllPorts:   true,
		DiscoverBondSlaves: true,
	}
	interfaces := GetInterfacesList(defaultNetClassPath, discoverConfig, allowedTypes)

	assert.Equal(t, interfaces, []string{"bond0", "eth0"})
}

func TestInterfacesBonded(t *testing.T) {
	allowedTypes := []int{1}
	discoverConfig := PortDiscoveryOptions{
		DiscoverAllPorts:   false,
		DiscoverBondSlaves: true,
	}
	interfaces := GetInterfacesList(defaultNetClassPath, discoverConfig, allowedTypes)

	assert.Equal(t, interfaces, []string{"eth0"})
}

func TestInterfacesBrokenPath(t *testing.T) {
	absentNetClassPath := "../testdata/interfaces/sys/class/net2"
	allowedTypes := []int{1}
	discoverConfig := PortDiscoveryOptions{
		DiscoverAllPorts:   false,
		DiscoverBondSlaves: true,
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
