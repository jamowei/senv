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

// ValReplacer replaces all variables given in the values
// of the first map with the appropriate values of the specified key
// and stores them in the second map.
//
// Example with StringValReplacer:
//
//   in := make(map[string]string)
//   out := make(map[string]string)
//   in["foo"] = "bar ${bar}"
//   in["bar"] = "bars"
//   repl := &StringValReplacer{"${", "}", true}
//   repl.Replace(in, out)
//   fmt.Println(out["foo"])   //prints: bar bars
type ValReplacer interface {
	Replace(map[string]string, map[string]string) error
}

// Config hold the information which is needed to receive the
// json data from the spring config server and parse and transform them correctly.
type Config struct {
	Host, Port, Name, Profile, Label string
	KeyFormatter                     func(string) string
	ValFormatter                     func(string) string
	ValReplacer                      ValReplacer
	environment                      *environment
	Properties                       map[string]string
}

// NewConfig returns a new Config as pointer value with a default ValReplacer for
// spring cloud config.
func NewConfig(host string, port string, name string, profiles []string, label string,
	keyFormatter func(string) string, valFormatter func(string) string) *Config {
	return &Config{host, port, name, strings.Join(profiles, ","), label,
		keyFormatter,
		valFormatter,
		&StringValReplacer{"${", "}", true},
		nil, nil}
}

// Fetch fetches the json data from the spring config server, see:
// https://cloud.spring.io/spring-cloud-config/single/spring-cloud-config.html#_quick_start
func (cfg *Config) Fetch(verbose bool) error {
	env := &environment{}
	url := fmt.Sprintf("http://%s:%s/%s/%s/%s", cfg.Host, cfg.Port, cfg.Name, cfg.Profile, cfg.Label)
	fmt.Fprintln(os.Stderr, "Fetching config from server at:", url)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(env)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Located environment: name=%#v, profiles=%v, label=%#v, version=%#v, state=%#v\n",
		env.Name, env.Profiles, env.Label, env.Version, env.State)

	if verbose {
		jsonStr, _ := json.MarshalIndent(env, "", "    ")
		fmt.Println(string(jsonStr))
	}

	cfg.environment = env
	return nil
}

// FetchFile download a file from the spring config server, see:
// https://cloud.spring.io/spring-cloud-config/single/spring-cloud-config.html#_serving_plain_text
func (cfg *Config) FetchFile(filename string, print bool) error {
	url := fmt.Sprintf("http://%s:%s/%s/%s/%s/%s", cfg.Host, cfg.Port, cfg.Name, cfg.Profile, cfg.Label, filename)
	fmt.Fprintf(os.Stderr, "Fetching file \"%s\" from server at: %s\n", filename, url)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if print {
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

// Process use given ValReplacer and formatter functions to process
// the fetched json data and must be called after Fetch
func (cfg *Config) Process(verbose bool) error {
	env := cfg.environment
	if env != nil && env.PropertySources != nil {
		//merge propertySources into one map
		mergedProperties := mergeProps(env.PropertySources)

		if cfg.ValReplacer != nil {

			//replace variables
			replacedProperties := make(map[string]string)
			if err := cfg.ValReplacer.Replace(mergedProperties, replacedProperties); err != nil {
				return err
			}
			cfg.Properties = replacedProperties

			//format keys & values
			if cfg.KeyFormatter != nil && cfg.ValFormatter != nil {
				formattedProps := make(map[string]string)
				for key, val := range replacedProperties {
					nKey := cfg.KeyFormatter(key)
					nVal := cfg.ValFormatter(val)
					formattedProps[nKey] = nVal
					if verbose {
						fmt.Println(nKey, "=", nVal)
					}
				}
				cfg.Properties = formattedProps
			}
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

// StringValReplacer needs the opening and closing string
// for detecting a variables that must be replaced.
// Optionally it can not fail on unknown variables which have no appropriate
// key in the map.
type StringValReplacer struct {
	Opener      string
	Closer      string
	FailOnError bool
}

// Replace replaces all variables with the defined opening and
// closing strings with the value of the key.
func (rpl *StringValReplacer) Replace(in map[string]string, out map[string]string) error {
	var err error
	for key, val := range in {
		var nVal string
		nVal, err = rpl.replStrVar(val, in)
		if err != nil && rpl.FailOnError {
			return err
		}
		out[key] = nVal
	}
	return nil
}

func (rpl *StringValReplacer) replStrVar(str string, m map[string]string) (string, error) {
	var f, s int
	f = strings.Index(str, rpl.Opener) + len(rpl.Opener)
	for f-len(rpl.Opener) > -1 {
		s = f + strings.Index(str[f:], rpl.Closer)
		key := str[f:s]
		if val, ok := m[key]; ok {
			str = str[:f-len(rpl.Opener)] + val + str[s+len(rpl.Closer):]
		} else {
			return str, fmt.Errorf("value for property ${%s} can't be found", key)
		}
		f = strings.Index(str, rpl.Opener) + len(rpl.Opener)
	}
	return str, nil
}
