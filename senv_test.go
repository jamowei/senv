package senv

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

const (
	host      = "127.0.0.1"
	port      = "9999"
	wrongport = "8888"
	name      = "test"
	badjson   = "badjson"
	badprops  = "badprops"
	file      = "test.txt"
	label     = "master"
)

var profiles = []string{"dev", "prod"}

var running = false

var jsonData = `{
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
        "ship-to.address.city": "${unknown:Royal Oak}",
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

var jsonDataWrong = `{
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

var plainData = "this is a test!"

var server *http.Server

func startServer() {
	if running == false {
		running = true

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

		http.HandleFunc(fmt.Sprintf("/%s/%s/%s/%s", name, strings.Join(profiles, ","), label, file), func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain;charset=UTF-8")
			io.WriteString(w, plainData)
		})

		go func() {
			server.ListenAndServe()
		}()
	}
}

func stopServer() {
	if running == true {
		d := time.Now().Add(1 * time.Second)
		ctx, cancel := context.WithDeadline(context.Background(), d)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			panic(err) // failure/timeout shutting down the server gracefully
		}
		running = false
	}
}

func TestConfig(t *testing.T) {
	startServer()

	time.Sleep(1 * time.Second)

	conf := NewConfig(host, port, name, profiles, label)

	err := conf.Fetch(true, true)
	check(t, err)
	env := conf.environment
	assertEqual(t, env.Name, name)
	assertEqual(t, env.Label, label)
	assertEqual(t, env.Profiles[0], profiles[0])
	assertEqual(t, env.Profiles[1], profiles[1])
	assertEqual(t, len(env.PropertySources), 2)

	err = conf.Process()
	check(t, err)
	props := conf.Properties
	assertEqual(t, props["invoice"], "34843")
	assertEqual(t, props["bill-to.given"], "Test")
	assertEqual(t, props["ship-to.address.lines"], "458 Walkman Dr.\nSuite #292\n")
	assertEqual(t, props["total[1]"], "12342.23")
	assertEqual(t, props["ship-to.address.city"], "Royal Oak")

	err = conf.FetchFile(file, true, true)
	check(t, err)
	err = conf.FetchFile(file, false, true)
	check(t, err)
	cnt, err := ioutil.ReadFile(file)
	check(t, err)
	assertEqual(t, string(cnt), plainData)
	os.Remove(file)

}

func TestFailures(t *testing.T) {
	startServer()
	defer stopServer()
	time.Sleep(1 * time.Second)

	cfg1 := NewConfig(host, port, badjson, profiles, label)
	cfg2 := NewConfig(host, wrongport, name, profiles, label)
	cfg3 := NewConfig(host, port, badprops, profiles, label)

	err := cfg1.Fetch(false, true)
	checkInverse(t, err)

	err = cfg2.Fetch(false, true)
	checkInverse(t, err)

	err = cfg2.FetchFile("test.txt", true, true)
	checkInverse(t, err)

	err = cfg3.Fetch(false, true)
	check(t, err)
	err = cfg3.Process()
	checkInverse(t, err)

}

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	t.Helper()
	if a == b {
		return
	}
	t.Fatal(fmt.Sprintf("%v != %v", a, b))
}

func check(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(t.Name(), err)
	}
}

func checkInverse(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal(t.Name(), "should fail, but didnt")
	}
}