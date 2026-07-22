package deduplicate

import (
	"bufio"
	"bytes"
	"fmt"
	"sort"
	"strings"

	"better-blocklist/src/internal/compression"
)

type Record struct {
	Comments []string
	Line     string
}

// SortFile sorts a file alphabetically while:
//
// - Preserving any comment block at the start of the file.
// - Keeping comments attached to the line below them.
// - Removing duplicate entries.
// - Moving comments from removed duplicates into a section at the top.
func SortFile(path string) error {

	// Read file (decompress if gzip).
	raw, err := compression.ReadMaybeGzip(path)
	if err != nil {
		return err
	}

	var (
		header          []string
		records         []Record
		pendingComments []string
		atTop           = true
	)

	scanner := bufio.NewScanner(bytes.NewReader(raw))
	for scanner.Scan() {
		line := scanner.Text()

		// Comment line
		if strings.HasPrefix(line, "#") {
			if atTop {
				header = append(header, line)
			} else {
				pendingComments = append(pendingComments, line)
			}
			continue
		}

		// Blank line
		if strings.TrimSpace(line) == "" {
			if atTop {
				header = append(header, "")
			}
			continue
		}

		atTop = false

		records = append(records, Record{
			Comments: pendingComments,
			Line:     line,
		})
		pendingComments = nil
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// Remove duplicates while remembering comments.
	seen := make(map[string]bool)
	var unique []Record
	var duplicateComments []string

	for _, r := range records {
		if seen[r.Line] {
			if len(r.Comments) > 0 {
				duplicateComments = append(
					duplicateComments,
					fmt.Sprintf("# Duplicate removed for: %s", r.Line),
				)
				duplicateComments = append(duplicateComments, r.Comments...)
				duplicateComments = append(duplicateComments, "")
			}
			continue
		}

		seen[r.Line] = true
		unique = append(unique, r)
	}

	// Sort alphabetically (case-insensitive).
	sort.Slice(unique, func(i, j int) bool {
		return strings.ToLower(unique[i].Line) < strings.ToLower(unique[j].Line)
	})

	var buf bytes.Buffer
	// Original header.
	for _, line := range header {
		fmt.Fprintln(&buf, line)
	}

	// Duplicate comment section.
	if len(duplicateComments) > 0 {
		if len(header) > 0 {
			fmt.Fprintln(&buf)
		}

		fmt.Fprintln(&buf, "# Comments from removed duplicate entries")
		for _, line := range duplicateComments {
			fmt.Fprintln(&buf, line)
		}
		fmt.Fprintln(&buf)
	}

	// Sorted unique entries.
	for i, r := range unique {
		if i > 0 && len(r.Comments) > 0 {
			fmt.Fprintln(&buf)
		}

		for _, c := range r.Comments {
			fmt.Fprintln(&buf, c)
		}
		fmt.Fprintln(&buf, r.Line)
	}

	// Write compressed output.
	return compression.WriteGzipFile(path, buf.Bytes())
}
