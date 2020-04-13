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

package libarchive

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

func CreateISO(dest string, files map[string]string, options []string) error {
	// creates an empty ISO
	tmpName := fmt.Sprintf("%s.%s", dest, time.Now().Format("20060102150405"))
	err := func() error {
		output, err := NewWriter(FormatISO9660)
		if err != nil {
			return errors.Wrapf(err, "unable to open archive %s for writing", dest)
		}

		defer output.Close()

		err = output.Open(tmpName)
		if err != nil {
			return err
		}

		return nil
	}()
	if err != nil {
		return errors.Wrapf(err, "failed to create ISO archive %s", dest)
	}

	defer os.Remove(tmpName)

	return PatchISO(tmpName, dest, files, options)
}

func ExtractISO(source string, dest string) error {
	input := NewReader()
	err := input.Open(source)
	if err != nil {
		return errors.Wrapf(err, "unable to open archive %s for reading", source)
	}

	defer input.Close()

	for {
		e, err := input.NextEntry()
		if err != nil {
			return errors.Wrapf(err, "failed to get entry from %s", source)
		}

		if e == nil {
			break
		}

		if !e.GetMode().IsRegular() {
			continue
		}

		destPath := filepath.Join(dest, e.GetPathName())
		err = os.MkdirAll(filepath.Dir(destPath), 0755)
		if err != nil {
			return err
		}

		f, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE, e.GetMode().GetPermission())
		if err != nil {
			return errors.Wrapf(err, "failed to open file %s", destPath)
		}

		_, err = input.CopyTo(f, e.GetSize())
		if err != nil {
			return errors.Wrapf(err, "failed to copy entry %s from %s", e.GetPathName(), source)
		}

		glog.V(4).Infof("Extracted file %s with mode %d", destPath, e.GetMode()&ModeMask)
	}

	return nil
}

func PatchISO(source string, dest string, files map[string]string, options []string) error {
	input := NewReader()
	err := input.Open(source)
	if err != nil {
		return errors.Wrapf(err, "unable to open archive %s for reading", source)
	}

	defer input.Close()

	output, err := NewWriter(FormatISO9660)
	if err != nil {
		return errors.Wrapf(err, "unable to open archive %s for writing", dest)
	}

	defer output.Close()

	err = output.SetOptions(options)
	if err != nil {
		return err
	}

	err = output.Open(dest)
	if err != nil {
		return err
	}

	for {
		e, err := input.NextEntry()
		if err != nil {
			return errors.Wrapf(err, "failed to get entry from %s", source)
		}

		if e == nil {
			break
		}

		err = output.WriteEntry(e)
		if err != nil {
			return errors.Wrapf(err, "failed to write entry %s", e.GetPathName())
		}

		glog.V(4).Infof("Wrote file %s with mode %d", e.GetPathName(), e.GetMode()&ModeMask)
		_, err = output.CopyFrom(input, e.GetSize())
		if err != nil {
			return errors.Wrapf(err, "failed to copy entry %s from %s", e.GetPathName(), source)
		}
	}

	for source, dest := range files {
		stat, err := os.Stat(source)
		if err != nil {
			return err
		}

		if !stat.Mode().IsRegular() {
			return errors.Errorf("%s is not a regular file", source)
		}

		buf := make([]byte, 8192)
		err = func() error {
			e := NewEntry()
			defer e.Free()

			e.SetPathName(dest)
			e.SetFileType(ModeRegularFile)
			e.SetPermission(stat.Mode() & os.ModePerm)

			e.SetSize(SSize(stat.Size()))
			mt := stat.ModTime()
			e.SetModifiedTime(&mt)
			e.SetAccessTime(&mt)

			err = output.WriteEntry(e)
			if err != nil {
				return err
			}

			f, err := os.Open(source)
			if err != nil {
				return err
			}

			n, err := f.Read(buf)
			if err != nil {
				return errors.Wrapf(err, "unable to read data from %s", source)
			}

			_, err = output.Write(buf[0:n])
			if err != nil {
				return err
			}

			return nil
		}()
		if err != nil {
			return errors.Wrapf(err, "failed to copy data from %s", source)
		}
	}

	return nil
}
