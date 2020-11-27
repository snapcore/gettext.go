package gettext

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"
)

func TestNewTranslations(t *testing.T) {
	// The result of NewTranslations can be assigned to a variable
	// using the deprecated Translations alias.
	var trans Translations = NewTranslations("localeDir", "domain", DefaultResolver)
	assert_equal(t, trans.Name, "domain")
	assert_equal(t, trans.LocaleDir, "localeDir")
}

func TestNullTranslations(t *testing.T) {
	translations := &TextDomain{Name: "messages", LocaleDir: "."}
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

func TestMessageContext(t *testing.T) {
	trans := NewTranslations("testdata/", "messages", my_resolver)
	es := trans.Locale("es")

	// The context is used to distinguish identical message IDs
	assert_equal(t, es.PGettext("knot", "bow"), "lazo")
	assert_equal(t, es.PGettext("weapon", "bow"), "arco")

	// A context can be used for ngettext style lookups too.
	assert_equal(t, es.PNGettext("knot", "%d bow", "%d bows", 1), "%d lazo")
	assert_equal(t, es.PNGettext("knot", "%d bow", "%d bows", 2), "%d lazos")
	assert_equal(t, es.PNGettext("weapon", "%d bow", "%d bows", 1), "%d arco")
	assert_equal(t, es.PNGettext("weapon", "%d bow", "%d bows", 2), "%d arcos")

	// There is no contextless translation
	assert_equal(t, es.Gettext("bow"), "bow")
	assert_equal(t, es.NGettext("%d bow", "%d bows", 1), "%d bow")

	// With no catalog, the message ID is returned and context ignored
	empty := trans.Locale()
	assert_equal(t, empty.PGettext("knot", "bow"), "bow")
	assert_equal(t, empty.PNGettext("knot", "%d bow", "%d bows", 1), "%d bow")
}

func TestFallbackCatalog(t *testing.T) {
	translations := &TextDomain{Name: "messages", LocaleDir: "testdata/", PathResolver: my_resolver}
	cat := translations.Locale("en_AU", "en")
	// A translation from en_AU
	assert_equal(t, cat.Gettext("greeting"), "G'day")
	// A translation from en
	assert_equal(t, cat.NGettext("order %d beer", "order %d beers", 0), "%d beers please")

	// Loading the catalogs in the other order shadows the en_AU string
	cat = translations.Locale("en", "en_AU")
	assert_equal(t, cat.Gettext("greeting"), "Hello")
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

	translations := &TextDomain{Name: "messages", LocaleDir: dir}
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
	translations := &TextDomain{Name: "messages", LocaleDir: "testdata/", PathResolver: my_resolver}

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
	translations := &TextDomain{Name: "messages", LocaleDir: "testdata/", PathResolver: po_resolver}
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

func TestUseLangpacks(t *testing.T) {
	dir, err := ioutil.TempDir("", "gogettext")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	oldLangpackDir := langpackLocaleDir
	defer func() {
		langpackLocaleDir = oldLangpackDir
	}()
	localeDir := filepath.Join(dir, "locale")
	langpackLocaleDir = filepath.Join(dir, "langpack")

	err = os.MkdirAll(filepath.Join(localeDir, "en", "LC_MESSAGES"), 0777)
	if err != nil {
		t.Fatal(err)
	}
	err = os.MkdirAll(filepath.Join(langpackLocaleDir, "ja", "LC_MESSAGES"), 0777)
	if err != nil {
		t.Fatal(err)
	}
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	err = os.Symlink(
		filepath.Join(curDir, "testdata/en/messages.mo"),
		filepath.Join(localeDir, "en", "LC_MESSAGES", "messages.mo"),
	)
	if err != nil {
		t.Fatal(err)
	}
	err = os.Symlink(
		filepath.Join(curDir, "testdata/ja/messages.mo"),
		filepath.Join(langpackLocaleDir, "ja", "LC_MESSAGES", "messages.mo"),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Without langpack support, english is chosen
	domain := &TextDomain{Name: "messages", LocaleDir: localeDir}
	locale := domain.Locale("ja", "en")
	assert_equal(t, locale.Gettext("greeting"), "Hello")

	// With langpack support enabled, the preferred Japanese
	// translation from the langpack is chosen instead.
	domain.UseLangpacks = true
	locale = domain.Locale("ja", "en")
	assert_equal(t, locale.Gettext("greeting"), "こんいちは")
}
