// Copyright © 2019, Lukas Werling
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

package c4group

import "time"

const (
	C4GzMagic1      = 0x1e
	C4GzMagic2      = 0x8c
	C4GroupFileID   = "RedWolf Design GrpFolder"
	C4GroupFileVer1 = 1
	C4GroupFileVer2 = 2
)

const (
	HeaderSize = 204 // size of the header in byte
	EntrySize  = 316 // size of each entry in byte
)

const originalMagic = 1234567

// Header on-disk format (C4GroupHeader)
type header struct {
	ID         [24 + 4]byte
	Ver1, Ver2 int32
	Entries    int32
	Author     [32]byte // reserved in OpenClonk
	_          [32]byte
	Ctime      int32 // creation time, reserved in OpenClonk
	Original   int32 // 1234567 if original pack, reserved in OpenClonk
	_          [92]byte
}

// Entry header on-disk format (C4GroupEntryCore)
type entry struct {
	Filename        [260]byte
	Packed          int32 // reserved in OpenClonk
	ChildGroup      int32
	Size, _, Offset int32
	Mtime           int32 // modification time, reserved in OpenClonk
	HasCRC          byte
	CRC             uint32
	Executable      byte
	_               [26]byte
}

func memScramble(buffer []byte) {
	for i := range buffer {
		buffer[i] ^= 237
	}
	for i := 0; (i + 2) < len(buffer); i += 3 {
		buffer[i], buffer[i+2] = buffer[i+2], buffer[i]
	}
}

// Header contains the public C4GroupHeader fields.
type Header struct {
	Entries    int32
	Author     string    // reserved in OpenClonk
	Ctime      time.Time // creation time, reserved in OpenClonk
	IsOriginal bool      // reserved in OpenClonk
}

// Entry contains the public C4GroupEntryCore fields.
type Entry struct {
	Filename   string
	IsGroup    bool
	Size       int
	Mtime      time.Time // modification time, reserved in openclonk
	Executable bool
}

// i2b converts an integer to a boolean.
func i2b(i int) bool {
	if i == 0 {
		return false
	}
	return true
}

// b2i converts a boolean to an integer.
func b2i(b bool) int8 {
	if b {
		return 1
	}
	return 0
}

// clen returns the length of a null-terminated string.
func clen(n []byte) int {
	for i := 0; i < len(n); i++ {
		if n[i] == 0 {
			return i
		}
	}
	return len(n)
}

// publicHeader adapts Header fields from file format to Go API format.
func publicHeader(public *Header, private *header) {
	public.Entries = private.Entries
	public.Author = string(private.Author[:clen(private.Author[:])])
	public.Ctime = time.Unix(int64(private.Ctime), 0)
	public.IsOriginal = private.Original == originalMagic
}

// privateHeader adapts header fields from Go API format to file format.
func privateHeader(private *header, public *Header) {
	private.Entries = public.Entries
	copy(private.Author[:], []byte(public.Author))
	private.Ctime = int32(public.Ctime.Unix())
	if public.IsOriginal {
		private.Original = originalMagic
	} else {
		private.Original = 0
	}
}

// publicEntry adapts Entry fields from file format to Go API format.
func publicEntry(public *Entry, private *entry) {

	public.Filename = string(private.Filename[:clen(private.Filename[:])])
	public.IsGroup = i2b(int(private.ChildGroup))
	public.Size = int(private.Size)
	public.Mtime = time.Unix(int64(private.Mtime), 0)
	public.Executable = i2b(int(private.Executable))
}

// privateHeader adapts Header fields from Go API format to file format.
func privateEntry(private *entry, public *Entry) {
	copy(private.Filename[:], public.Filename)
	private.ChildGroup = int32(b2i(public.IsGroup))
	private.Size = int32(public.Size)
	private.Mtime = int32(public.Mtime.Unix())
	private.Executable = byte(b2i(public.Executable))
}
