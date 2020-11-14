package gettext

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestNullTranslations(t *testing.T) {
	translations := NewTranslations(".", "messages", DefaultResolver)
	en := translations.Locale("en")
	en_gettext := en.Gettext("mymsgid")
	assert_equal(t, en_gettext, "mymsgid")
	en_ngettext_0 := en.NGettext("mymsgid", "mymsgidp", 0)
	assert_equal(t, en_ngettext_0, "mymsgidp")
	en_ngettext_1 := en.NGettext("mymsgid", "mymsgidp", 1)
	assert_equal(t, en_ngettext_1, "mymsgid")
	en_ngettext_2 := en.NGettext("mymsgid", "mymsgidp", 2)
	assert_equal(t, en_ngettext_2, "mymsgidp")
	ja := translations.Locale("ja")
	ja_gettext := ja.Gettext("mymsgid")
	assert_equal(t, ja_gettext, "mymsgid")
	ja_ngettext_0 := ja.NGettext("mymsgid", "mymsgidp", 0)
	assert_equal(t, ja_ngettext_0, "mymsgidp")
	ja_ngettext_1 := ja.NGettext("mymsgid", "mymsgidp", 1)
	assert_equal(t, ja_ngettext_1, "mymsgid")
	ja_ngettext_2 := ja.NGettext("mymsgid", "mymsgidp", 2)
	assert_equal(t, ja_ngettext_2, "mymsgidp")
}

func my_resolver(root string, locale string, domain string) string {
	return path.Join(root, locale, fmt.Sprintf("%s.mo", domain))
}

func TestRealTranslations(t *testing.T) {
	translations := NewTranslations("testdata/", "messages", my_resolver)
	en := translations.Locale("en")
	assert_equal(t, en.Gettext("greeting"), "Hello")
	assert_equal(t,
		fmt.Sprintf(en.NGettext("order %d beer", "order %d beers", 0), 0),
		"0 beers please",
	)
	assert_equal(t,
		fmt.Sprintf(en.NGettext("order %d beer", "order %d beers", 1), 1),
		"1 beer please",
	)
	assert_equal(t,
		fmt.Sprintf(en.NGettext("order %d beer", "order %d beers", 2), 2),
		"2 beers please",
	)
	ja := translations.Locale("ja")
	assert_equal(t, ja.Gettext("greeting"), "こんいちは")
	assert_equal(t,
		fmt.Sprintf(ja.NGettext("order %d beer", "order %d beers", 0), 0),
		"ビールを0杯ください",
	)
	assert_equal(t,
		fmt.Sprintf(ja.NGettext("order %d beer", "order %d beers", 1), 1),
		"ビールを1杯ください",
	)
	assert_equal(t,
		fmt.Sprintf(ja.NGettext("order %d beer", "order %d beers", 2), 2),
		"ビールを2杯ください",
	)
	de := translations.Locale("de")
	assert_equal(t, de.Gettext("greeting"), "greeting")
	assert_equal(t,
		fmt.Sprintf(de.NGettext("order %d beer", "order %d beers", 0), 0),
		"order 0 beers",
	)
	assert_equal(t,
		fmt.Sprintf(de.NGettext("order %d beer", "order %d beers", 1), 1),
		"order 1 beer",
	)
	assert_equal(t,
		fmt.Sprintf(de.NGettext("order %d beer", "order %d beers", 2), 2),
		"order 2 beers",
	)
}

