package gettext

// Catalog of translations for a given locale.
type Catalog struct {
	mos []*mocatalog
}

func (c Catalog) findMsg(msgid string, usePlural bool, n uint32) (msgstr string, ok bool) {
	for _, mo := range c.mos {
		if msgstr, ok := mo.findMsg(msgid, usePlural, n); ok {
			return msgstr, true
		}
	}
	return "", false
}

func (c Catalog) Gettext(msgid string) string {
	if msgstr, ok := c.findMsg(msgid, false, 0); ok {
		return msgstr
	}
	// Fallback to original message
	return msgid
}

func (c Catalog) NGettext(msgid, msgid_plural string, n uint32) string {
	if msgstr, ok := c.findMsg(msgid, true, n); ok {
		return msgstr
	}
	// Fallback to original message based on Germanic plural rule.
	if n == 1 {
		return msgid
	}
	return msgid_plural
}

func (c Catalog) PGettext(msgctxt, msgid string) string {
	if msgstr, ok := c.findMsg(msgctxt+"\x04"+msgid, false, 0); ok {
		return msgstr
	}
	return msgid
}

func (c Catalog) PNGettext(msgctxt, msgid, msgid_plural string, n uint32) string {
	if msgstr, ok := c.findMsg(msgctxt+"\x04"+msgid, true, n); ok {
		return msgstr
	}
	// Fallback to original message based on Germanic plural rule.
	if n == 1 {
		return msgid
	}
	return msgid_plural
}
