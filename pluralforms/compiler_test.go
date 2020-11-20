package pluralforms

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func assertEqual(t *testing.T, expected, got interface{}) {
	t.Helper()
	if !reflect.DeepEqual(expected, got) {
		t.Logf("%#v != %#v", expected, got)
		t.Fail()
	}
}

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

func TestParser(t *testing.T) {
	expr, err := Compile("1+n/5*10")
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, expr, addExpr{
		left: numberExpr{1},
		right: mulExpr{
			left: divExpr{
				left:  varExpr{},
				right: numberExpr{5},
			},
			right: numberExpr{10},
		},
	})

	expr, err = Compile("1-(2+n)/3")
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, expr, subExpr{
		left: numberExpr{1},
		right: divExpr{
			left: addExpr{
				left:  numberExpr{2},
				right: varExpr{},
			},
			right: numberExpr{3},
		},
	})

	expr, err = Compile("(n==1)?0:n>=2&&n<=4?1:2")
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, expr, ternaryExpr{
		test: eqExpr{
			left:  varExpr{},
			right: numberExpr{1},
		},
		ifTrue: numberExpr{0},
		ifFalse: ternaryExpr{
			test: andExpr{
				left: gteExpr{
					left:  varExpr{},
					right: numberExpr{2},
				},
				right: lteExpr{
					left:  varExpr{},
					right: numberExpr{4},
				},
			},
			ifTrue:  numberExpr{1},
			ifFalse: numberExpr{2},
		},
	})
}

func TestParserFailures(t *testing.T) {
	for _, expr := range []string{
		"1 + + 2",
		"n=1",
		"(n==1",
		"1 +",
		"m==1",
		"n=>1",
		"n>1 ? 0",
	} {
		_, err := Compile(expr)
		if err == nil {
			t.Logf("Expression %q unexpectedly compiled", expr)
			t.Fail()
		}
	}
}
