package interfaces_test

import (
	"testing"

	"github.com/newrushbolt/go-ethtool-exporter/interfaces"
	"github.com/stretchr/testify/assert"
)

func TestInterfacesAll(t *testing.T) {
	stubNetClassPath := "interfaces_test/sys/class/net"
	interfaces := interfaces.GetInterfacesList(stubNetClassPath, false)

	assert.Equal(t, interfaces, []string{"bond0", "eth0"})
}

func TestInterfacesBonded(t *testing.T) {
	stubNetClassPath := "interfaces_test/sys/class/net"
	interfaces := interfaces.GetInterfacesList(stubNetClassPath, true)

	assert.Equal(t, interfaces, []string{"eth0"})
}

func TestInterfacesBrokenPath(t *testing.T) {
	stubNetClassPath := "interfaces_test/sys/class/net2"

	assert.Panics(t, func() { interfaces.GetInterfacesList(stubNetClassPath, false) }, "Should panic without proper path")
}
