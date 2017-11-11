package senv

import (
	"strings"
	"fmt"
	"strconv"
)

// flattens a map of type map[string]interface{}
type MapFlattener interface {
	Flatten(map[string]interface{}) (map[string]interface{}, error)
}

type StringMapFlattener struct {}

// Flatten takes a map and returns a new one where nested maps are replaced
// by dot-delimited keys.
func (jfl *StringMapFlattener) Flatten(m map[string]interface{}) (map[string]interface{}, error) {
	o := make(map[string]interface{})
	for k, v := range m {
		switch child := v.(type) {
		case map[string]interface{}:
			nm, _ := jfl.Flatten(child)
			for nk, nv := range nm {
				o[k+"."+nk] = nv
			}
		default:
			o[k] = v
		}
	}
	return o, nil
}

type KeyFormatter interface {
	Format(string) (string, error)
}

type EnvKeyFormatter struct {
	sep string
	upperCase bool
}

func (ekf *EnvKeyFormatter) Format(in string) (string, error) {
	out := strings.Replace(in, ".", ekf.sep, -1)
	if ekf.upperCase {
		out = strings.ToUpper(out)
	}
	return out, nil
}


type MapReplacer interface {
	Replace(map[string]interface{}) (map[string]string, error)
}

type FlatMapReplacer struct {
	Opener      string
	Closer      string
	FailOnError bool
	KeyFormatter KeyFormatter
}

func (rpl *FlatMapReplacer) Test(m map[string]interface{}) bool {
	for _, v := range m {
		switch v.(type) {
		case map[string]interface{}:
			return false
		}
	}
	return true
}

func (rpl *FlatMapReplacer) Replace(m map[string]interface{}) (map[string]string, error) {
	o := make(map[string]string)
	var err error
	for key, t := range m {
		var nVal string
		switch val := t.(type) {
		case string:
			nVal, err = rpl.replStrVar(nVal, m)
			if err != nil && rpl.FailOnError {
				return nil, err
			}
		default:
			nVal, err = rpl.conv2Str(val)
			if err != nil && rpl.FailOnError {
				return nil, err
			}
		}
		nKey, err := rpl.KeyFormatter.Format(key)
		if err != nil && rpl.FailOnError {
			return nil, err
		}
		o[nKey] = nVal
	}
	return o, nil
}

func (rpl *FlatMapReplacer) conv2Str(i interface{}) (string, error) {
	var res string
	switch val := i.(type) {
	case string:
		res = val
	case int:
		res = strconv.Itoa(val)
	case uint:
		res = strconv.FormatUint(uint64(val), 10)
	case bool:
		res = strconv.FormatBool(val)
	case float32:
		res = strconv.FormatFloat(float64(val), 'f', -1, 32)
	case float64:
		res = strconv.FormatFloat(float64(val), 'f', -1, 64)
	default:
		return "", fmt.Errorf("can't convert %+v to string", val)
	}
	return res, nil
}

func (rpl *FlatMapReplacer) replStrVar(str string, m map[string]interface{}) (string, error) {
	var f, s int
	f = strings.Index(str, rpl.Opener) + len(rpl.Opener)
	for f - len(rpl.Opener) > -1 {
		s = f + strings.Index(str[f:], rpl.Closer)
		key := str[f:s]
		if i, ok := m[key]; ok {
			val, err := rpl.conv2Str(i)
			if err != nil {
				return str, err
			}
			str = str[:f-len(rpl.Opener)] + val + str[s+len(rpl.Closer):]
		} else {
			return str, fmt.Errorf("value for property ${%s} can't be found", key)
		}
		f = strings.Index(str, rpl.Opener) + len(rpl.Opener)
	}
	return str, nil
}