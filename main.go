package main

import (
	"flag"
	"strings"
	"os"
	"net/http"
	"fmt"
	"encoding/json"
)

const (
	host_default string = "127.0.0.1"
	port_default string = "8888"
	name_default string = "application"
	profile_default string = "default"
	label_default string = "master"
)

var (
	host    string
	port    string
	name    string
	profile string
	label   string
)

type Config struct {
	Name string
	Profiles []string
	Label string
	Version string
	State string
	PropertySources []Sources
}

type Sources struct {
	Name string
	Source map[string]string
}

func main() {
	params()
	conf := new(Config)
	if err := conf.recieve(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		conf.process()
	}
}

func (config *Config) process() {
	if config.PropertySources != nil {
		for _, prop := range config.PropertySources {
			for key, val := range prop.Source {
				println(prop.Name, key, val)
			}
		}
	}
}

func (config *Config) recieve() (err error) {
	url := fmt.Sprintf("http://%s:%s/%s/%s/%s", host, port, name, profile, label)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	if resp.Body != nil {
		defer resp.Body.Close()
		err := json.NewDecoder(resp.Body).Decode(config)
		if err != nil {
			return err
		}
	}
	return
}

func params() {
	flag.StringVar(&host, "host", host_default, "host")
	flag.StringVar(&port, "port", port_default, "port")
	flag.StringVar(&name, "name", name_default, "name")
	flag.StringVar(&profile, "profile", profile_default, "profile")
	flag.StringVar(&label, "label", label_default, "port")
	if help := flag.Arg(0); strings.HasSuffix(help, "h") || strings.HasSuffix(help, "help") {
		flag.Usage()
		os.Exit(0)
	}
	flag.Parse()
}
