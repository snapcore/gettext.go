package gettext

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/snapcore/go-gettext/pluralforms"
)

const le_magic = 0x950412de
const be_magic = 0xde120495

type header struct {
	Magic          uint32
	Version        uint32
	NumStrings     uint32
	OrigTabOffset  uint32
	TransTabOffset uint32
	HashTabSize    uint32
	HashTabOffset  uint32
}

func (header header) get_major_version() uint32 {
	return header.Version >> 16
}

func (header header) get_minor_version() uint32 {
	return header.Version & 0xffff
}

// Catalog of translations for a given locale.
type Catalog interface {
	Gettext(msgid string) string
	NGettext(msgid string, msgid_plural string, n uint32) string
}

type mocatalog struct {
	m     *fileMapping
	order binary.ByteOrder

	numStrings int
	origTab    []byte
	transTab   []byte
	hashTab    []byte

	info        map[string]string
	language    string
	pluralforms pluralforms.Expression
	charset     string
}

type nullcatalog struct{}

func (catalog nullcatalog) Gettext(msgid string) string {
	return msgid
}

func (catalog nullcatalog) NGettext(msgid string, msgid_plural string, n uint32) string {
	if n == 1 {
		return msgid
	} else {
		return msgid_plural
	}
}

func (catalog *mocatalog) Gettext(msgid string) string {
	idx, ok := catalog.msgIndex(msgid)
	if !ok {
		return msgid
	}
	return string(catalog.msgStr(idx, 0))
}

func (catalog *mocatalog) NGettext(msgid string, msgid_plural string, n uint32) string {
	idx, ok := catalog.msgIndex(msgid)
	if !ok {
		if n == 1 {
			return msgid
		} else {
			return msgid_plural
		}
	}

	var plural int
	if catalog.pluralforms != nil {
		plural = catalog.pluralforms.Eval(n)
	} else {
		// Bogus/missing pluralforms in mo: Use the Germanic
		// plural rule.
		if n == 1 {
			plural = 0
		} else {
			plural = 1
		}
	}

	return string(catalog.msgStr(idx, plural))
}

func (catalog *mocatalog) msgID(idx int) []byte {
	strLen := catalog.order.Uint32(catalog.origTab[8*idx:])
	strOffset := catalog.order.Uint32(catalog.origTab[8*idx+4:])
	msgid := catalog.m.data[strOffset : strOffset+strLen]

	zero := bytes.IndexByte(msgid, '\x00')
	if zero >= 0 {
		msgid = msgid[:zero]
	}
	return msgid
}

func (catalog *mocatalog) msgStr(idx, n int) []byte {
	strLen := catalog.order.Uint32(catalog.transTab[8*idx:])
	strOffset := catalog.order.Uint32(catalog.transTab[8*idx+4:])
	msgstr := catalog.m.data[strOffset : strOffset+strLen]

	for ; n >= 0; n-- {
		zero := bytes.IndexByte(msgstr, '\x00')
		if n == 0 {
			if zero >= 0 {
				msgstr = msgstr[:zero]
			}
			break
		} else {
			// fast forward to next string.  If there is
			// no nul byte, then this is a no-op
			msgstr = msgstr[zero+1:]
		}
	}
	return msgstr
}

func (catalog *mocatalog) msgIndex(msgid string) (idx int, ok bool) {
	// perform a binary search over origTab message IDs
	idx = sort.Search(catalog.numStrings, func(i int) bool {
		return string(catalog.msgID(i)) >= msgid
	})
	if idx < catalog.numStrings && string(catalog.msgID(idx)) == msgid {
		return idx, true
	}
	return 0, false
}

func (catalog *mocatalog) read_info(info string) error {
	catalog.info = make(map[string]string)
	lastk := ""
	for _, line := range strings.Split(info, "\n") {
		item := strings.TrimSpace(line)
		if len(item) == 0 {
			continue
		}
		var k string
		var v string
		if strings.Contains(item, ":") {
			tmp := strings.SplitN(item, ":", 2)
			k = strings.ToLower(strings.TrimSpace(tmp[0]))
			v = strings.TrimSpace(tmp[1])
			catalog.info[k] = v
			lastk = k
		} else if len(lastk) != 0 {
			catalog.info[lastk] += "\n" + item
		}
		if k == "content-type" {
			catalog.charset = strings.Split(v, "charset=")[1]
		} else if k == "plural-forms" {
			p := strings.Split(v, ";")[1]
			s := strings.Split(p, "plural=")[1]
			expr, err := pluralforms.Compile(s)
			if err != nil {
				return err
			}
			catalog.pluralforms = expr
		}
	}
	return nil
}

