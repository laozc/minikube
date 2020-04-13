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
// #cgo windows LDFLAGS: -static -larchive -lbcrypt -liconv -lexpat -lxml2 -lbz2 -llz4 -llzma -lz -lzstd
// #cgo linux LDFLAGS: -static -larchive -lexpat -lnettle -lxml2 -licui18n -licuuc -licudata -lbz2 -llz4 -llzma -lz -lzstd -lstdc++ -Wl,-Bdynamic -lc -lm -lpthread -ldl -llzo2
// #cgo darwin LDFLAGS: -L/usr/local/opt/libarchive/lib -larchive
//#include <stdlib.h>
//#include <archive.h>
//#include <archive_entry.h>
import "C"
import (
	"os"
	"time"
	"unsafe"
)

type Entry struct {
	entry   *C.struct_archive_entry
	managed bool
}

func NewEntryFromEntryStruct(entry *C.struct_archive_entry) *Entry {
	return &Entry{
		entry:   entry,
		managed: false,
	}
}

func NewEntry() *Entry {
	entry := C.archive_entry_new()
	return &Entry{
		entry:   entry,
		managed: true,
	}
}

func (e *Entry) GetPathName() string {
	return C.GoString(C.archive_entry_pathname(e.entry))
}

func (e *Entry) SetPathName(pathname string) {
	cstr := C.CString(pathname)
	defer C.free(unsafe.Pointer(cstr))
	C.archive_entry_set_pathname(e.entry, cstr)
}

func (e *Entry) GetMode() Mode {
	mode := C.archive_entry_mode(e.entry)
	return Mode(mode)
}

func (e *Entry) SetMode(mode Mode) {
	C.archive_entry_set_mode(e.entry, mode.toC())
}

func (e *Entry) GetFileType() FileType {
	ft := C.archive_entry_filetype(e.entry)
	return FileType(ft)
}

func (e *Entry) SetFileType(ft FileType) {
	C.archive_entry_set_filetype(e.entry, ft.toC())
}

func (e *Entry) GetPermission() os.FileMode {
	perm := C.archive_entry_perm(e.entry)
	return os.FileMode(Mode(perm))
}

func (e *Entry) SetPermission(perm os.FileMode) {
	C.archive_entry_set_perm(e.entry, Mode(perm).toC())
}

func (e *Entry) GetUserID() UserID {
	return UserID(C.archive_entry_uid(e.entry))
}

func (e *Entry) SetUserID(uid UserID) {
	C.archive_entry_set_uid(e.entry, uid.toC())
}

func (e *Entry) SetUserName(uname string) {
	cstr := C.CString(uname)
	defer C.free(unsafe.Pointer(cstr))
	C.archive_entry_set_uname(e.entry, cstr)
}

func (e *Entry) GetGroupID() GroupID {
	return GroupID(C.archive_entry_gid(e.entry))
}

func (e *Entry) SetGroupID(gid GroupID) {
	C.archive_entry_set_gid(e.entry, gid.toC())
}

func (e *Entry) SetGroupName(gname string) {
	cstr := C.CString(gname)
	defer C.free(unsafe.Pointer(cstr))
	C.archive_entry_set_gname(e.entry, cstr)
}

func (e *Entry) Free() {
	if e.managed && e.entry != nil {
		C.archive_entry_free(e.entry)
		e.entry = nil
	}
}

func (e *Entry) GetSize() SSize {
	return SSize(C.archive_entry_size(e.entry))
}

func (e *Entry) SetSize(size SSize) {
	C.archive_entry_set_size(e.entry, size.toC())
}

func (e *Entry) GetAccessTime() *time.Time {
	if C.archive_entry_atime_is_set(e.entry) != 0 {
		return nil
	}

	atime := UnixTime(C.archive_entry_atime(e.entry))
	atimeNanos := Nanosecond(C.archive_entry_atime_nsec(e.entry))
	t := time.Time(NewTimeWithNanosecond(atime, atimeNanos))
	return &t
}

func (e *Entry) SetAccessTime(t *time.Time) {
	if t == nil {
		C.archive_entry_unset_atime(e.entry)
		return
	}

	tm, nsec := TimeWithNanosecond(*t).ToC()
	C.archive_entry_set_atime(e.entry, tm.toC(), nsec.toC())
}

func (e *Entry) GetModifiedTime() *time.Time {
	if C.archive_entry_mtime_is_set(e.entry) != 0 {
		return nil
	}

	atime := UnixTime(C.archive_entry_mtime(e.entry))
	atimeNanos := Nanosecond(C.archive_entry_mtime_nsec(e.entry))
	t := time.Time(NewTimeWithNanosecond(atime, atimeNanos))
	return &t
}

func (e *Entry) SetModifiedTime(t *time.Time) {
	if t == nil {
		C.archive_entry_unset_mtime(e.entry)
		return
	}

	tm, nsec := TimeWithNanosecond(*t).ToC()
	C.archive_entry_set_mtime(e.entry, tm.toC(), nsec.toC())
}
