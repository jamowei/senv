package senv

import (
	"testing"
	"net/http"
	"fmt"
	"io"
	"time"
	"context"
	"strings"
)

const (
	host    string = "127.0.0.1"
	port    string = "9999"
	name    string = "test"
	profile string = "dev"
	label   string = "master"
)

var jsonData string = fmt.Sprintf(
	`{
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
        "bill-to.address.postal": 48046,
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
}`, name, profile, label)

var server *http.Server

func startServer() {
	server = &http.Server{Addr: ":" + port}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, jsonData)
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

	conf := NewConfig(host, port, name, profile, label, formatKey, formatVal)
	if err := conf.Fetch(); err != nil {
		t.Fatalf("TestConfig: %s", err)
	}

	env := conf.enviroment;
	assertEqual(t, env.Name, name)
	assertEqual(t, env.Label, label)
	assertEqual(t, env.Profiles[0], profile)
	assertEqual(t, len(env.PropertySources), 2)

	if err := conf.Process(); err != nil {
		t.Fatalf("TestConfig: %s", err)
	}

	props := conf.Properties
	assertEqual(t, props["INVOICE"], "34843")
	assertEqual(t, props["BILL-TO_GIVEN"], "Test")
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
	out = strings.Replace(s, "\r\n", "", -1)
	out = strings.Replace(out, "\n", "", -1)
	return
}