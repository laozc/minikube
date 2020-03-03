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
)

const (
	ipv4AddressFamily = 2

	dhcpOrigin   = 3
	manualOrigin = 1
)

type netAddressFamily int

func (f netAddressFamily) IsV4() bool {
	return f == ipv4AddressFamily
}

type prefixOrigin int

func (o prefixOrigin) IsDHCP() bool {
	return o == dhcpOrigin
}

func (o prefixOrigin) IsManual() bool {
	return o == manualOrigin
}

type netIPAddress struct {
	IPAddress      string
	InterfaceIndex int
	InterfaceAlias string
	AddressFamily  netAddressFamily
	PrefixLength   int
	PrefixOrigin   prefixOrigin
}

func getNetIPAddresses(condition string) ([]netIPAddress, error) {
	cmd := []string{"Get-NetIPAddress"}
	if condition != "" {
		cmd = append(cmd, fmt.Sprintf("Where-Object {%s}", condition))
	}
	cmd = append(cmd, "Select-Object -Property IPAddress, InterfaceIndex, InterfaceAlias, AddressFamily, PrefixLength, PrefixOrigin")
	stdout, err := cmdOut(fmt.Sprintf("ConvertTo-Json @(%s)", strings.Join(cmd, " | ")))
	if err != nil {
		return nil, err
	}

	var ips []netIPAddress
	err = json.Unmarshal([]byte(strings.TrimSpace(stdout)), &ips)
	if err != nil {
		return nil, err
	}

	return ips, nil
}

// Modifies IP address configuration
func setNetIPAddress(adapter netAdapter, ip net.IP, prefixLength int) error {
	err := cmd(fmt.Sprintf("Set-NetIPAddress -IPAddress \"%s\" -PrefixLength %d -InterfaceIndex %d", ip.String(), prefixLength, adapter.InterfaceIndex))
	return err
}

// Creates and configures an IP address
func newNetIPAddress(adapter netAdapter, ip net.IP, prefixLength int) error {
	err := cmd(fmt.Sprintf("New-NetIPAddress -IPAddress \"%s\" -PrefixLength %d -InterfaceIndex %d", ip.String(), prefixLength, adapter.InterfaceIndex))
	return err
}

// Removes an IP address
func removeNetIPAddress(ip net.IP) error {
	err := cmd(fmt.Sprintf("Remove-NetIPAddress -IPAddress \"%s\" -Confirm:$false", ip.String()))
	return err
}
