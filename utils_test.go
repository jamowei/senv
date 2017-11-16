package main
//
//import (
//	"testing"
//	"fmt"
//)
//
//const (
//	opener string = "${"
//	closer string = "}"
//)
//
//func TestStringValReplacer(t *testing.T) {
//	// test if it replaces all correct
//	data := map[string]string{
//		"a1.b1.c1": "foo ${a1.b2}!",
//		"a1.b2": "bar",
//		"a2.b3": "100",
//		"a2.b4": "${a2.b3}x ${a1.b1.c1}!",
//		"a3.b5": "pi=${a3.b6}",
//		"a3.b6": "3,41...",
//		"a3.b7": "true",
//		"a3.b8[0]": "1.23",
//		"a3.b8[1].c2": "go is ${a3.b7}ly amazing",
//	}
//
//	smap := make(map[string]string)
//
//	replacer := StringValReplacer{opener, closer, true}
//	err := replacer.Replace(data, smap);
//	if err != nil {
//		t.Fatalf("Error in Replacer: %s", err)
//	}
//	assertEqual(t, smap["a1.b1.c1"], "foo bar!")
//	assertEqual(t, smap["a2.b4"], "100x foo bar!!")
//	assertEqual(t, smap["a2.b4"], "100x foo bar!!")
//	assertEqual(t, smap["a3.b5"], "pi=3,41...")
//	assertEqual(t, smap["a3.b8[0]"], "1.23")
//	assertEqual(t, smap["a3.b8[1].c2"], "go is truely amazing")
//
//	//test if it fails correct
//	data = map[string]string{
//		"a1.b1": "foo ${error}",
//	}
//	if err := replacer.Replace(data, smap); err == nil {
//		t.Fatalf("Error in Replacer: doesnt detect missing property value")
//	}
//}
//
//func assertEqual(t *testing.T, a interface{}, b interface{}) {
//	t.Helper()
//	if a == b {
//		return
//	}
//	t.Fatal(fmt.Sprintf("%v != %v", a, b))
//}