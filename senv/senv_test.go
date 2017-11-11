package senv

import (
	"testing"
	"net/http"
	"fmt"
	"io"
	"time"
)

const (
	host string = "127.0.0.1"
	port string = "9999"
	name string = "test"
	profile string = "dev"
	label string = "master"
)

var jsonData string = fmt.Sprintf(
	`{
			"name": "%s",
			"profiles": ["%s"],
			"label": "%s",
			"version": "132342454453",
			"state": null,
			"propertySources": [
				{
					"name": "application",
					"source": {
						"foo": "bar",
						"bar": {
							"foo": "foo ${foo}"
						}
					}
				}
			]
		}`, name, profile, label)


func startServer() *http.Server {
	srv := &http.Server{Addr: ":" + port}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, jsonData)
	})
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			// cannot panic, because this probably is an intentional close
			fmt.Printf("Httpserver: ListenAndServe() error: %s", err)
		}
	}()

	time.Sleep(2 * time.Second)

	// returning reference so caller can call Shutdown()
	return srv
}

func stopServer(srv *http.Server) {
	// now close the server gracefully ("shutdown")
	// timeout could be given instead of nil as a https://golang.org/pkg/context/
	if err := srv.Shutdown(nil); err != nil {
		panic(err) // failure/timeout shutting down the server gracefully
	}
}

func TestConfig_Recieve(t *testing.T) {
	srv := startServer()
	conf := NewConfig(host, port, name, profile, label)
	if err := conf.Recieve(); err != nil {
		t.Fatalf("TestRecieve: %s", err)
	}
	data := conf.Env;
	assertEqual(t, data.Name, name)
	assertEqual(t, data.Label, label)
	assertEqual(t, data.Profiles[0], profile)

	stopServer(srv)
}

//
//func TestServer(t *testing.T) {
//	startServer()
//	fmt.Scanln()
//}

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a == b {
		return
	}
	t.Fatal(fmt.Sprintf("%v != %v", a, b))
}