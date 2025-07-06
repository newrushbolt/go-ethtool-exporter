package interfaces

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
)

func isInterfaceTypeValid(devicePath string, allowedInterfaceTypes []int) bool {
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

func isInterfaceBonded(devicePath string) bool {
	slaveStatePath := path.Join(devicePath, "bonding_slave/state")
	if _, err := os.Stat(slaveStatePath); os.IsNotExist(err) {
		slog.Debug("Device is not a bond slave, skipping", "devicePath", devicePath)
		return false
	} else if err != nil {
		slog.Warn("Device slave status file cannot be accessed", "devicePath", devicePath, "error", err)
		return false
	}

	_, err := os.ReadFile(slaveStatePath)
	if err != nil {
		slog.Warn("Cannot read device bond slave", "devicePath", devicePath, "error", err)
		return false
	}
	return true
}

func GetInterfacesList(netClassDirectory string, detectOnlyBondedPorts bool, allowedInterfaceTypes []int) []string {
	resultInterfaces := []string{}

	allInterfaces, err := os.ReadDir(netClassDirectory)
	if err != nil {
		errMsg := fmt.Sprintf("Cannot access netclass directory %s: %v", netClassDirectory, err)
		panic(errMsg)
	}

	for _, deviceDir := range allInterfaces {
		deviceName := deviceDir.Name()
		if deviceDir.Type().IsRegular() {
			slog.Debug("Not a valid device, skipping", "deviceName", deviceName)
			continue
		}

		interfacePath := path.Join(netClassDirectory, deviceName)
		if !isInterfaceTypeValid(interfacePath, allowedInterfaceTypes) {
			slog.Debug("Not a valid device type, skipping", "deviceName", deviceName)
			continue
		}

		if detectOnlyBondedPorts {
			if !isInterfaceBonded(interfacePath) {
				slog.Debug("Not a bonded port, skipping", "deviceName", deviceName)
				continue
			}
		}
		// TODO: filter out ovs ports by reading link, eg `/sys/class/net/tap2473528a-b1/master -> ../ovs-system`
		resultInterfaces = append(resultInterfaces, deviceName)
	}
	return resultInterfaces
}
