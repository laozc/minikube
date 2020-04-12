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

package util

import (
	"github.com/pkg/errors"
	"net"
)

// guess next IP of the sub network
func IncrementIP(startIP net.IP, ipNet net.IPNet) (net.IP, error) {
	ip := make(net.IP, len(startIP))
	copy(ip, startIP)
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] != 0 {
			break
		}
	}
	if !ipNet.Contains(ip) {
		return nil, errors.New("overflowed CIDR while incrementing IP")
	}
	return ip, nil
}

func GetIPNetRange(ipNet net.IPNet) (net.IP, net.IP, net.IP, error) {
	networkStart := make(net.IP, len(ipNet.IP))
	copy(networkStart, ipNet.IP)

	ones, sizes := ipNet.Mask.Size()
	zeroes := sizes - ones
	networkBits := zeroes % 8
	byteOffset := zeroes / 8
	var hostBitsMask byte
	for i := 0; i < networkBits; i += 1 {
		hostBitsMask = hostBitsMask | (1 << i)
	}
	lastNetworkBitMask := byte((1 << networkBits) - 1)

	var bytePos int
	if networkStart.To4() != nil {
		// IPv4
		bytePos = len(networkStart) - 1 - byteOffset

	} else {
		// IPv6
		bytePos = len(networkStart) - byteOffset
	}

	// clear the host bits
	networkStart[bytePos] = networkStart[bytePos] & ^hostBitsMask

	// clear the rest host bits to 0
	for i := len(networkStart) - 1; i > bytePos; i-- {
		networkStart[i] = 0
	}

	ipStart := make(net.IP, len(networkStart))
	copy(ipStart, networkStart)
	ipStart[len(ipStart)-1]++
	if !ipNet.Contains(ipStart) {
		return nil, nil, nil, errors.New("overflowed CIDR for start IP")
	}

	ipBroadcast := make(net.IP, len(networkStart))
	copy(ipBroadcast, networkStart)
	ipBroadcast[bytePos] = ipBroadcast[bytePos] | lastNetworkBitMask
	// set the rest host bits to 0xFF
	for i := len(ipBroadcast) - 1; i > bytePos; i-- {
		ipBroadcast[i] = 255
	}

	ipEnd := make(net.IP, len(ipBroadcast))
	copy(ipEnd, ipBroadcast)
	ipEnd[len(ipEnd)-1]--
	if !ipNet.Contains(ipEnd) {
		return nil, nil, nil, errors.New("overflowed CIDR for end IP")
	}

	return ipStart, ipEnd, ipBroadcast, nil
}

func AllocateIPs(cidr string) (net.IP, net.IPMask, net.IP, net.IP, net.IP, error) {
	gatewayIP, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return net.IP{}, net.IPMask{}, net.IP{}, net.IP{}, net.IP{}, errors.Wrapf(err, "specified CIDR %s is not valid", cidr)
	}

	guestIP, err := IncrementIP(gatewayIP, *ipNet)
	if err != nil {
		return net.IP{}, net.IPMask{}, net.IP{}, net.IP{}, net.IP{}, errors.Wrapf(err, "unable to allocate IP for guest VM")
	}

	loadBalancerStartIP, err := IncrementIP(guestIP, *ipNet)
	if err != nil {
		return net.IP{}, net.IPMask{}, net.IP{}, net.IP{}, net.IP{}, errors.Wrapf(err, "unable to allocate start IP for load balancer")
	}

	_, endIP, _, err := GetIPNetRange(*ipNet)
	if err != nil {
		return net.IP{}, net.IPMask{}, net.IP{}, net.IP{}, net.IP{}, errors.Wrapf(err, "unable to allocate end IP for load balancer")
	}

	loadBalancerEndIP := endIP
	return gatewayIP, ipNet.Mask, guestIP, loadBalancerStartIP, loadBalancerEndIP, nil
}
