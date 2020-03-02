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

type netAddressFamily string

func (f netAddressFamily) isV4() bool {
	return strings.ToLower(string(f)) == "ipv4"
}

type netIPAddress struct {
	IPAddress      string
	InterfaceIndex int
	InterfaceAlias string
	AddressFamily  netAddressFamily
	PrefixLength   int
}

type netNAT struct {
	Name                             string
	ExternalIPInterfaceAddressPrefix string
	InternalIPInterfaceAddressPrefix string
	InterfaceGUID                    string `json:"interfaceGuid"`
	InterfaceDescription             string
}

// returns NAT matching to the given filtering condition
func getNetNAT(physical bool, condition string) ([]netNAT, error) {
	cmdlet := []string{"Get-NetNAT"}
	if physical {
		cmdlet = append(cmdlet, "-Physical")
	}
	cmd := []string{strings.Join(cmdlet, " ")}
	if condition != "" {
		cmd = append(cmd, fmt.Sprintf("Where-Object {%s}", condition))
	}
	cmd = append(cmd, "Select-Object -Property InterfaceGuid, InterfaceDescription")
	stdout, err := cmdOut(fmt.Sprintf("ConvertTo-Json @(%s)", strings.Join(cmd, " | ")))
	if err != nil {
		return nil, err
	}

	var netNATObject []netNAT
	err = json.Unmarshal([]byte(strings.TrimSpace(stdout)), &netNATObject)
	if err != nil {
		return nil, err
	}

	return netNATObject, nil
}

func newNATGateway(adapter netAdapter, gatewayIP net.IP, netRange net.IPNet) error {
	ones, _ := netRange.Mask.Size()
	_, err := cmdOut(fmt.Sprintf("New-NetIPAddress -IPAddress \"%s\" -PrefixLength %d -InterfaceIndex %d", gatewayIP.String(), ones, adapter.InterfaceIndex))
	if err != nil {
		return err
	}

	return nil
}

// create NetNAT for specified internal switch
func setupNetNAT(natNetworkName string, ipNet *net.IPNet) error {
	_, err := cmdOut(fmt.Sprintf("New-NetNat -Name \"%s\" -InternalIPInterfaceAddressPrefix \"%s\"", natNetworkName, ipNet.String()))
	if err != nil {
		return err
	}

	return nil
}

// creates a NAT switch of the name
func createNATSwitch(switchName string, cidr string) error {
	err := createVMSwitch(switchName, internalSwitch, netAdapter{})
	if err != nil {
		return errors.Wrapf(err, "failed to create internal switch %s", switchName)
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
		return errors.Errorf( "unable to find adapter for virtual switch %s", switchName)
	}

	err = newNATGateway(netAdapters[0], gatewayIP, *ipNet)
	if err != nil {
		return errors.Wrapf(err, "failed to create NAT gateway for switch %s", switchName)
	}

	err = setupNetNAT(switchName, ipNet)
	if err != nil {
		return errors.Wrapf(err, "failed to setup NAT on switch %s", switchName)
	}

	return nil
}
