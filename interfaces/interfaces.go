package interfaces

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

type PortDiscoveryOptions struct {
	PortsRegexp          *regexp.Regexp
	DiscoverAllPorts     bool
	DiscoverBondSlaves   bool
	DiscoverBridgeSlaves bool
	// TODO: filter out ovs ports by reading link, eg `/sys/class/net/tap2473528a-b1/master -> ../ovs-system`
}

func isInterfaceTypeValid(devicePath string, allowedInterfaceTypes []int) bool {
	if len(allowedInterfaceTypes) == 0 {
		slog.Debug("No allowed interface types specified, allowing all types", "devicePath", devicePath)
		return true
	}
	typePath := path.Join(devicePath, "type")
	interfaceTypeRaw, err := os.ReadFile(typePath)
	if err != nil {
		slog.Debug("Cannot read device type for device", "devicePath", devicePath, "error", err)
		return false
	}
	interfaceType, err := strconv.Atoi(strings.TrimSpace(string(interfaceTypeRaw)))
	if err != nil {
		slog.Debug("Cannot parse device type for device", "devicePath", devicePath, "error", err)
		return false
	}
	if !slices.Contains(allowedInterfaceTypes, interfaceType) {
		slog.Debug("Interface type is not allowed", "interfaceType", interfaceType, "allowedTypes", allowedInterfaceTypes, "devicePath", devicePath)
		return false
	}
	return true
}

func isInterfaceBondSlave(devicePath string) bool {
	slaveStatePath := path.Join(devicePath, "bonding_slave/state")
	// TODO: better check of file state (symlink, dir, etc) and maybe dumb read
	_, err := os.Stat(slaveStatePath)
	if os.IsNotExist(err) {
		slog.Debug("Device is not a bond slave", "devicePath", devicePath)
		return false
	} else if err != nil {
		slog.Warn("Device slave status file cannot be accessed", "devicePath", devicePath, "error", err)
		return false
	}
	return true
}

func isInterfaceBridgeSlave(devicePath string) bool {
	bridgeFlagsPath := path.Join(devicePath, "brport/bridge/type")
	// TODO: better check of file state (symlink, dir, etc) and maybe dumb read
	_, err := os.Stat(bridgeFlagsPath)
	if os.IsNotExist(err) {
		slog.Debug("Device is not a bridge slave", "devicePath", devicePath)
		return false
	} else if err != nil {
		slog.Warn("Device bridge slave status file cannot be accessed", "devicePath", devicePath, "error", err)
		return false
	}
	return true
}

func GetInterfacesList(netClassDirectory string, portDetectionOptions PortDiscoveryOptions, allowedInterfaceTypes []int) []string {
	resultInterfaces := []string{}

	allInterfaces, err := os.ReadDir(netClassDirectory)
	if err != nil {
		errMsg := fmt.Sprintf("Cannot access netclass directory %s: %v", netClassDirectory, err)
		panic(errMsg)
	}

	reservedNetClassEntry := []string{"bonding_masters"}

	slog.Debug("Got unfiltered interface list", "allInterfaces", allInterfaces)
	for _, deviceDir := range allInterfaces {
		deviceName := deviceDir.Name()

		if slices.Contains(reservedNetClassEntry, deviceName) {
			slog.Debug("Skipping reserved netclass entry", "deviceName", deviceName)
			continue
		}

		if deviceDir.Type().IsRegular() {
			slog.Debug("Not a valid device, skipping", "deviceName", deviceName)
			continue
		}

		if !portDetectionOptions.PortsRegexp.MatchString(deviceName) {
			slog.Debug("Port doesn't match regexp, skipping", "deviceName", deviceName, "regexp", portDetectionOptions.PortsRegexp)
			continue
		}

		interfacePath := path.Join(netClassDirectory, deviceName)
		if !isInterfaceTypeValid(interfacePath, allowedInterfaceTypes) {
			slog.Debug("Not a valid device type, skipping", "deviceName", deviceName)
			continue
		}

		portShouldBeAdded := false
		if portDetectionOptions.DiscoverAllPorts {
			portShouldBeAdded = true
		} else {
			if portDetectionOptions.DiscoverBondSlaves {
				if isInterfaceBondSlave(interfacePath) {
					portShouldBeAdded = true
				} else {
					slog.Debug("Not a bonded port", "deviceName", deviceName)
				}
			}
			if portDetectionOptions.DiscoverBridgeSlaves {
				if isInterfaceBridgeSlave(interfacePath) {
					portShouldBeAdded = true
				} else {
					slog.Debug("Not a bridged port", "deviceName", deviceName)
				}
			}
		}

		if portShouldBeAdded {
			slog.Debug("Port passed one of filters, adding to final list", "deviceName", deviceName)
			resultInterfaces = append(resultInterfaces, deviceName)
		} else {
			slog.Debug("Port haven't passed any of filters, skipping", "deviceName", deviceName)
		}
	}
	slog.Debug("Discovered following interfaces", "interfaces", resultInterfaces)
	return resultInterfaces
}
