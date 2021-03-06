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

package main

import (
	"fmt"
	"os"

	"github.com/lluchs/c4group-go/git2group"
)

func main() {
	if len(os.Args) != 5 {
		fmt.Println("Usage:", os.Args[0], "<repository> <path> <revision> <output>")
		return
	}
	repoPath := os.Args[1]
	objPath := os.Args[2]
	revision := os.Args[3]
	outputPath := os.Args[4]

	packer, err := git2group.NewPacker(repoPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	f, err := os.Create(outputPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	err = packer.PackTo(f, revision, objPath)
	if err != nil {
		fmt.Println(err)
		return
	}
}
