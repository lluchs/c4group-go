package main

import (
	"fmt"
	"io"
	"os"

	"github.com/lluchs/c4group-go"
	"gopkg.in/libgit2/git2go.v27"
)

func main() {
	packer, err := NewPacker("/home/lukas/src/openclonk")
	if err != nil {
		fmt.Println(err)
		return
	}

	f, err := os.Create("/tmp/Objects.ocd")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	err = packer.PackTo(f, "v8.1", "planet/Objects.ocd")
	if err != nil {
		fmt.Println(err)
		return
	}
}

type Packer struct {
	repo *git.Repository
}

func NewPacker(repoPath string) (*Packer, error) {
	repo, err := git.OpenRepository(repoPath)
	if err != nil {
		return nil, err
	}
	return &Packer{repo: repo}, nil
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
	err = p.writeEntries(treeToPack, cw)
	return err
}

func (p *Packer) writeEntries(tree *git.Tree, cw *c4group.Writer) error {
	count := tree.EntryCount()
	// First pass: write entry headers
	for i := uint64(0); i < count; i++ {
		entry := tree.EntryByIndex(i)
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
			blob, err := p.repo.LookupBlob(entry.Id)
			if err != nil {
				return err
			}
			c4entry.Size = int(blob.Size())
			c4entry.Executable = entry.Filemode == git.FilemodeBlobExecutable
		default:
			panic("invalid git entry type")
		}
		if err := cw.WriteEntry(&c4entry); err != nil {
			return err
		}
	}

	// Second pass: write entry contents, including subgroups
	for i := uint64(0); i < count; i++ {
		entry := tree.EntryByIndex(i)
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
			err = p.writeEntries(subtree, subgroup)
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
			// TODO: memoize sub-sizes
			s, err2 := p.calcTreeSize(entry)
			if err2 != nil {
				err = err2
				return
			}
			size += c4group.EntrySize + s
		case git.ObjectBlob:
			blob, err2 := p.repo.LookupBlob(entry.Id)
			if err2 != nil {
				err = err2
				return
			}
			size += c4group.EntrySize + int(blob.Size())
		default:
			panic("invalid git entry type")
		}
	}
	return
}