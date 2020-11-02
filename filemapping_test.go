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
