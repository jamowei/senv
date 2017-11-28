package senv

import "testing"

const rawJson string = `{"bool":"true","float":"123.123","int":"123","string":"test"}`

func TestEnvironment(t *testing.T) {
	var src source
	jsonString := []byte(rawJson)
	err := src.UnmarshalJSON(jsonString)
	check(t, err)
	assertEqual(t, src.content["string"], "test")
	assertEqual(t, src.content["bool"], "true")
	assertEqual(t, src.content["int"], "123")
	assertEqual(t, src.content["float"], "123.123")

	res, err := src.MarshalJSON()
	check(t, err)
	assertEqual(t, string(res), rawJson)

	err = src.UnmarshalJSON([]byte("error"))
	checkInverse(t, err)
}
