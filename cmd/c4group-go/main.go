package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"regexp"
	"strconv"
	"text/tabwriter"

	"github.com/lluchs/c4group-go"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: ", os.Args[0], " <group>")
		return
	}
	filename := os.Args[1]

	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	// Calculate CRC32 and SHA-1 hashes.
	crcW := crc32.NewIEEE()
	shaW := sha1.New()
	mw := io.MultiWriter(crcW, shaW)
	_, err = io.Copy(mw, file)
	if err != nil {
		fmt.Println("error calculating hashes: ", err)
		return
	}
	sha := shaW.Sum(nil)
	crc := crcW.Sum32()

	// Rewind to read contents.
	file.Seek(0, 0)
	reader, err := c4group.NewReader(file)
	if err != nil {
		fmt.Println(err)
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, "Author:\t%s\n", reader.Header.Author)
	fmt.Fprintf(w, "Time:\t%s\n", reader.Header.Ctime)
	fmt.Fprintf(w, "CRC32:\t%d\n", crc)
	fmt.Fprintf(w, "SHA-1:\t%s\n", hex.EncodeToString(sha))
	fmt.Fprintln(w)
	w.Flush()

	w.Init(os.Stdout, 5, 0, 3, ' ', tabwriter.AlignRight)
	for _, entry := range reader.Entries {
		info := ""
		if entry.IsGroup {
			info += " (Group)"
		}
		if entry.Executable {
			info += " (Executable)"
		}
		fmt.Fprintf(w, "%s\t%d Bytes\t%s\n", entry.Filename, entry.Size, info)
	}
	fmt.Fprintln(w)
	w.Flush()

	// For league info, look for Title.txt and Scenario.txt
	icon := -1
	var titleDE, titleUS string
	for {
		entry, err := reader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println(err)
			return
		}
		switch entry.Filename {
		case "Scenario.txt":
			scenario, err := readToString(reader)
			if err != nil {
				fmt.Println("error reading Scenario.txt: ", err)
				return
			}
			m := regexp.MustCompile(`(?m)^Icon=(\d+)`).FindStringSubmatch(scenario)
			if m != nil {
				icon, _ = strconv.Atoi(m[1])
			}
		case "Title.txt":
			title, err := readToString(reader)
			if err != nil {
				fmt.Println("error reading Title.txt: ", err)
				return
			}
			m := regexp.MustCompile(`(?m)^DE:(.*)$`).FindStringSubmatch(title)
			if m != nil {
				titleDE = m[1]
			}
			m = regexp.MustCompile(`(?m)^US:(.*)$`).FindStringSubmatch(title)
			if m != nil {
				titleUS = m[1]
			}
			// Scenario.txt will always come before Title.txt, we can skip reading the rest.
			break
		}
	}
	fmt.Printf("Title (DE): %s\n", titleDE)
	fmt.Printf("Title (US): %s\n", titleUS)
	fmt.Printf("Icon: %d\n", icon)
}

// readToString reads from r until EOF and converts to a string.
func readToString(r io.Reader) (string, error) {
	var buf bytes.Buffer
	_, err := io.Copy(&buf, r)
	if err != nil {
		return "", err
	}
	return string(buf.Bytes()), nil
}
