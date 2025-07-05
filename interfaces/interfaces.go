package interfaces

import (
	"fmt"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
)

func isInterfaceTypeValid(devicePath string) bool {
	typePath := path.Join(devicePath, "type")
	allowedInterfaceTypes := []int{
		// 1 is ethernet
		1,
		// Other protocol types
		// https://github.com/torvalds/linux/blob/master/include/uapi/linux/if_arp.h#L29
	}
	interfaceTypeRaw, err := os.ReadFile(typePath)
	if err != nil {
		fmt.Printf("Cannot read device type for device: %s\n", err)
		return false
	}
	interfaceType, err := strconv.Atoi(strings.TrimSpace(string(interfaceTypeRaw)))
	if err != nil {
		fmt.Printf("Cannot parse device type for device: %v\n", err)
		return false
	}
	if !slices.Contains(allowedInterfaceTypes, interfaceType) {
		fmt.Printf("Interface type <%d> is not allowed, must be on of %v\n", interfaceType, allowedInterfaceTypes)
		return false
	}
	return true
}

func isInterfaceBonded(devicePath string) bool {
	slaveStatePath := path.Join(devicePath, "bonding_slave/state")
	_, err := os.ReadFile(slaveStatePath)
	if err != nil {
		fmt.Printf("Cannot read slave state for device: %s\n", err)
		fmt.Printf("Device with path <%s> is not bond slave, skipping\n", devicePath)
		return false
	}
	return true
}

func GetInterfacesList(netClassDirectory string, detectOnlyBondedPorts bool) []string {
	resultInterfaces := []string{}

	allInterfaces, err := os.ReadDir(netClassDirectory)
	if err != nil {
		errMsg := fmt.Sprintf("Cannot access netclass directory %s: %v", netClassDirectory, err)
		panic(errMsg)
	}

	for _, deviceDir := range allInterfaces {
		deviceName := deviceDir.Name()
		if deviceDir.Type().IsRegular() {
			fmt.Printf("<%s> is not a valid device, skipping\n", deviceName)
			continue
		}

		interfacePath := path.Join(netClassDirectory, deviceName)
		if !isInterfaceTypeValid(interfacePath) {
			fmt.Printf("<%s> is not a valid device, skipping\n", deviceName)
			continue
		}

		if detectOnlyBondedPorts {
			if !isInterfaceBonded(interfacePath) {
				continue
			}
		}
		// TODO: filter out ovs ports by reading link, eg `/sys/class/net/tap2473528a-b1/master -> ../ovs-system`
		resultInterfaces = append(resultInterfaces, deviceName)
	}
	return resultInterfaces
}
