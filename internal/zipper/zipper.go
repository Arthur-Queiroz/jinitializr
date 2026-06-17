// Package zipper builds a zip archive in memory from a map of path to content.
package zipper

import (
	"archive/zip"
	"bytes"
	"fmt"
	"sort"
)

// Zipper builds zip archives in memory.
type Zipper struct{}

// New constructs a Zipper.
func New() *Zipper {
	return &Zipper{}
}

// Zip packs the given files — keyed by their path inside the archive — into a
// zip archive built entirely in memory. Entries are written in sorted path
// order so the same input always yields byte-identical output.
func (z *Zipper) Zip(files map[string][]byte) ([]byte, error) {
	paths := make([]string, 0, len(files))
	for p := range files {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for _, p := range paths {
		f, err := w.Create(p)
		if err != nil {
			return nil, fmt.Errorf("create zip entry %q: %w", p, err)
		}
		if _, err := f.Write(files[p]); err != nil {
			return nil, fmt.Errorf("write zip entry %q: %w", p, err)
		}
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("close zip writer: %w", err)
	}
	return buf.Bytes(), nil
}
