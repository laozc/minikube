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
import (
	"os"
	"time"
)

const (
	ModeMask = 0170000 /* These bits determine file type.  */

	ModeDirectory       = 0040000 /* Directory.  */
	ModeCharacterDevice = 0020000 /* Character device.  */
	ModeBlockDevice     = 0060000 /* Block device.  */
	ModeRegularFile     = 0100000 /* Regular file.  */
	ModeFIFO            = 0010000 /* FIFO.  */
	ModeSymbolicLink    = 0120000 /* Symbolic link.  */
	ModeSocket          = 0140000 /* Socket.  */
)

const (
	FormatISO9660 = iota
)

const (
	bufferSize = 8192
)

type Mode uint32

func (m Mode) IsRegular() bool {
	return m.GetFileType().IsRegular()
}

func (m Mode) GetFileType() FileType {
	return FileType(m & ModeMask)
}

func (m Mode) GetPermission() os.FileMode {
	n := ^ModeMask
	return os.FileMode(uint32(n) & uint32(m))
}

type FileType uint32

func (ft FileType) IsRegular() bool {
	return ft == ModeRegularFile
}

func (ft FileType) ToC() C.uint {
	return C.uint(ft)
}

type UserID int64

type GroupID int64

type Size uint64

type SSize int64

type Format int

type UnixTime int64

type Nanosecond int64

type TimeWithNanosecond time.Time

func (tn TimeWithNanosecond) ToC() (UnixTime, Nanosecond) {
	t := time.Time(tn)
	return UnixTime(t.Unix()), Nanosecond(t.Nanosecond())
}

func NewTimeWithNanosecond(t UnixTime, ns Nanosecond) TimeWithNanosecond {
	return TimeWithNanosecond(time.Unix(int64(t), int64(ns)))
}
