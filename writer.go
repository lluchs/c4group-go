package c4group

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"io"
)

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

// Header on-disk format (C4GroupHeader)
type header struct {
	id         [24 + 4]byte
	ver1, ver2 int32
	entries    int32
	reserved   [164]byte
}

// Entry header on-disk format (C4GroupEntryCore)
type entry struct {
	filename                [260]byte
	packed, childGroup      int32
	size, reserved1, offset int32
	reserved2               int32
	reserved3               byte
	reserved4               uint32
	executable              byte
	reserved5               [26]byte
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
	Entries int32
}

// Entry contains the public C4GroupEntryCore fields.
type Entry struct {
	Filename   string
	IsGroup    bool
	Size       int
	Executable bool
}

var (
	ErrHeaderAlreadyWritten error = errors.New("c4group: header already written")
	ErrNoHeader             error = errors.New("c4group: initial header missing")
	ErrTooManyEntries       error = errors.New("c4group: more entries than specified in header")
	ErrNotEnoughEntries     error = errors.New("c4group: not enough entry headers written")
	ErrTooMuchWritten       error = errors.New("c4group: too much file data")
	ErrNotEnoughWritten     error = errors.New("c4group: not enough file data")
)

// Writer provides sequential writing writing of c4group archives.
type Writer struct {
	w               io.Writer
	gz              *gzip.Writer
	offset          int32 // current file offset, incremented with each entry header
	haveHeader      bool  // header already written?
	expectedEntries int32 // number of entries specified in the header
	written         int32 // amount of file data already written
}

type magicBytesWriter struct {
	w               io.Writer
	wroteMagicBytes bool
}

func (mw *magicBytesWriter) Write(b []byte) (int, error) {
	written := 0
	if !mw.wroteMagicBytes {
		n, err := mw.w.Write([]byte{C4GzMagic1, C4GzMagic2})
		if err != nil {
			return n, err
		}
		written += n
		mw.wroteMagicBytes = true
		b = b[2:]
	}
	n, err := mw.w.Write(b)
	written += n
	return written, err
}

// NewWriter creates a new Writer writing to w.
func NewWriter(w io.Writer) *Writer {
	mw := &magicBytesWriter{w: w}
	gz := gzip.NewWriter(mw)
	return &Writer{w: gz, gz: gz}
}

// CreateSubGroup starts a subfolder as part of the group's file data.
func (cw *Writer) CreateSubGroup(hdr *Header) (*Writer, error) {
	sub := &Writer{w: cw}
	err := sub.WriteHeader(hdr)
	if err != nil {
		return nil, err
	}
	return sub, nil
}

// WriteHeader writes a new group header to the group.
func (cw *Writer) WriteHeader(hdr *Header) error {
	if cw.haveHeader {
		return ErrHeaderAlreadyWritten
	}
	header := header{
		ver1:    C4GroupFileVer1,
		ver2:    C4GroupFileVer2,
		entries: hdr.Entries,
	}
	copy(header.id[:], C4GroupFileID)
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, header)
	memScramble(buf.Bytes())
	_, err := io.Copy(cw.w, &buf)
	//_, err := buf.WriteTo(cw.w)
	cw.offset = 0
	cw.haveHeader = true
	cw.expectedEntries = hdr.Entries
	return err
}

func b2i(b bool) int8 {
	if b {
		return 1
	} else {
		return 0
	}
}

// WriteHeader writes an entry header to the group.
func (cw *Writer) WriteEntry(e *Entry) error {
	if !cw.haveHeader {
		return ErrNoHeader
	}
	if cw.expectedEntries <= 0 {
		return ErrTooManyEntries
	}
	entry := entry{
		childGroup: int32(b2i(e.IsGroup)),
		size:       int32(e.Size),
		offset:     cw.offset,
		executable: byte(b2i(e.Executable)),
	}
	copy(entry.filename[:], e.Filename)
	cw.offset += int32(e.Size)
	cw.expectedEntries--
	err := binary.Write(cw.w, binary.LittleEndian, entry)
	return err
}

// Write writes file data to the group. The file has to match a previously
// written entry, but this is not checked.
func (cw *Writer) Write(b []byte) (int, error) {
	if !cw.haveHeader {
		return 0, ErrNoHeader
	}
	if cw.expectedEntries != 0 {
		return 0, ErrNotEnoughEntries
	}
	// TODO: More checking. Does this write correspond to a file entry? etc.
	n, err := cw.w.Write(b)
	cw.written += int32(n)
	if cw.written > cw.offset {
		return n, ErrTooMuchWritten
	}
	return n, err
}

// Close closes the Writer by flushing any unwritten data and writing the footer.
func (cw *Writer) Close() error {
	if cw.written < cw.offset {
		return ErrNotEnoughWritten
	}
	if cw.gz != nil {
		return cw.gz.Close()
	}
	return nil
}
