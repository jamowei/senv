package main

import (
	"fmt"
	"strings"
	"net/http"
	"net/http/httptest"
	"testing"
	"os"
)

var server *http.Server
var running = false
var profile = strings.Join([]string{"dev", "prod"}, ",")

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

func mockServer() *httptest.Server {
	f := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, jsonData)
	}

	server := httptest.NewServer(f)
	server.URL = "http://localhost:8888/test/dev,prod/master"
	return server
}

func TestMain(m *testing.M) {

	server := mockServer()
	defer server.Close()

	rootCmd.SetArgs([]string{"-n", "test", "env", "cmd", "/k", "echo", "${date}"})

	os.Exit(m.Run())
}
