package compression

import (
	"bytes"
	"compress/gzip"
	"io"
	"os"
)

// ReadMaybeGzip reads a file and returns its uncompressed contents.
// If the file is gzip-compressed it will be decompressed, otherwise
// the raw bytes are returned.
func ReadMaybeGzip(path string) ([]byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// detect gzip by magic number
	if len(raw) >= 2 && raw[0] == 0x1f && raw[1] == 0x8b {
		gzr, err := gzip.NewReader(bytes.NewReader(raw))
		if err != nil {
			return nil, err
		}
		defer gzr.Close()
		return io.ReadAll(gzr)
	}

	return raw, nil
}

// WriteGzipFile writes the given data to path using gzip compression
// with BestCompression level. The file is truncated or created.
func WriteGzipFile(path string, data []byte) error {
	// Check the path ends in .gz, if not, append it
	if len(path) < 3 || path[len(path)-3:] != ".gz" {
		os.Remove(path)
		path += ".gz"
	}

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	gzw, err := gzip.NewWriterLevel(out, gzip.BestCompression)
	if err != nil {
		return err
	}
	defer gzw.Close()

	_, err = gzw.Write(data)
	return err
}
