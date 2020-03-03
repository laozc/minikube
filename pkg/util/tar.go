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
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func CreateTarArchive(tarFile io.Writer, src string, gzipCompressed bool) error {
	var gzArchive io.WriteCloser
	var tarArchive *tar.Writer
	if gzipCompressed {
		gzArchive = gzip.NewWriter(tarFile)
		defer gzArchive.Close()

		tarArchive = tar.NewWriter(gzArchive)

	} else {
		tarArchive = tar.NewWriter(tarFile)
	}

	defer tarArchive.Close()

	err := filepath.Walk(src, func(filePath string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		header.Name = filepath.ToSlash(strings.TrimPrefix(strings.Replace(filePath, src, "", -1), string(filepath.Separator)))
		if err := tarArchive.WriteHeader(header); err != nil {
			return err
		}

		if !fi.Mode().IsRegular() {
			return nil
		}

		err = func() error {
			f, err := os.Open(filePath)
			if err != nil {
				return err
			}

			defer f.Close()
			if _, err := io.Copy(tarArchive, f); err != nil {
				return err
			}

			return nil
		}()
		return err
	})

	return err
}
