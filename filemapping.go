package gettext

import (
	"io/ioutil"
	"os"
	"runtime"
)

type fileMapping struct {
	data []byte

	isMapped bool
}

func (m *fileMapping) Close() error {
	runtime.SetFinalizer(m, nil)
	if !m.isMapped {
		return nil
	}
	return m.closeMapping()
}

func openMapping(f *os.File) (*fileMapping, error) {
	m := new(fileMapping)

	err := m.tryMap(f)
	if err == nil {
		runtime.SetFinalizer(m, (*fileMapping).Close)
		return m, nil
	}
	// On mapping failure, fall back to reading the file into
	// memory directly.
	if _, err = f.Seek(0, os.SEEK_SET); err != nil {
		return nil, err
	}
	m.data, err = ioutil.ReadAll(f)
	return m, err
}
