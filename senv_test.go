package senv

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

const (
	host     string = "127.0.0.1"
	port     string = "9999"
	name     string = "test"
	badjson  string = "badjson"
	badprops string = "badprops"
	notfound string = "notfound"
	label    string = "master"
)

var profiles []string = []string{"dev", "prod"}

var jsonData string = `{
  "Name": "test",
  "Profiles": [
    "dev",
	"prod"
  ],
  "Label": "master",
  "Version": "f65dfb395b177a3eac3bc29d3c3829e47543dcb2",
  "State": null,
  "PropertySources": [
    {
      "Name": "file://test.yml",
      "Source": {
        "invoice": 34843,
        "date": "${test.date}",
        "bill-to.given": "${test.Name}",
        "bill-to.family": "Dumars",
        "bill-to.address.lines": "458 Walkman Dr.\nSuite #292\n",
        "bill-to.address.city": "Royal Oak",
        "bill-to.address.State": "MI",
        "bill-to.given": "${test.Name}",
        "ship-to.given": "${test.Name}",
        "ship-to.family": "Dumars",
        "ship-to.address.lines": "458 Walkman Dr.\nSuite #292\n",
        "ship-to.address.city": "Royal Oak",
        "ship-to.address.State": "MI",
        "ship-to.address.postal": 48046,
        "product[0].sku": "BL394D",
        "product[0].quantity": 4,
        "product[0].description": "Basketball",
        "product[0].price": "${test.price}",
        "product[1].sku": "BL4438H",
        "product[1].quantity": 1,
        "product[1].description": "Super Hoop",
        "product[1].price": 2392.0,
        "tax": 251.42,
        "total[0]": 4443.52,
        "total[1]": 12342.23,
        "total[2]": 23.2342344,
        "total[3]": "Hallo",
        "comments": "Late afternoon is best. Backup contact is Nancy Billsmer @ 338-4338."
      }
    },
    {
      "Name": "file://application.yml",
      "Source": {
		"invoice": "100",
        "test.Name": "Test",
        "test.date": "2001-01-23",
        "test.price": 450.0
      }
    }
  ]
}`

var jsonDataWrong string = `{
  "Name": "test",
  "Profiles": [
    "dev",
	"prod"
  ],
  "Label": "master",
  "Version": "f65dfb395b177a3eac3bc29d3c3829e47543dcb2",
  "State": null,
  "PropertySources": [
    {
      "Name": "file://test.yml",
      "Source": {
        "invoice": 34843,
        "date": "${test.date}",
        "bill-to.given": "${test.Name}",
        "bill-to.family": "Dumars"
	  }
	}
  ]
}`

var server *http.Server

func startServer() {
	server = &http.Server{Addr: ":" + port}

	http.HandleFunc(fmt.Sprintf("/%s/%s/%s", name, strings.Join(profiles, ","), label), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, jsonData)
	})

	http.HandleFunc(fmt.Sprintf("/%s/%s/%s", badjson, strings.Join(profiles, ","), label), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, "this is no json")
	})

	http.HandleFunc(fmt.Sprintf("/%s/%s/%s", badprops, strings.Join(profiles, ","), label), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, jsonDataWrong)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	go func() {
		server.ListenAndServe()
	}()
}

func stopServer() {
	d := time.Now().Add(1 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), d)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		panic(err) // failure/timeout shutting down the server gracefully
	}
}

func TestConfig(t *testing.T) {
	startServer()
	defer stopServer()
	time.Sleep(1 * time.Second)

	conf := NewConfig(host, port, name, profiles, label, formatKey, formatVal)

	if err := conf.Fetch(); err != nil {
		t.Fatalf("TestConfig: %s", err)
	}

	env := conf.environment
	assertEqual(t, env.Name, name)
	assertEqual(t, env.Label, label)
	assertEqual(t, env.Profiles[0], profiles[0])
	assertEqual(t, env.Profiles[1], profiles[1])
	assertEqual(t, len(env.PropertySources), 2)

	if err := conf.Process(false); err != nil {
		t.Fatalf("TestConfig: %s", err)
	}

	props := conf.Properties
	assertEqual(t, props["INVOICE"], "34843")
	assertEqual(t, props["BILL-TO_GIVEN"], "Test")
	assertEqual(t, props["SHIP-TO_ADDRESS_LINES"], "458 Walkman Dr. Suite #292 ")
	assertEqual(t, props["TOTAL[1]"], "12342.23")
}

func TestFailures(t *testing.T) {
	startServer()
	defer stopServer()
	time.Sleep(1 * time.Second)

	cfg1 := NewConfig(host, port, badjson, profiles, label, formatKey, formatVal)
	cfg2 := NewConfig(host, port, notfound, profiles, label, formatKey, formatVal)
	cfg3 := NewConfig(host, port, badprops, profiles, label, formatKey, formatVal)

	if err := cfg1.Fetch(); err == nil {
		t.Fatal("TestFailure1: should fail on wrong json")
	} else {
		fmt.Printf("Expected error1: %#v\n", err)
	}

	if err := cfg2.Fetch(); err == nil {
		t.Fatal("TestFailure2: should fail on wrong url")
	} else {
		fmt.Printf("Expected error2: %#v\n", err)
	}

	if err := cfg3.Fetch(); err != nil {
		t.Fatal("TestFailure3: should fail on parsing properties")
	} else {
		if err := cfg3.Process(true); err == nil {
			t.Fatal("TestFailure3: should fail on parsing properties")
		} else {
			fmt.Printf("Expected error3: %#v\n", err)
		}
	}

}

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	t.Helper()
	if a == b {
		return
	}
	t.Fatal(fmt.Sprintf("%v != %v", a, b))
}

func formatKey(in string) (out string) {
	out = strings.Replace(in, ".", "_", -1)
	out = strings.ToUpper(out)
	return
}

func formatVal(s string) (out string) {
	out = strings.Replace(s, "\r\n", " ", -1)
	out = strings.Replace(out, "\n", " ", -1)
	return
}
