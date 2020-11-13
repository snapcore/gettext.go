package gettext

import (
	"fmt"
	"os"
	"testing"
)

func TestEnGettext(t *testing.T) {
	for _, name := range []string{"messages.mo", "messages-be.mo", "messages-nohash.mo"} {
		name := name
		t.Run(name, func(t *testing.T) {
			file, err := os.Open("testdata/en/" + name)
			if err != nil {
				t.Fatal(err)
			}
			catalog, err := ParseMO(file)
			if err != nil {
				t.Fatal(err)
			}
			assert_equal(t, catalog.Gettext("greeting"), "Hello")
		})
	}
}

func TestEnNGettext(t *testing.T) {
	for _, name := range []string{"messages.mo", "messages-be.mo", "messages-nohash.mo"} {
		name := name
		t.Run(name, func(t *testing.T) {
			file, err := os.Open("testdata/en/" + name)
			if err != nil {
				t.Fatal(err)
			}
			catalog, err := ParseMO(file)
			if err != nil {
				t.Fatal(err)
			}
			assert_equal(t,
				fmt.Sprintf(catalog.NGettext("order %d beer", "order %d beers", 0), 0),
				"0 beers please",
			)
			assert_equal(t,
				fmt.Sprintf(catalog.NGettext("order %d beer", "order %d beers", 1), 1),
				"1 beer please",
			)
			assert_equal(t,
				fmt.Sprintf(catalog.NGettext("order %d beer", "order %d beers", 2), 2),
				"2 beers please",
			)
		})
	}
}

func TestNGettextBrokenMoNoCrash(t *testing.T) {
	file, err := os.Open("testdata/en-no-plural-forms/messages.mo")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := ParseMO(file)
	if err != nil {
		t.Fatal(err)
	}
	assert_equal(t,
		catalog.NGettext("order %d beer", "order %d beers", 1),
		"%d beer please",
	)
	assert_equal(t,
		catalog.NGettext("order %d beer", "order %d beers", 2),
		"%d beers please",
	)
}

func TestJaGettext(t *testing.T) {
	file, err := os.Open("testdata/ja/messages.mo")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := ParseMO(file)
	if err != nil {
		t.Fatal(err)
	}
	assert_equal(t, catalog.Gettext("greeting"), "こんいちは")
}

func TestJaNGettext(t *testing.T) {
	file, err := os.Open("testdata/ja/messages.mo")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := ParseMO(file)
	if err != nil {
		t.Fatal(err)
	}
	assert_equal(t,
		fmt.Sprintf(catalog.NGettext("order %d beer", "order %d beers", 0), 0),
		"ビールを0杯ください",
	)
	assert_equal(t,
		fmt.Sprintf(catalog.NGettext("order %d beer", "order %d beers", 1), 1),
		"ビールを1杯ください",
	)
	assert_equal(t,
		fmt.Sprintf(catalog.NGettext("order %d beer", "order %d beers", 2), 2),
		"ビールを2杯ください",
	)
}