func validateStringTable(m *fileMapping, table []byte, numStrings int, order binary.ByteOrder) error {
	for i := 0; i < numStrings; i++ {
		strLen := order.Uint32(table[8*i:])
		strOffset := order.Uint32(table[8*i+4:])
		if int(strLen+strOffset) > len(m.data) {
			return fmt.Errorf("string %d data (len=%x, offset=%x) is out of bounds", i, strLen, strOffset)
		}
	}
	return nil
}

func validateHashTable(table []byte, numStrings int, order binary.ByteOrder) error {
	for i := 0; i < numStrings; i++ {
		strIndex := order.Uint32(table[4*i:])
		// hash entries are either zero or a string index
		// incremented by one
		if int(strIndex) >= numStrings+1 {
			return fmt.Errorf("hash table is corrupt")
		}
	}
	return nil
}

// ParseMO parses a mo file into a Catalog if possible.
func ParseMO(file *os.File) (Catalog, error) {
	m, err := openMapping(file)
	if err != nil {
		return nil, err
	}
	defer func() {
		if m != nil {
			m.Close()
		}
	}()

	var header header
	headerSize := binary.Size(&header)
	if len(m.data) < headerSize {
		return nil, fmt.Errorf("message catalogue is too short")
	}

	var order binary.ByteOrder = binary.LittleEndian
	magic := order.Uint32(m.data)
	switch magic {
	case le_magic:
		// nothing
	case be_magic:
		order = binary.BigEndian
	default:
		return nil, fmt.Errorf("Wrong magic: %d", magic)
	}
	if err := binary.Read(bytes.NewBuffer(m.data[:headerSize]), order, &header); err != nil {
		return nil, err
	}
	if header.get_major_version() != 0 && header.get_major_version() != 1 {
		return nil, fmt.Errorf("Unsupported version: %d.%d", header.get_major_version(), header.get_minor_version())
	}
	if int64(int(header.NumStrings)) != int64(header.NumStrings) {
		return nil, fmt.Errorf("too many strings in catalog")
	}
	numStrings := int(header.NumStrings)

	if int(header.OrigTabOffset+8*header.NumStrings) > len(m.data) {
		return nil, fmt.Errorf("original strings table out of bounds")
	}
	origTab := m.data[header.OrigTabOffset : header.OrigTabOffset+8*header.NumStrings]
	if err := validateStringTable(m, origTab, numStrings, order); err != nil {
		return nil, err
	}

	if int(header.TransTabOffset+8*header.NumStrings) > len(m.data) {
		return nil, fmt.Errorf("translated strings table out of bounds")
	}
	transTab := m.data[header.TransTabOffset : header.TransTabOffset+8*header.NumStrings]
	if err := validateStringTable(m, transTab, numStrings, order); err != nil {
		return nil, err
	}

	var hashTab []byte
	if header.HashTabSize > 2 {
		if int(header.HashTabOffset+4*header.HashTabSize) > len(m.data) {
			return nil, fmt.Errorf("hash table out of bounds")
		}
		hashTab = m.data[header.HashTabOffset : header.HashTabOffset+4*header.HashTabSize]
		if err := validateHashTable(hashTab, numStrings, order); err != nil {
			return nil, err
		}
	}

	catalog := &mocatalog{
		m:     m,
		order: order,

		numStrings: numStrings,
		origTab:    origTab,
		transTab:   transTab,
		hashTab:    hashTab,
	}
	// Read catalog header if available
	if catalog.numStrings > 0 && len(catalog.msgID(0)) == 0 {
		if err := catalog.read_info(string(catalog.msgStr(0, 0))); err != nil {
			return nil, err
		}
	}

	m = nil
	return catalog, nil
}
