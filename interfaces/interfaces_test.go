package interfaces

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInterfacesAll(t *testing.T) {
	stubNetClassPath := "interfaces_test/sys/class/net"
	allowedTypes := []int{1}
	interfaces := GetInterfacesList(stubNetClassPath, false, allowedTypes)

	assert.Equal(t, interfaces, []string{"bond0", "eth0"})
}

func TestInterfacesBonded(t *testing.T) {
	stubNetClassPath := "interfaces_test/sys/class/net"
	allowedTypes := []int{1}
	interfaces := GetInterfacesList(stubNetClassPath, true, allowedTypes)

	assert.Equal(t, interfaces, []string{"eth0"})
}

func TestInterfacesBrokenPath(t *testing.T) {
	stubNetClassPath := "interfaces_test/sys/class/net2"
	allowedTypes := []int{1}

	assert.Panics(t, func() { GetInterfacesList(stubNetClassPath, false, allowedTypes) })
}

func TestIsInterfaceBondedFilesystemErrors(t *testing.T) {
	unreadableFiles := []string{
		"interfaces_test/sys/class/net/unreadable_file",
		"interfaces_test/sys/class/net/eth3/bonding_slave/state",
	}
	for _, file := range unreadableFiles {
		os.Chmod(file, 0000) // Set permissions to unreadable
		defer func(f string) {
			t.Cleanup(func() {
				os.Chmod(f, 0644) // Restore permissions after test
			})
		}(file)
	}

	assert.False(t, isInterfaceBonded("interfaces_test/sys/class/net/unreadable_file"))
	assert.False(t, isInterfaceBonded("interfaces_test/sys/class/net/eth3"))
}
