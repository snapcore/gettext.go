package gettext

import (
	"bytes"
	"os"
	"testing"
)

func TestFileMapping(t *testing.T) {
	file, err := os.Open("testdata/en/messages.mo")
	if err != nil {
		t.Fatal(err)
	}
	fi, err := file.Stat()
	if err != nil {
		t.Fatal(err)
	}

	m, err := openMapping(file)
	if err != nil {
		t.Fatal(err)
	}
	if !m.isMapped {
		t.Fatal("file content was not mapped")
	}

	if int64(len(m.data)) != fi.Size() {
		t.Logf("mapping size mismatch: %d != %d", len(m.data), fi.Size())
		t.Fail()
	}
	// Expect message catalogue magic number
	if !bytes.Equal(m.data[:4], []byte{0xde, 0x12, 0x04, 0x95}) {
		t.Logf("unexpected data in mapping: %q", m.data[:4])
		t.Fail()
	}

	err = m.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestFileMappingFallback(t *testing.T) {
	// We can't memory map a pipe, so this should result in
	// falling back to simply reading the data in to memory
	r, w, err := os.Pipe()
	go func() {
		if _, err := w.Write([]byte("Hello world!")); err != nil {
			t.Fatal(err)
		}
		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	m, err := openMapping(r)
	if err != nil {
		t.Fatal(err)
	}
	if m.isMapped {
		t.Fatal("expected file content not to be mapped")
	}

	// Expect content read from pipe
	if !bytes.Equal(m.data, []byte("Hello world!")) {
		t.Logf("unexpected data: %q", m.data)
		t.Fail()
	}

	err = m.Close()
	if err != nil {
		t.Fatal(err)
	}
}
