package pluralforms

import (
	"encoding/json"
	"os"
	"testing"
)

func TestCompiler(t *testing.T) {
	f, err := os.Open("testdata/plural_forms.json")
	if err != nil {
		t.Fatal(err)
	}
	dec := json.NewDecoder(f)
	var fixtures []struct {
		PluralForm string
		Fixture    []int
	}
	err = dec.Decode(&fixtures)
	if err != nil {
		t.Fatal(err)
	}
	for _, data := range fixtures {
		data := data
		t.Run(data.PluralForm, func(t *testing.T) {
			expr, err := Compile(data.PluralForm)
			if err != nil {
				t.Fatal(err)
			} else if expr == nil {
				t.Fatalf("'%s' compiled to nil", data.PluralForm)
			}
			for n, e := range data.Fixture {
				i := expr.Eval(uint32(n))
				if i != e {
					t.Logf("n = %d, expected %d, got %d, compiled to %s", n, e, i, expr)
					t.Fail()
				}
			}
		})
	}
}
