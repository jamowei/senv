package senv

import (
	"fmt"
	"net/http"
	"encoding/json"
	"os"
	"strings"
)

type Formatter interface {
	Format(string) (string, error)
}

type ValReplacer interface {
	Replace(map[string]string, map[string]string) (error)
}

type Config struct {
	Host, Port, Name, Profile, Label string
	KeyFormatter Formatter
	ValFormatter Formatter
	ValReplacer ValReplacer
	enviroment  *enviroment
	Properties  map[string]string
}

func NewConfig(host string, port string, name string, profile string, label string,
	keyFormatter Formatter, valFormatter Formatter) *Config {
	return &Config{host, port, name, profile, label,
		keyFormatter,
		valFormatter,
		&StringValReplacer{"${", "}", true},
		nil, nil}
}

func (cfg *Config) Fetch() error {
	env := new(enviroment)
	url := fmt.Sprintf("http://%s:%s/%s/%s/%s", cfg.Host, cfg.Port, cfg.Name, cfg.Profile, cfg.Label)
	fmt.Fprintln(os.Stderr,"Fetching config from server at:", url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	if resp.Body != nil {
		defer resp.Body.Close()
		err := json.NewDecoder(resp.Body).Decode(env)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr,"Located environment: name=%#v, profiles=%v, label=%#v, version=%#v, state=%#v\n",
			env.Name, env.Profiles, env.Label, env.Version, env.State)
		cfg.enviroment = env
	}
	return nil
}

func (cfg *Config) Process() error {
	env := cfg.enviroment
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
					nKey, err := cfg.KeyFormatter.Format(key)
					if err != nil {
						return err
					}
					nVal, err := cfg.ValFormatter.Format(val)
					if err != nil {
						return err
					}
					formattedProps[nKey] = nVal
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
	for i := len(pSources)-1; i >= 0; i-- {
		data := pSources[i]
		// merge all propertySources to one map
		for key, val := range data.Source.content {
			merged[key] = val
		}
	}
	return
}

type StringValReplacer struct {
	Opener       string
	Closer       string
	FailOnError  bool
}

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