func TestPreload(t *testing.T) {
	dir, err := ioutil.TempDir("", "gogettext")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	err = os.MkdirAll(path.Join(dir, "en", "LC_MESSAGES"), 0777)
	if err != nil {
		t.Fatal(err)
	}
	err = os.MkdirAll(path.Join(dir, "ja", "LC_MESSAGES"), 0777)
	if err != nil {
		t.Fatal(err)
	}
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	err = os.Symlink(
		path.Join(curDir, "testdata/en/messages.mo"),
		path.Join(dir, "en", "LC_MESSAGES", "messages.mo"),
	)
	if err != nil {
		t.Fatal(err)
	}
	err = os.Symlink(
		path.Join(curDir, "testdata/ja/messages.mo"),
		path.Join(dir, "ja", "LC_MESSAGES", "messages.mo"),
	)
	if err != nil {
		t.Fatal(err)
	}

	translations := NewTranslations(dir, "messages", DefaultResolver)
	translations.Preload("en")
	err = os.Remove(path.Join(dir, "en", "LC_MESSAGES", "messages.mo"))
	if err != nil {
		t.Fatal(err)
	}
	err = os.Remove(path.Join(dir, "ja", "LC_MESSAGES", "messages.mo"))
	if err != nil {
		t.Fatal(err)
	}
	// EN is preloaded so should still work without the files there
	en := translations.Locale("en")
	assert_equal(t, en.Gettext("greeting"), "Hello")
	assert_equal(t,
		fmt.Sprintf(en.NGettext("order %d beer", "order %d beers", 0), 0),
		"0 beers please",
	)
	assert_equal(t,
		fmt.Sprintf(en.NGettext("order %d beer", "order %d beers", 1), 1),
		"1 beer please",
	)
	assert_equal(t,
		fmt.Sprintf(en.NGettext("order %d beer", "order %d beers", 2), 2),
		"2 beers please",
	)
	// JA wasn't preloaded so should do nothing since files aren't there
	ja := translations.Locale("ja")
	assert_equal(t, ja.Gettext("greeting"), "greeting")
	assert_equal(t,
		fmt.Sprintf(ja.NGettext("order %d beer", "order %d beers", 0), 0),
		"order 0 beers",
	)
	assert_equal(t,
		fmt.Sprintf(ja.NGettext("order %d beer", "order %d beers", 1), 1),
		"order 1 beer",
	)
	assert_equal(t,
		fmt.Sprintf(ja.NGettext("order %d beer", "order %d beers", 2), 2),
		"order 2 beers",
	)
}

func TestUserLocale(t *testing.T) {
	translations := NewTranslations("testdata/", "messages", my_resolver)

	restore := mockGetenv(map[string]string{
		"LANGUAGE": "fr_FR:ja_JP:en",
	})
	defer restore()

	// The first available catalog is returned
	cat := translations.UserLocale()
	assert_equal(t, cat.Gettext("greeting"), "こんいちは")

	// If no matches are found, a NULL catalog is returned
	restore = mockGetenv(map[string]string{
		"LANGUAGE": "de_DE",
	})
	defer restore()
	cat = translations.UserLocale()
	assert_equal(t, cat.Gettext("greeting"), "greeting")
}

func po_resolver(root string, locale string, domain string) string {
	return path.Join(root, locale, fmt.Sprintf("%s.po", domain))
}

func TestNotMoFile(t *testing.T) {
	translations := NewTranslations("testdata/", "messages", po_resolver)
	en := translations.Locale("en")
	assert_equal(t, en.Gettext("greeting"), "greeting")
	assert_equal(t,
		fmt.Sprintf(en.NGettext("order %d beer", "order %d beers", 0), 0),
		"order 0 beers",
	)
	assert_equal(t,
		fmt.Sprintf(en.NGettext("order %d beer", "order %d beers", 1), 1),
		"order 1 beer",
	)
	assert_equal(t,
		fmt.Sprintf(en.NGettext("order %d beer", "order %d beers", 2), 2),
		"order 2 beers",
	)
	ja := translations.Locale("ja")
	assert_equal(t, ja.Gettext("greeting"), "greeting")
	assert_equal(t,
		fmt.Sprintf(ja.NGettext("order %d beer", "order %d beers", 0), 0),
		"order 0 beers",
	)
	assert_equal(t,
		fmt.Sprintf(ja.NGettext("order %d beer", "order %d beers", 1), 1),
		"order 1 beer",
	)
	assert_equal(t,
		fmt.Sprintf(ja.NGettext("order %d beer", "order %d beers", 2), 2),
		"order 2 beers",
	)
	de := translations.Locale("de")
	assert_equal(t, de.Gettext("greeting"), "greeting")
	assert_equal(t,
		fmt.Sprintf(de.NGettext("order %d beer", "order %d beers", 0), 0),
		"order 0 beers",
	)
	assert_equal(t,
		fmt.Sprintf(de.NGettext("order %d beer", "order %d beers", 1), 1),
		"order 1 beer",
	)
	assert_equal(t,
		fmt.Sprintf(de.NGettext("order %d beer", "order %d beers", 2), 2),
		"order 2 beers",
	)

}
