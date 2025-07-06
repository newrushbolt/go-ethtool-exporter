package interfaces

import (
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

	assert.Panics(t, func() { GetInterfacesList(stubNetClassPath, false, allowedTypes) }, "Should panic without proper path")
}
