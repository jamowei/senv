package main

import (
	"fmt"
	"os"
	"testing"
)

func TestSetEnvVars(t *testing.T) {
	const key string = "GO_TEST_VAR"
	os.Setenv(key, "test")
	defer os.Unsetenv(key)

	data := map[string]string{
		"GO_TEST_VAR1": "foo",
		"GO_TEST_VAR2": "bar",
	}
	defer func(m map[string]string) {
		for k := range m {
			os.Unsetenv(k)
		}
	}(data)

	setEnvVars(data, false)
	for k, v := range data {
		assertEqual(t, v, os.Getenv(k))
	}

	data[key] = "new test"
	if err := setEnvVars(data, false); err == nil {
		t.Fatalf("SetEnvVars: should fail on property \"%s\"=\"%s\"", key, data[key])
	}

	if err := setEnvVars(data, true); err != nil {
		t.Fatalf("SetEnvVars: should not fail on property \"%s\"=\"%s\"", key, data[key])
	}
	assertEqual(t, data[key], os.Getenv(key))
}

func TestFormatter(t *testing.T) {
	assertEqual(t, formatKey("foo.bar"), "FOO_BAR")
	assertEqual(t, formatVal("foo\nbar\r\nbars"), "foo bar bars")
}

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	t.Helper()
	if a == b {
		return
	}
	t.Fatal(fmt.Sprintf("%v != %v", a, b))
}
