package gettext

import (
	"reflect"
	"testing"
)

func assert_equal(t *testing.T, expected string, got string) {
	t.Helper()
	if expected != got {
		t.Logf("%s != %s", expected, got)
		t.Fail()
	}
}

func assertDeepEqual(t *testing.T, expected, got interface{}) {
	t.Helper()
	if !reflect.DeepEqual(expected, got) {
		t.Logf("%v != %v", expected, got)
		t.Fail()
	}
}
