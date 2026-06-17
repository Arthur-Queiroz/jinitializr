package zipper

import (
	"archive/zip"
	"bytes"
	"testing"
)

func TestZipRoundTrip(t *testing.T) {
	in := map[string][]byte{
		"go.mod":           []byte("module example.com/x\n"),
		"cmd/api/main.go":  []byte("package main\n"),
		"internal/keep.go": []byte("package internal\n"),
	}

	out, err := New().Zip(in)
	if err != nil {
		t.Fatalf("Zip: %v", err)
	}

	r, err := zip.NewReader(bytes.NewReader(out), int64(len(out)))
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	if len(r.File) != len(in) {
		t.Fatalf("got %d entries, want %d", len(r.File), len(in))
	}

	for _, f := range r.File {
		want, ok := in[f.Name]
		if !ok {
			t.Errorf("unexpected entry %q", f.Name)
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("open %q: %v", f.Name, err)
		}
		var got bytes.Buffer
		if _, err := got.ReadFrom(rc); err != nil {
			t.Fatalf("read %q: %v", f.Name, err)
		}
		rc.Close()
		if !bytes.Equal(got.Bytes(), want) {
			t.Errorf("entry %q = %q, want %q", f.Name, got.Bytes(), want)
		}
	}
}

func TestZipEmpty(t *testing.T) {
	out, err := New().Zip(map[string][]byte{})
	if err != nil {
		t.Fatalf("Zip(empty): %v", err)
	}
	r, err := zip.NewReader(bytes.NewReader(out), int64(len(out)))
	if err != nil {
		t.Fatalf("empty archive is not a valid zip: %v", err)
	}
	if len(r.File) != 0 {
		t.Errorf("got %d entries, want 0", len(r.File))
	}
}

func TestZipDeterministic(t *testing.T) {
	in := map[string][]byte{"b.go": []byte("b"), "a.go": []byte("a"), "c.go": []byte("c")}
	a, err := New().Zip(in)
	if err != nil {
		t.Fatal(err)
	}
	b, err := New().Zip(in)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a, b) {
		t.Error("Zip output is not deterministic for identical input")
	}
}
