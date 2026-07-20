package deduplicate

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
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
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	var (
		header          []string
		records         []Record
		pendingComments []string
		atTop           = true
	)

	scanner := bufio.NewScanner(f)
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

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	w := bufio.NewWriter(out)

	// Original header.
	for _, line := range header {
		fmt.Fprintln(w, line)
	}

	// Duplicate comment section.
	if len(duplicateComments) > 0 {
		if len(header) > 0 {
			fmt.Fprintln(w)
		}

		fmt.Fprintln(w, "# Comments from removed duplicate entries")
		for _, line := range duplicateComments {
			fmt.Fprintln(w, line)
		}
		fmt.Fprintln(w)
	}

	// Sorted unique entries.
	for i, r := range unique {
		if i > 0 && len(r.Comments) > 0 {
			fmt.Fprintln(w)
		}

		for _, c := range r.Comments {
			fmt.Fprintln(w, c)
		}
		fmt.Fprintln(w, r.Line)
	}

	return w.Flush()
}
