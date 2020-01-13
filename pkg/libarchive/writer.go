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

import "C"

// #cgo CFLAGS: -I./include -I/usr/local/opt/libarchive/include
// #cgo LDFLAGS: -larchive
/*
#include <stdlib.h>
#include <archive.h>
#include <archive_entry.h>

*/
import "C"
import (
	"strings"
	"unsafe"

	"github.com/pkg/errors"
)

const (
	bufferSize = 8192
)

type Writer struct {
	ar *C.struct_archive
}

func NewWriter(format Format) (*Writer, error) {
	a := C.archive_write_new()
	if a == nil {
		//die("Couldn't create archive reader.");
		return nil, errors.Errorf("")
	}

	switch format {
	case FormatISO9660:
		if C.archive_write_set_format_iso9660(a) != C.ARCHIVE_OK {
			return nil, errors.Errorf("couldn't set output format")
		}

	default:
		return nil, errors.Errorf("unsupported format %d: ", format)
	}

	return &Writer{
		ar: a,
	}, nil
}

func (r *Writer) Open(filename string) error {
	cstrFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cstrFilename))

	if C.archive_write_open_filename(r.ar, cstrFilename) != C.ARCHIVE_OK {
		return errors.Errorf("couldn't open output archive: %s", filename)
	}

	return nil
}

func (r *Writer) SetOptions(options []string) error {
	cstr := C.CString(strings.Join(options, ","))
	defer C.free(unsafe.Pointer(cstr))
	if C.archive_write_set_options(r.ar, cstr) != C.ARCHIVE_OK {
		return  errors.Errorf("couldn't set options")
	}
	return nil
}

func (r *Writer) Close() error {
	if r.ar != nil && C.archive_write_free(r.ar) != C.ARCHIVE_OK {
		return errors.Errorf("error closing input archive")
	}

	r.ar = nil
	return nil
}

func (r *Writer) WriteEntry(entry *Entry) error {
	if C.archive_write_header(r.ar, entry.entry) != C.ARCHIVE_OK {
		return errors.Errorf("error writing entry header to output archive")
	}

	return nil
}


func (r *Writer) Write(data []byte) (int64, error) {
	var written int64
	len := len(data)
	if len > 0 {
		r := C.archive_write_data(r.ar, unsafe.Pointer(&data[0]), C.ulonglong(len))
		if int(r) != len {
			return written, errors.Errorf("error writing data to output archive")
		}
	}

	return written, nil
}

func (r *Writer) CopyFrom(reader *Reader, size int64) (int64, error) {
	var written int64
	if size > 0 {
		buf := C.malloc(bufferSize)
		defer C.free(buf)

		len := C.archive_read_data(reader.ar, buf, bufferSize)
		for len > 0 {
			if C.archive_write_data(r.ar, buf, C.ulonglong(len)) != len {
				return written, errors.Errorf("error writing data to output archive")
			}

			written += int64(len)
			len = C.archive_read_data(reader.ar, buf, bufferSize)
		}
		if len < 0 {
			return written, errors.Errorf("error reading input archive")
		}
	}

	return written, nil
}
