package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path"
	"regexp"
	"strconv"
	"text/tabwriter"

	"github.com/lluchs/c4group-go"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: ", os.Args[0], " <action> <group>")
		return
	}
	action := os.Args[1]
	filename := os.Args[2]

	// TODO: Distinguish read and write actions.

	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	reader, err := c4group.NewReader(file)
	if err != nil {
		fmt.Println(err)
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	if reader.Header.Author != "" {
		fmt.Fprintf(w, "Author:\t%s\n", reader.Header.Author)
	}
	if reader.Header.Ctime.Unix() != 0 {
		fmt.Fprintf(w, "Time:\t%s\n", reader.Header.Ctime)
	}

	switch action {
	case "list":
		fmt.Fprintln(w)
		w.Flush()
		PrintGroupContents(reader)

	case "league-info":
		crc, sha, err := CalculateHashes(filename)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Fprintf(w, "CRC32:\t%d\n", crc)
		fmt.Fprintf(w, "SHA-1:\t%s\n", sha)

		// For league info, look for Title.txt and Scenario.txt
		icon := -1
		maxPlayer := -1
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
				if m := regexp.MustCompile(`(?m)^Icon=(\d+)`).FindStringSubmatch(scenario); m != nil {
					icon, _ = strconv.Atoi(m[1])
				}
				if m := regexp.MustCompile(`(?m)^MaxPlayer=(\d+)`).FindStringSubmatch(scenario); m != nil {
					maxPlayer, _ = strconv.Atoi(m[1])
				}
			case "Title.txt":
				title, err := readToString(reader)
				if err != nil {
					fmt.Println("error reading Title.txt: ", err)
					return
				}
				// Skip whitespace at the end to avoid capturing \r
				m := regexp.MustCompile(`(?m)^DE:(.*?)\s*$`).FindStringSubmatch(title)
				if m != nil {
					titleDE = m[1]
				}
				m = regexp.MustCompile(`(?m)^US:(.*?)\s*$`).FindStringSubmatch(title)
				if m != nil {
					titleUS = m[1]
				}
				// Scenario.txt will always come before Title.txt, we can skip reading the rest.
				break
			}
		}
		fmt.Fprintf(w, "Title (DE): %s\n", titleDE)
		fmt.Fprintf(w, "Title (US): %s\n", titleUS)
		if icon != -1 {
			fmt.Fprintf(w, "Icon: %d\n", icon)
		}
		fmt.Fprintln(w)
		w.Flush()

		printScenarioAddJS("scenario[name][1]", titleUS)
		printScenarioAddJS("scenario[name][2]", titleDE)
		if icon != -1 {
			fmt.Printf("document.all['scenario[icon_number]'][%d].checked=true;", icon)
		}
		printScenarioAddJS("versions[0][hash]", strconv.FormatUint(uint64(crc), 10))
		printScenarioAddJS("versions[0][hash_sha]", sha)
		printScenarioAddJS("versions[0][filename]", path.Base(filename))
		printScenarioAddJS("versions[0][author]", reader.Header.Author)
		printScenarioAddJS("versions[0][comment]", "manual import")
		if maxPlayer != -1 {
			fmt.Printf(`document.querySelectorAll('[name$="[max_player_count]"]').forEach(e=>e.value=%d);`, maxPlayer)
		}
		fmt.Println()
	}

}

// CalculateHashes calculates CRC32 and SHA-1 of the given file.
func CalculateHashes(filename string) (crc uint32, sha string, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	// Calculate CRC32 and SHA-1 hashes.
	crcW := crc32.NewIEEE()
	shaW := sha1.New()
	mw := io.MultiWriter(crcW, shaW)
	_, err = io.Copy(mw, file)
	if err != nil {
		err = fmt.Errorf("error calculating hashes: %s", err)
		return
	}
	sha = hex.EncodeToString(shaW.Sum(nil))
	crc = crcW.Sum32()
	return
}

// PrintGroupContents prints filename, size and attributes similar to c4group -l.
func PrintGroupContents(reader *c4group.Reader) {
	w := tabwriter.NewWriter(os.Stdout, 5, 0, 3, ' ', tabwriter.AlignRight)
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

func printScenarioAddJS(key, value string) {
	fmt.Printf("document.all['%s'].value=`%s`;", key, value)
}
