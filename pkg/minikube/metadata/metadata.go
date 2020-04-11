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
	"fmt"
	"io/ioutil"
	"k8s.io/minikube/pkg/minikube/config"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/golang/glog"
	"github.com/pkg/errors"

	"k8s.io/minikube/pkg/util"
)

type Network struct {
	MachineIP string
	Netmask   string
	GatewayIP string
	DNS       []string
	ForceIPv4 bool
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

func CreateMetadataTar(fp string, cfg config.MachineConfig) error {
	tmpDir, err := ioutil.TempDir("", "metadata")
	if err != nil {
		return err
	}

	defer os.RemoveAll(tmpDir)

	err = generateNetworkConfig(tmpDir, cfg)
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

func getISOArchiverBinaryName() string {
	name := fmt.Sprintf("iso-archiver-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return name
}

func SupportsPatch() (bool, error) {
	isoArchiverBinary := getISOArchiverBinaryName()
	_, err := exec.LookPath(isoArchiverBinary)
	if err != nil {
		glog.V(4).Infof("Unable to find platform-specific iso-archiver binary %s.", isoArchiverBinary)
	}
	return err == nil, err
}

func PatchISO(isoPath, baseDir, outISO string, options []string) error {
	_, err := SupportsPatch()
	if err != nil {
		return errors.Wrapf(err, "Unable to find platform-specific iso-archiver binary %s", getISOArchiverBinaryName)
	}

	glog.V(4).Infof("Patching ISO %s...", isoPath)
	binary := getISOArchiverBinaryName()
	cmd := exec.Command(binary, "patch", "--source", isoPath, "--base-dir", baseDir, "--out", outISO, "--options", strings.Join(options, ","))
	err = cmd.Run()
	if err != nil {
		return errors.Wrapf(err, "failed to patch ISO")
	}

	return nil
}
