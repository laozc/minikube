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

package metadata

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

var systemdNetworkdIPv4ConfigTemplate = template.Must(template.New("systemdNetworkdIPv4ConfigTemplate").Parse(`[Match]
Name={{ .NetworkInterface }}

[Network]
Address={{ .IPAddress }}
{{ if not .GatewayIP }}Gateway={{ .GatewayIP }}{{- end}}
{{- range $s := .DNS}}
DNS={{ $s -}}
{{end}}
IPv6AcceptRA={{ .IPv6AcceptRA }}
`))

var networkInitScript = `#!/bin/bash
cp *.network /etc/systemd/network/

echo "Restarting network..."
systemctl restart systemd-networkd
`

const (
	networkConfigPriority = 10
)

func generateNetworkConfig(dir string, md Metadata) error {
	d := filepath.Join(dir, "network")

	err := os.Mkdir(d, 0755)
	if err != nil {
		return errors.Wrapf(err, "failed to create dir %s", d)
	}

	for ifName, c := range md.Networks {
		acceptRA := "yes"
		if c.ForceIPv4 {
			acceptRA = "no"
		}

		data := struct {
			NetworkInterface string
			IPAddress        string
			GatewayIP        string
			DNS              []string
			IPv6AcceptRA     string
		}{
			NetworkInterface: ifName,
			IPAddress:        fmt.Sprintf("%s/%s", c.MachineIP, c.Netmask),
			GatewayIP:        c.GatewayIP,
			DNS:              c.DNS,
			IPv6AcceptRA:     acceptRA,
		}

		var buf bytes.Buffer
		err := systemdNetworkdIPv4ConfigTemplate.Execute(&buf, data)
		if err != nil {
			return errors.Wrapf(err, "failed to generate systemd-networkd configuration")
		}

		fn := fmt.Sprintf("%2d-%s.network", networkConfigPriority, ifName)
		err = ioutil.WriteFile(filepath.Join(d, fn), buf.Bytes(), 0644)
		if err != nil {
			return errors.Wrapf(err, "failed to write systemd-networkd configuration")
		}

		glog.V(4).Infof("Created systemd-networkd config %s: %s", fn, buf.String())
	}

	err = ioutil.WriteFile(filepath.Join(d, "init.sh"), []byte(networkInitScript), 0744)
	if err != nil {
		return errors.Wrapf(err, "failed to write network init script")
	}

	return nil
}
