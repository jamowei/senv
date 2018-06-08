package senv

import (
	"encoding/json"
	"strconv"
)

type environment struct {
	Name            string           `json:"Name"`
	Profiles        []string         `json:"Profiles"`
	Label           string           `json:"Label"`
	Version         string           `json:"Version"`
	State           string           `json:"State"`
	PropertySources []propertySource `json:"PropertySources"`
}

type propertySource struct {
	Name   string `json:"Name"`
	Source source `json:"Source"`
}

type source struct {
	content map[string]string
}

func (src *source) UnmarshalJSON(b []byte) error {
	src.content = make(map[string]string)
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	for k, v := range m {
		switch val := v.(type) {
		case string:
			src.content[k] = val
		case bool:
			src.content[k] = strconv.FormatBool(val)
		case float64:
			src.content[k] = strconv.FormatFloat(float64(val), 'f', -1, 64)
		}
	}
	return nil
}

func (src source) MarshalJSON() ([]byte, error) {
	return json.Marshal(src.content)
}
