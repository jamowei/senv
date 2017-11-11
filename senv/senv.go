package senv

import (
	"fmt"
	"net/http"
	"encoding/json"
	"log"
)

type Config struct {
	Host, Port, Name, Profile, Label string
	Flattener                        MapFlattener
	Replacer                         MapReplacer
	Env                              *Enviroment
}

type Enviroment struct {
	Name     string   `json:"name"`
	Profiles []string `json:"profiles"`
	Label    string   `json:"label"`
	Version  string   `json:"version"`
	State    string   `json:"state"`
	Sources  []Source `json:"propertySources"`
}

type Source struct {
	Name    string                 `json:"name"`
	Content map[string]interface{} `json:"source"`
}

func NewConfig(host string, port string, name string, profile string, label string) *Config {
	return &Config{host, port, name, profile, label,
		&StringMapFlattener{},
		&FlatMapReplacer{"${", "}", true, &EnvKeyFormatter{"_", true}},
		nil}
}

func (cfg *Config) Recieve() error {
	env := new(Enviroment)
	url := fmt.Sprintf("http://%s:%s/%s/%s/%s", cfg.Host, cfg.Port, cfg.Name, cfg.Profile, cfg.Label)
	log.Println("Fetching cfg from server at: ", url)
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
		log.Printf("Located environment: name=%s, profiles=%s, label=%s, version=%s, state=%s\n",
			env.Name, env.Profiles, env.Label, env.Version, env.State)
		cfg.Env = env
	}
	return nil
}

func (cfg *Config) Process() error {
	data := cfg.Env
	if data.Sources != nil {
		for _, source := range data.Sources {
			var rawProps map[string]interface{}
			var strProps map[string]string
			var err error
			rawProps, err = cfg.Flattener.Flatten(source.Content)
			if err != nil {
				return err
			}
			strProps, err = cfg.Replacer.Replace(rawProps)
			if err != nil {
				return err
			}
			fmt.Print(strProps)
		}
	}
	return nil
}
