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
	"io/ioutil"
	"net"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"k8s.io/minikube/pkg/util"
)

const (
	DefaultNetworkInterface = "eth0"
)

type Network struct {
	MachineIPNet net.IPNet
	GatewayIP    net.IP
	DNS          []string
	ForceIPv4    bool
}

type Metadata struct {
	Networks map[string]Network
}

var metadataInitScript = `#!/bin/bash
for f in *; do
  if [[ -d "$f" ]] && [[ -f "$f/init.sh" ]]; then
    pushd "$f" > /dev/null && chmod +x ./init.sh && . ./init.sh && popd > /dev/null
  fi
done
`

func CreateMetadataTar(fp string, md Metadata) error {
	tmpDir, err := ioutil.TempDir("", "metadata")
	if err != nil {
		return err
	}

	defer os.RemoveAll(tmpDir)

	err = generateNetworkConfig(tmpDir, md)
	if err != nil {
		return errors.Wrapf(err, "failed to generate network config")
	}

	err = ioutil.WriteFile(filepath.Join(tmpDir, "init.sh"), []byte(metadataInitScript), 0744)
	if err != nil {
		return errors.Wrapf(err, "failed to write metadata init script")
	}

	f, err := os.OpenFile(fp, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return errors.Wrapf(err, "failed to open file %s", fp)
	}

	defer f.Close()

	err = util.CreateTarArchive(f, tmpDir, false)
	if err != nil {
		return errors.Wrapf(err, "failed to create metadata tar file %s", fp)
	}

	return nil
}
