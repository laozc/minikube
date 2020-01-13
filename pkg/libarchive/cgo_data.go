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

/*
#include <stdlib.h>
#include <archive.h>
#include <archive_entry.h>

*/
import "C"
import "time"

func toGoTime(atime C.longlong, atimeNanos C.long) time.Time {
	return time.Unix(int64(atime), int64(atimeNanos))
}

func fromGoTime(t time.Time) (C.longlong, C.long) {
	return C.longlong(t.Unix()), C.long(t.Nanosecond())
}
