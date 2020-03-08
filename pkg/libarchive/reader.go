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

// #cgo darwin CFLAGS: -I/usr/local/opt/libarchive/include
/*
#include <stdlib.h>
#include <string.h>
#include <archive.h>
#include <archive_entry.h>
*/
import "C"
import (
	"io"
	"unsafe"

	"github.com/pkg/errors"
)

type Reader struct {
	ar *C.struct_archive
}

func NewReader() *Reader {
	return &Reader{}
}

func (r *Reader) Open(filename string) error {
	a := C.archive_read_new()
	if a == nil {
		return errors.Errorf("couldn't create archive reader.")
	}

	r.ar = a

	if C.archive_read_support_filter_all(a) != C.ARCHIVE_OK {
		return errors.Errorf("couldn't enable decompression")
	}

	if C.archive_read_support_format_all(a) != C.ARCHIVE_OK {
		return errors.Errorf("couldn't enable all supported format")
	}

	cstrFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cstrFilename))

	if C.archive_read_open_filename(r.ar, cstrFilename, 4096) != C.ARCHIVE_OK {
		return errors.Errorf("couldn't open input archive %s", filename)
	}

	return nil
}

func (r *Reader) Read(size Size) ([]byte, error) {
	var result []byte
	if size > 0 {
		buf := C.malloc(bufferSize)
		defer C.free(buf)

		var readBytes Size
		readBytes = bufferSize
		if size < readBytes {
			readBytes = size
		}

		len := SSize(C.archive_read_data(r.ar, buf, readBytes.ToC()))
		for len > 0 {
			goBuf := make([]byte, readBytes)
			C.memcpy(unsafe.Pointer(&goBuf[0]), buf, readBytes.ToC())
			result = append(result, goBuf...)
		}
		if len < 0 {
			return result, errors.Errorf("error reading input archive")
		}
	}

	return result, nil
}

func (r *Reader) CopyTo(writer io.Writer, size SSize) (SSize, error) {
	var written SSize
	if size > 0 {
		buf := C.malloc(bufferSize)
		defer C.free(buf)

		len := SSize(C.archive_read_data(r.ar, buf, bufferSize))
		for len > 0 {
			goBuf := make([]byte, len)
			C.memcpy(unsafe.Pointer(&goBuf[0]), buf, Size(len).ToC())

			_, err := writer.Write(goBuf)
			if err != nil {
				return written, errors.Errorf("error writing data")
			}

			written += len
			len = SSize(C.archive_read_data(r.ar, buf, bufferSize))
		}
		if len < 0 {
			return written, errors.Errorf("error reading input archive")
		}
	}

	return written, nil
}

func (r *Reader) NextEntry() (*Entry, error) {
	var entry *C.struct_archive_entry
	switch C.archive_read_next_header(r.ar, &entry) {
	case C.ARCHIVE_OK:
		return NewEntryFromEntryStruct(entry), nil

	case C.ARCHIVE_EOF:
		return nil, nil

	default:
		return nil, errors.Errorf("unable to read from archive")
	}
}

func (r *Reader) Close() error {
	if r.ar != nil && C.archive_read_free(r.ar) != C.ARCHIVE_OK {
		return errors.Errorf("error closing input archive")
	}

	r.ar = nil
	return nil
}
