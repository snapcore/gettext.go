package gettext

import (
	"bytes"
	"testing"
)

func TestParseLocaleAlias(t *testing.T) {
	buf := bytes.NewBufferString(`
# Comment
      # also a comment
one-word-ignored
spanish         es_ES.ISO-8859-1
swedish         sv_SE.ISO-8859-1
`)
	aliases, err := parseLocaleAlias(buf)
	if err != nil {
		t.Fatal(err)
	}
	assertDeepEqual(t, aliases, map[string]string{
		"spanish": "es_ES.ISO-8859-1",
		"swedish": "sv_SE.ISO-8859-1",
	})
}

func TestNormalizeCodeset(t *testing.T) {
	assert_equal(t, normalizeCodeset(".UTF-8"), ".utf8")
	assert_equal(t, normalizeCodeset(".utf8"), ".utf8")

	assert_equal(t, normalizeCodeset(".ISO-8859-1"), ".iso88591")
	assert_equal(t, normalizeCodeset(".iso-8859-1"), ".iso88591")
	assert_equal(t, normalizeCodeset(".iso88591"), ".iso88591")
	assert_equal(t, normalizeCodeset(".8859-1"), ".iso88591")
	assert_equal(t, normalizeCodeset(".88591"), ".iso88591")
}

func TestExpandLocale(t *testing.T) {
	assertDeepEqual(t, expandLocale("en"), []string{"en"})
	assertDeepEqual(t, expandLocale("en_AU"), []string{"en_AU", "en"})
	assertDeepEqual(t, expandLocale("en_AU.UTF-8"), []string{"en_AU.UTF-8", "en_AU.utf8", "en_AU", "en.UTF-8", "en.utf8", "en"})
	assertDeepEqual(t, expandLocale("en_AU.utf8"), []string{"en_AU.utf8", "en_AU", "en.utf8", "en"})
	assertDeepEqual(t, expandLocale("en_AU.UTF-8@mod"), []string{"en_AU.UTF-8@mod", "en_AU.utf8@mod", "en_AU@mod", "en.UTF-8@mod", "en.utf8@mod", "en@mod", "en_AU.UTF-8", "en_AU.utf8", "en_AU", "en.UTF-8", "en.utf8", "en"})
}

func mockGetenv(env map[string]string) (restore func()) {
	old := osGetenv
	osGetenv = func(name string) string {
		return env[name]
	}
	return func() {
		osGetenv = old
	}
}

func TestUserLanguages(t *testing.T) {
	env := map[string]string{}
	restore := mockGetenv(env)
	defer restore()

	// By default, no locale is set
	assertDeepEqual(t, UserLanguages(), []string(nil))

	// If LANG is set, use that
	env["LANG"] = "en_AU@lang"
	assertDeepEqual(t, UserLanguages(), []string{"en_AU@lang"})

	// LC_MESSAGES overrides LANG
	env["LC_MESSAGES"] = "en_AU@messages"
	assertDeepEqual(t, UserLanguages(), []string{"en_AU@messages"})

	// LC_ALL overrides LC_MESSAGES
	env["LC_ALL"] = "en_AU.UTF-8"
	assertDeepEqual(t, UserLanguages(), []string{"en_AU.UTF-8"})

	// LANGUAGE overrides LC_ALL, and can specify multiple locales
	env["LANGUAGE"] = "en_AU:en_GB:en"
	assertDeepEqual(t, UserLanguages(), []string{"en_AU", "en_GB", "en"})
}

func TestNormalizeLanguages(t *testing.T) {
	restore := mockGetenv(map[string]string{
		"LANGUAGE": "en_AU:en_GB:en:C:fr",
	})
	defer restore()

	assertDeepEqual(t, normalizeLanguages([]string{"en_AU", "en_GB", "en", "C", "fr"}), []string{"en_AU", "en", "en_GB"})
}
