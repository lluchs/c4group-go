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
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/lluchs/c4group-go/git2group"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage:", os.Args[0], "<repository>")
		os.Exit(1)
	}
	repoPath := os.Args[1]

	packer, err := git2group.NewPacker(repoPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	packPathRegexp := regexp.MustCompile(`^/pack/([\w\.]+)/(.*\.oc[dgfs])$`)
	// /pack/<revision>/<path>
	http.HandleFunc("/pack/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "", http.StatusMethodNotAllowed)
			return
		}
		m := packPathRegexp.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.Error(w, "invalid URL "+r.URL.Path, http.StatusBadRequest)
			return
		}
		err = packer.PackTo(w, m[1], m[2])
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	listPathRegexp := regexp.MustCompile(`^/list/([\w\.]+)/(.*)$`)
	// /list/<revision>/<path>
	http.HandleFunc("/list/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "", http.StatusMethodNotAllowed)
			return
		}
		m := listPathRegexp.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.Error(w, "invalid URL "+r.URL.Path, http.StatusBadRequest)
			return
		}
		list, err := packer.ListGroups(m[1], m[2])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.Encode(list)
	})

	log.Fatal(http.ListenAndServe(os.Getenv("PORT"), nil))
}
