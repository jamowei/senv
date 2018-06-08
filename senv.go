package senv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// Replacer replaces all variables given in the first string
// with the appropriate values of the specified key in the second map
// and give it replaced back or error otherwise.
type Replacer interface {
	Replace(str string, m map[string]string) (string, error)
}

// Config hold the information which is needed to receive the
// json data from the spring config server and parse and transform them correctly.
type Config struct {
	Host, Port, Name, Profile, Label string
	Replacer                         Replacer
	environment                      *environment
	Properties                       map[string]string
}

// NewConfig returns a new Config as pointer value with a default Replacer for
// spring cloud config.
func NewConfig(host string, port string, name string, profiles []string, label string) *Config {
	return &Config{host, port, name, strings.Join(profiles, ","), label,
		&SpringReplacer{"${", "}", ":"},
		nil, nil}
}

// Fetch fetches the json data from the spring config server, see:
// https://cloud.spring.io/spring-cloud-config/single/spring-cloud-config.html#_quick_start
func (cfg *Config) Fetch(showJson bool, verbose bool) error {
	env := &environment{}
	url := fmt.Sprintf("http://%s:%s/%s/%s/%s", cfg.Host, cfg.Port, cfg.Name, cfg.Profile, cfg.Label)

	if verbose {
		fmt.Fprintln(os.Stderr, "Fetching config from server at:", url)
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(env)
	if err != nil {
		return err
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Located environment: name=%#v, profiles=%v, label=%#v, version=%#v, state=%#v\n",
			env.Name, env.Profiles, env.Label, env.Version, env.State)
	}

	if showJson {
		jsonStr, _ := json.MarshalIndent(env, "", "    ")
		fmt.Println(string(jsonStr))
	}

	cfg.environment = env
	return nil
}

// FetchFile download a file from the spring config server, see:
// https://cloud.spring.io/spring-cloud-config/single/spring-cloud-config.html#_serving_plain_text
func (cfg *Config) FetchFile(filename string, printFile bool, verbose bool) error {
	url := fmt.Sprintf("http://%s:%s/%s/%s/%s/%s", cfg.Host, cfg.Port, cfg.Name, cfg.Profile, cfg.Label, filename)

	if verbose {
		fmt.Fprintf(os.Stderr, "Fetching file \"%s\" from server at: %s\n", filename, url)
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if printFile {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		fmt.Println(buf.String())
	} else {
		out, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer out.Close()

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return err
		}
	}
	return nil
}

// Process use given Replacer and formatter functions to process
// the fetched json data and must be called after Fetch
func (cfg *Config) Process() error {
	env := cfg.environment
	if env != nil && env.PropertySources != nil {
		//merge propertySources into one map
		mergedProperties := mergeProps(env.PropertySources)

		if cfg.Replacer != nil {

			//replace variables
			replacedProperties := make(map[string]string)
			for key, val := range mergedProperties {
				nVal, err := cfg.Replacer.Replace(val, mergedProperties)
				if err != nil {
					return err
				}
				replacedProperties[key] = nVal
			}
			cfg.Properties = replacedProperties
		}
	}
	return nil
}

// reverse iterating over all propertySource for overriding
// more specific values with the same key
func mergeProps(pSources []propertySource) (merged map[string]string) {
	merged = make(map[string]string)
	for i := len(pSources) - 1; i >= 0; i-- {
		data := pSources[i]
		// merge all propertySources to one map
		for key, val := range data.Source.content {
			merged[key] = val
		}
	}
	return
}

// SpringReplacer needs the opening and closing string
// for detecting a variables that must be replaced.
type SpringReplacer struct {
	Opener      string
	Closer      string
	Default		string
}

// Replace replaces all variables with the defined opening and
// closing strings with and default separator the value of the
// key or when available with the default value.
func (rpl *SpringReplacer) Replace(str string, m map[string]string) (string, error) {
	var f, s int
	f = strings.Index(str, rpl.Opener) + len(rpl.Opener)
	for f-len(rpl.Opener) > -1 {
		s = f + strings.Index(str[f:], rpl.Closer)
		key := str[f:s]
		var val string
		var ok, def bool
		i := strings.Index(key, rpl.Default)
		if i > 0 {
			def = true
			val, ok = m[key[:i]]
		} else {
			val, ok = m[key]
		}
		if !ok {
			if def {
				val = key[i+1:]
			} else {
				return str, fmt.Errorf("cannot find value for key %s in \"%s\"", key, str)
			}
		}
		str = str[:f-len(rpl.Opener)] + val + str[s+len(rpl.Closer):]
		f = strings.Index(str, rpl.Opener) + len(rpl.Opener)
	}
	return str, nil
}
