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

func (u UserID) ToC() C.longlong {
	return C.longlong(u)
}

func (g GroupID) ToC() C.longlong {
	return C.longlong(g)
}

func (m Mode) ToC() C.ushort {
	return C.ushort(m)
}

func (s Size) ToC() C.ulong {
	return C.ulong(s)
}

func (s SSize) ToC() C.longlong {
	return C.longlong(s)
}

func (t UnixTime) ToC() C.long {
	return C.long(t)
}

func (t Nanosecond) ToC() C.long {
	return C.long(t)
}
