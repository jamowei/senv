package main

import (
	"strings"
)

type EnvKeyFormatter struct {
	sep       string
	upperCase bool
}

func (ekf *EnvKeyFormatter) Format(in string) (string, error) {
	out := strings.Replace(in, ".", ekf.sep, -1)
	if ekf.upperCase {
		out = strings.ToUpper(out)
	}
	return out, nil
}

type EnvValFormatter struct {}

func (evf *EnvValFormatter) Format(s string) (res string, err error) {
	res = strings.Replace(s, "\r\n", "", -1)
	res = strings.Replace(res, "\n", "", -1)
	return
}
