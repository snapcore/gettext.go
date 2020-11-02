// +build !windows

package gettext

import (
	"fmt"
	"os"
	"syscall"
)

func (m *fileMapping) tryMap(f *os.File) error {
	fi, err := f.Stat()
	if err != nil {
		return err
	}

	size := fi.Size()
	if size == 0 {
		return nil
	}
	if size < 0 {
		return fmt.Errorf("file %q has negative size", fi.Name())
	}
	if size != int64(int(size)) {
		return fmt.Errorf("file %q is too large", fi.Name())
	}
	m.data, err = syscall.Mmap(int(f.Fd()), 0, int(size), syscall.PROT_READ, syscall.MAP_PRIVATE)
	if err != nil {
		return err
	}
	m.isMapped = true
	return nil
}

func (m *fileMapping) closeMapping() error {
	return syscall.Munmap(m.data)
}
