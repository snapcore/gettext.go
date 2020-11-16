// Implements gettext in pure Go with Plural Forms support.

package gettext

import (
	"fmt"
	"os"
	"path"
	"sync"
)

// Translations holds the translations in the different locales your app
// supports. Use NewTranslations to create an instance.
type Translations struct {
	// Ideally NewTranslations would return a *Translations
	// pointer.  As we don't want the mutex protecting the catalog
	// cache to be copied, we embed a pointer to an ancillary
	// struct holding our data.
	*translations
}

type translations struct {
	mu       sync.Mutex
	cache    map[string]*mocatalog
	root     string
	domain   string
	resolver PathResolver
}

// PathResolver resolves a path to a mo file
type PathResolver func(root string, locale string, domain string) string

// DefaultResolver resolves paths in the standard format of:
// <root>/<locale>/LC_MESSAGES/<domain>.mo
func DefaultResolver(root string, locale string, domain string) string {
	return path.Join(root, locale, "LC_MESSAGES", fmt.Sprintf("%s.mo", domain))
}

// NewTranslations is the main entry point for gogettext. Use this to set up
// the locales for your app.
// root is the root of your locale folder, domain the domain you want to load
// and resolver a function that resolves mo file paths.
// If your structure is <root>/<locale>/LC_MESSAGES/<domain>.mo, you can use
// DefaultResolver.
func NewTranslations(root string, domain string, resolver PathResolver) Translations {
	return Translations{&translations{
		root:     root,
		resolver: resolver,
		domain:   domain,
		cache:    map[string]*mocatalog{},
	}}
}

// Preload a list of locales (if they're available). This is useful if you want
// to limit IO to a specific time in your app, for example startup. Subsequent
// calls to Preload or Locale using a locale given here will not do any IO.
func (t Translations) Preload(locales ...string) {
	for _, locale := range locales {
		t.load(locale)
	}
}

func (t Translations) load(locale string) *mocatalog {
	t.mu.Lock()
	defer t.mu.Unlock()

	if catalog, ok := t.cache[locale]; ok {
		return catalog
	}

	t.cache[locale] = nil
	path := t.resolver(t.root, locale, t.domain)
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	catalog, err := parseMO(f)
	if err != nil {
		return nil
	}
	t.cache[locale] = catalog
	return catalog
}

// Locale returns the catalog translations for a list of locales.
//
// If translations are not found in the first locale, the each
// subsequent one is consulted until a match is found.  If no match is
// found, the original strings are returned.
func (t Translations) Locale(languages ...string) Catalog {
	var mos []*mocatalog
	for _, lang := range normalizeLanguages(languages) {
		mo := t.load(lang)
		if mo != nil {
			mos = append(mos, mo)
		}
	}
	return Catalog{mos}
}

// UserLocale returns the catalog translations for the user's Locale.
func (t Translations) UserLocale() Catalog {
	return t.Locale(UserLanguages()...)
}
