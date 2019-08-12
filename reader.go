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

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"io"
	"io/ioutil"
)

var (
	ErrInvalidMagic  error = errors.New("c4group: invalid magic bytes")
	ErrInvalidHeader error = errors.New("c4group: header fields (id or version) invalid")
	ErrNoChildGroup  error = errors.New("c4group: entry is not a child group")
	ErrAlreadyRead   error = errors.New("c4group: entry has already been read, cannot read child group")
)

// Reader provides read access to c4group archives.
type Reader struct {
	Header  Header  // valid after NewReader
	Entries []Entry // same

	r       io.Reader
	gz      *gzip.Reader
	offset  int // offset after all headers
	curFile int // index of current file
	entries []entry
}

// magicBytesReader is an adapter for the c4group magic bytes to gzip magic bytes.
type magicBytesReader struct {
	r              io.Reader
	readMagicBytes int
}

const (
	gzipID1 = 0x1f
	gzipID2 = 0x8b
)

func (mr *magicBytesReader) Read(b []byte) (int, error) {
	read, err := mr.r.Read(b)
	if err != nil {
		return read, err
	}
	if read > 0 && mr.readMagicBytes < 2 {
		// We might read only a single byte, so handle both separately.
		alreadyRead := mr.readMagicBytes
		if alreadyRead == 0 {
			if b[0] != C4GzMagic1 {
				return read, ErrInvalidMagic
			}
			b[0] = gzipID1
			mr.readMagicBytes++
		}
		if alreadyRead+read >= 2 {
			if b[1-alreadyRead] != C4GzMagic2 {
				return read, ErrInvalidMagic
			}
			b[1] = gzipID2
			mr.readMagicBytes++
		}
	}
	return read, err
}

// NewReader creates a new c4group reader reading from r.
func NewReader(r io.Reader) (*Reader, error) {
	mr := &magicBytesReader{r: r}
	gz, err := gzip.NewReader(mr)
	if err != nil {
		return nil, err
	}
	cr := &Reader{r: gz, gz: gz}
	if err = cr.init(); err != nil {
		return nil, err
	}
	return cr, nil
}

// init initialized the reader by reading the header structures.
func (cr *Reader) init() error {
	cr.curFile = -1
	// Read main header.
	if err := cr.readHeader(); err != nil {
		return err
	}
	// Read all entry headers.
	cr.Entries = make([]Entry, cr.Header.Entries)
	cr.entries = make([]entry, cr.Header.Entries)
	for i := int32(0); i < cr.Header.Entries; i++ {
		err := cr.readEntry(&cr.entries[i])
		publicEntry(&cr.Entries[i], &cr.entries[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// readHeader reads the initial group header.
func (cr *Reader) readHeader() error {
	header := header{}
	headerSize := binary.Size(&header)
	var buf bytes.Buffer
	_, err := io.CopyN(&buf, cr.r, int64(headerSize))
	if err != nil {
		return err
	}
	memScramble(buf.Bytes())
	err = binary.Read(&buf, binary.LittleEndian, &header)
	if err != nil {
		return err
	}

	if string(header.ID[:len(C4GroupFileID)]) != C4GroupFileID || header.Ver1 != C4GroupFileVer1 || header.Ver2 != C4GroupFileVer2 {
		return ErrInvalidHeader
	}

	publicHeader(&cr.Header, &header)

	return nil
}

// readEntry reads a single entry header.
func (cr *Reader) readEntry(e *entry) error {
	err := binary.Read(cr.r, binary.LittleEndian, e)
	if err != nil {
		return err
	}
	return nil
}

// Next skips to the next file.
//
// Returns io.EOF if all files have been read.
func (cr *Reader) Next() (*Entry, error) {
	// curFile is initialized to -1
	if cr.curFile+1 >= len(cr.entries) {
		return nil, io.EOF
	}
	cr.curFile++
	entry := &cr.entries[cr.curFile]
	// Skip to the file's data.
	n, err := io.CopyN(ioutil.Discard, cr.r, int64(entry.Offset)-int64(cr.offset))
	if err != nil {
		return nil, err
	}
	cr.offset += int(n)
	return &cr.Entries[cr.curFile], nil
}

// Read from the current file. Returns io.EOF after finishing.
func (cr *Reader) Read(b []byte) (int, error) {
	entry := &cr.entries[cr.curFile]
	if cr.offset >= int(entry.Offset+entry.Size) {
		return 0, io.EOF
	}
	// Ensure that we don't read into the next file.
	max := int(entry.Size) - (cr.offset - int(entry.Offset))
	if len(b) > max {
		b = b[:max]
	}
	n, err := cr.r.Read(b)
	cr.offset += n
	return n, err
}

// ReadGroup reads a sub group from the archive.
func (cr *Reader) ReadGroup() (*Reader, error) {
	entry := &cr.entries[cr.curFile]
	if entry.ChildGroup == 0 {
		return nil, ErrNoChildGroup
	}
	if int(entry.Offset) != cr.offset {
		return nil, ErrAlreadyRead
	}
	sub := &Reader{r: cr}
	if err := sub.init(); err != nil {
		return nil, err
	}
	return sub, nil
}

// Close closes the Reader.
func (cr *Reader) Close() error {
	// Sub group readers don't decompress.
	if cr.gz != nil {
		return cr.gz.Close()
	}
	return nil
}
