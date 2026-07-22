package manage

import (
	"bytes"
	"fmt"
	"strings"

	"better-blocklist/src/internal/compression"
)

// addToFile appends a value to a text file. If the file is already
// gzip-compressed, it is decompressed, modified, and recompressed.
// If it is not compressed, it is read as plain text and then saved
// back as a gzip-compressed file.
func addToFile(path string, value string) error {
	v := strings.TrimSpace(value)
	if v == "" {
		return fmt.Errorf("empty value")
	}

	// Read the file (decompress if needed).
	data, err := compression.ReadMaybeGzip(path)
	if err != nil {
		return err
	}

	// Append the new value.
	var buf bytes.Buffer
	buf.Write(data)

	if len(data) > 0 && data[len(data)-1] != '\n' {
		buf.WriteByte('\n')
	}
	buf.WriteString(v)
	buf.WriteByte('\n')

	// Write back compressed.
	return compression.WriteGzipFile(path, buf.Bytes())
}
