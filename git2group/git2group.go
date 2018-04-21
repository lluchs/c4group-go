// Copyright © 2018, Lukas Werling
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

package git2group

import (
	"io"
	"sort"

	"github.com/lluchs/c4group-go"
	"gopkg.in/libgit2/git2go.v26"
)

type Packer struct {
	repo     *git.Repository
	odb      *git.Odb
	treeSize map[git.Oid]int
}

func NewPacker(repoPath string) (*Packer, error) {
	repo, err := git.OpenRepository(repoPath)
	if err != nil {
		return nil, err
	}
	odb, err := repo.Odb()
	if err != nil {
		return nil, err
	}

	return &Packer{
		repo:     repo,
		odb:      odb,
		treeSize: make(map[git.Oid]int),
	}, nil
}

func (p *Packer) PackTo(w io.Writer, rev, path string) error {
	obj, err := p.repo.RevparseSingle(rev + "^{tree}")
	if err != nil {
		return err
	}
	tree, err := obj.AsTree()
	if err != nil {
		return err
	}

	entry, err := tree.EntryByPath(path)
	if err != nil {
		return err
	}
	treeToPack, err := p.repo.LookupTree(entry.Id)
	if err != nil {
		return err
	}

	cw := c4group.NewWriter(w)
	err = cw.WriteHeader(&c4group.Header{
		Entries: int32(treeToPack.EntryCount()),
	})
	if err != nil {
		return err
	}
	err = p.writeEntries(entry.Name, treeToPack, cw)
	return err
}

type entrySlice struct {
	s    []*git.TreeEntry
	less func(a, b string) bool
}

func newEntrySlice(name string, tree *git.Tree) *entrySlice {
	count := tree.EntryCount()
	slice := make([]*git.TreeEntry, count)
	for i := uint64(0); i < count; i++ {
		slice[i] = tree.EntryByIndex(i)
	}
	less := c4group.NameLess(name)
	return &entrySlice{s: slice, less: less}
}

func (e *entrySlice) Len() int {
	return len(e.s)
}

func (e *entrySlice) Less(i int, j int) bool {
	return e.less(e.s[i].Name, e.s[j].Name)
}

func (e *entrySlice) Swap(i int, j int) {
	e.s[i], e.s[j] = e.s[j], e.s[i]
}

func (p *Packer) writeEntries(name string, tree *git.Tree, cw *c4group.Writer) error {
	entries := newEntrySlice(name, tree)
	sort.Sort(entries)
	// First pass: write entry headers
	for _, entry := range entries.s {
		c4entry := c4group.Entry{
			Filename: entry.Name,
		}
		switch entry.Type {
		case git.ObjectTree:
			c4entry.IsGroup = true
			size, err := p.calcTreeSize(entry)
			if err != nil {
				return err
			}
			c4entry.Size = size
		case git.ObjectBlob:
			// Avoid reading the whole blob into memory here.
			s, _, err := p.odb.ReadHeader(entry.Id)
			if err != nil {
				return err
			}
			c4entry.Size = int(s)
			c4entry.Executable = entry.Filemode == git.FilemodeBlobExecutable
		default:
			panic("invalid git entry type")
		}
		if err := cw.WriteEntry(&c4entry); err != nil {
			return err
		}
	}

	// Second pass: write entry contents, including subgroups
	for _, entry := range entries.s {
		switch entry.Type {
		case git.ObjectTree:
			subtree, err := p.repo.LookupTree(entry.Id)
			if err != nil {
				return err
			}
			subgroup, err := cw.CreateSubGroup(&c4group.Header{
				Entries: int32(subtree.EntryCount()),
			})
			if err != nil {
				return err
			}
			err = p.writeEntries(entry.Name, subtree, subgroup)
			if err != nil {
				return err
			}
		case git.ObjectBlob:
			blob, err := p.repo.LookupBlob(entry.Id)
			if err != nil {
				return err
			}
			_, err = cw.Write(blob.Contents())
			if err != nil {
				return err
			}
		default:
			panic("invalid git entry type")
		}
	}
	return cw.Close()
}

func (p *Packer) calcTreeSize(entry *git.TreeEntry) (size int, err error) {
	tree, err := p.repo.LookupTree(entry.Id)
	if err != nil {
		return
	}
	size = c4group.HeaderSize

	count := tree.EntryCount()
	for i := uint64(0); i < count; i++ {
		entry := tree.EntryByIndex(i)
		switch entry.Type {
		case git.ObjectTree:
			var s int
			// memoize sub-sizes
			if s2, ok := p.treeSize[*entry.Id]; ok {
				s = s2
			} else {
				s, err = p.calcTreeSize(entry)
				if err != nil {
					return
				}
				p.treeSize[*entry.Id] = s
			}
			size += c4group.EntrySize + s
		case git.ObjectBlob:
			// Avoid reading the whole blob into memory here.
			s, _, err2 := p.odb.ReadHeader(entry.Id)
			if err2 != nil {
				err = err2
				return
			}
			size += c4group.EntrySize + int(s)
		default:
			panic("invalid git entry type")
		}
	}
	return
}
