package interfaces

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var defaultNetClassPath = "../testdata/interfaces/sys/class/net"

func TestInterfacesAll(t *testing.T) {
	allowedTypes := []int{1}
	interfaces := GetInterfacesList(defaultNetClassPath, false, allowedTypes)

	assert.Equal(t, interfaces, []string{"bond0", "eth0"})
}

func TestInterfacesBonded(t *testing.T) {
	allowedTypes := []int{1}
	interfaces := GetInterfacesList(defaultNetClassPath, true, allowedTypes)

	assert.Equal(t, interfaces, []string{"eth0"})
}

func TestInterfacesBrokenPath(t *testing.T) {
	absentNetClassPath := "../testdata/interfaces/sys/class/net2"
	allowedTypes := []int{1}

	assert.Panics(t, func() { GetInterfacesList(absentNetClassPath, false, allowedTypes) })
}

func TestIsInterfaceBondedPermissionError(t *testing.T) {
	unreadableFile := "../testdata/interfaces/sys/class/net/unreadable_file"
	os.Chmod(unreadableFile, 0000) // Set permissions to unreadable
	defer func(f string) {
		t.Cleanup(func() {
			os.Chmod(f, 0644) // Restore permissions after test
		})
	}(unreadableFile)

	isBonded := isInterfaceBonded(unreadableFile)
	assert.False(t, isBonded)
}
