package senv

import (
	"testing"
	"math"
	"strconv"
)

const (
	sep       string = "_"
	upperCase bool   = true
)

func TestJsonFlattener_Flatten(t *testing.T) {
	data := map[string]interface{}{
		"foo": map[string]interface{}{
			"bar": map[string]interface{}{
				"foo": "foo bar!",
			},
			"foo": "bar",
		},
		"bar": map[string]interface{}{
			"foo": 100,
			"bar": "bars",
		},
	}

	flattener := StringMapFlattener{}

	res, _ := flattener.Flatten(data)
	assertEqual(t, "foo bar!", res["foo.bar.foo"])
	assertEqual(t, "bar", res["foo.foo"])
	assertEqual(t, 100, res["bar.foo"])
	assertEqual(t, "bars", res["bar.bar"])
}

//TODO: check case sensivity from config server
func TestFlatMapReplacer_Replace(t *testing.T) {
	// test if it replaces all correct
	data := map[string]interface{}{
		"a1": map[string]interface{}{
			"b1": map[string]interface{}{
				"c1": "foo ${a1.b2}!",
			},
			"b2": "bar",
		},
		"a2": map[string]interface{}{
			"b3": 100,
			"b4": "${a2.b3}x ${a1.b1.c1}!",
		},
		"a3": map[string]interface{}{
			"b5": "pi=${a3.b6}",
			"b6": math.Pi,
			"b7": true,
			"b8": map[string]interface{} {
				"c2": "go is ${a3.b7}ly amazing",
			},
		},
	}

	flattener := StringMapFlattener{}
	fmap, _ := flattener.Flatten(data)
	replacer := FlatMapReplacer{"${", "}", true, &EnvKeyFormatter{"_", true}}
	smap, err := replacer.Replace(fmap);
	if err != nil {
		t.Fatalf("Error in Replacer: %s", err)
	}
	assertEqual(t, smap["A1_B1_C1"], "foo bar!")
	assertEqual(t, smap["A2_B4"], "100x foo bar!!")
	assertEqual(t, smap["A2_B4"], "100x foo bar!!")
	assertEqual(t, smap["A3_B5"], "pi="+strconv.FormatFloat(math.Pi, 'f', -1, 64))
	assertEqual(t, smap["A3_B8_C2"], "go is truely amazing")

	//test if it fails correct
	data = map[string]interface{}{
		"a1": map[string]interface{}{
			"b1": "foo ${error}",
		},
	}
	fmap, _ = flattener.Flatten(data)
	if _, err := replacer.Replace(fmap); err == nil {
		t.Fatalf("Error in Replacer: doesnt detect missing property value")
	}
}