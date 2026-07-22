package manage

import (
	"fmt"
	"strings"

	"better-blocklist/src/internal/compression"
)

// removeFromFile removes all occurrences of a value from a text file.
// If the file is gzip-compressed, it is decompressed first.
// The file is always written back as gzip-compressed.
func removeFromFile(path string, value string) error {
	v := strings.TrimSpace(value)
	if v == "" {
		return fmt.Errorf("empty value")
	}
	// Read the file (decompress if needed).
	data, err := compression.ReadMaybeGzip(path)
	if err != nil {
		return err
	}

	// Remove matching lines.
	lines := strings.Split(string(data), "\n")
	var kept []string

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.TrimSpace(line) != v {
			kept = append(kept, line)
		}
	}

	output := strings.Join(kept, "\n")
	if output != "" {
		output += "\n"
	}

	// Rewrite as gzip-compressed.
	return compression.WriteGzipFile(path, []byte(output))
}
