package util

import "github.com/charter-oss/structured/log"

func LogSettings(args *[]string) []func(*log.Logger) error {
	ret := []func(*log.Logger) error{}
	for _, v := range *args {
		switch v {
		case "debug", "info", "warn", "error":
			ret = append(ret, log.OptSetLevel(v))
		case "JSON":
			ret = append(ret, log.OptSetJSON())
		}
	}
	return ret
}
