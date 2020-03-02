// +build windows

/*
Copyright 2020 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package hyperv

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/pkg/errors"
)

type netNATNetwork struct {
	Name                             string
	ExternalIPInterfaceAddressPrefix string
	InternalIPInterfaceAddressPrefix string
	InterfaceGUID                    string `json:"interfaceGuid"`
	InterfaceDescription             string
}

// returns NAT network matching to the given filtering condition
func getNetNAT(condition string) ([]netNATNetwork, error) {
	cmd := []string{"Get-NetNAT"}
	if condition != "" {
		cmd = append(cmd, fmt.Sprintf("Where-Object {%s}", condition))
	}
	cmd = append(cmd, "Select-Object -Property Name, ExternalIPInterfaceAddressPrefix, InternalIPInterfaceAddressPrefix, interfaceGuid, InterfaceDescription")
	stdout, err := cmdOut(fmt.Sprintf("ConvertTo-Json @(%s)", strings.Join(cmd, " | ")))
	if err != nil {
		return nil, err
	}

	var netNATObjects []netNATNetwork
	err = json.Unmarshal([]byte(strings.TrimSpace(stdout)), &netNATObjects)
	if err != nil {
		return nil, err
	}

	return netNATObjects, nil
}

// removes NAT network configuration
func removeNetNATNetwork(natNetworkName string) error {
	err := cmd(fmt.Sprintf("Remove-NetNat -Name \"%s\"", natNetworkName))
	return err
}

// create NAT network for specified CIDR notation
func createNetNATNetwork(natNetworkName string, ipNet net.IPNet) error {
	err := cmd(fmt.Sprintf("New-NetNat -Name \"%s\" -InternalIPInterfaceAddressPrefix \"%s\"", natNetworkName, ipNet.String()))
	return err
}

// ensures gateway for the NAT network exists and its configuration is correct
func ensureNATGateway(adapter netAdapter, gatewayIP net.IP, netRange net.IPNet) error {
	ips, err := getNetIPAddresses(fmt.Sprintf("($_.IPAddress -eq \"%s\")", gatewayIP.String()))
	if err != nil {
		return err
	}

	prefixLength, _ := netRange.Mask.Size()
	if len(ips) > 0 {
		ip := ips[0]
		if ip.InterfaceIndex != adapter.InterfaceIndex || ip.PrefixLength != prefixLength {
			err = setNetIPAddress(adapter, gatewayIP, prefixLength)
			if err != nil {
				return errors.Wrapf(err, "failed to update IP configuration for NAT gateway on adapter %s", adapter.InterfaceDescription)
			}
		}

	} else {
		err = newNetIPAddress(adapter, gatewayIP, prefixLength)
		if err != nil {
			return errors.Wrapf(err, "failed to create NAT gateway on adapter %s", adapter.InterfaceDescription)
		}
	}

	return nil
}

// ensure NAT network exist and its configuration is correct
func ensureNetNATNetwork(natNetworkName string, netRange net.IPNet) error {
	cidr := netRange.String()
	natNetworks, err := getNetNAT(fmt.Sprintf("($_.Name -eq \"%s\") -Or ($_.InternalIPInterfaceAddressPrefix -eq \"%s\")", natNetworkName, cidr))
	if err != nil {
		return err
	}

	if len(natNetworks) > 0 {
		if natNetworks[0].InternalIPInterfaceAddressPrefix == cidr {
			return nil
		}

		err := removeNetNATNetwork(natNetworkName)
		if err != nil {
			return errors.Wrapf(err, "failed to remove NAT network %s for re-creating", natNetworkName)
		}
	}

	err = createNetNATNetwork(natNetworkName, netRange)
	if err != nil {
		return errors.Wrapf(err, "failed to create NAT network %s for the specified CIDR %s", natNetworkName, cidr)
	}

	return nil
}

// ensure NAT switch exist and its configuration is correct
func ensureNATSwitch(switchName string, natNetworkName string, cidr string) error {
	foundSwitches, err := getVMSwitch(fmt.Sprintf("($_.Name -eq \"%s\") -And ($_.SwitchType -eq 1)", switchName))
	if err != nil {
		return err
	}

	if len(foundSwitches) < 1 {
		// create a new internal switch if not exist
		err := createVMSwitch(switchName, internalSwitch, netAdapter{})
		if err != nil {
			return errors.Wrapf(err, "failed to create internal switch %s", switchName)
		}
	}

	// set up NAT for internal switch
	gatewayIP, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return errors.Wrapf(err, "specified CIDR %s is not valid", cidr)
	}

	adapterName := fmt.Sprintf("vEthernet (%s)", switchName)
	netAdapters, err := getNetAdapters(false, fmt.Sprintf("$_.Name -eq \"%s\"", adapterName))
	if err != nil {
		return err
	}

	if len(netAdapters) < 1 {
		return errors.Errorf("unable to find adapter for virtual switch %s", switchName)
	}

	err = ensureNATGateway(netAdapters[0], gatewayIP, *ipNet)
	if err != nil {
		return errors.Wrapf(err, "failed to ensure NAT gateway for switch %s", switchName)
	}

	err = ensureNetNATNetwork(natNetworkName, *ipNet)
	if err != nil {
		return errors.Wrapf(err, "failed to setup NAT on switch %s", switchName)
	}

	return nil
}
