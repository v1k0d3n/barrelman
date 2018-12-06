package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"

	"github.com/charter-se/barrelman/manifest/chartsync"
	"github.com/charter-se/structured/errors"
	"github.com/spf13/viper"
)

const (
	Token = iota
)

type Config struct {
	Account chartsync.AccountTable
}

type BarrelmanConfig struct {
	FilePath string
	Viper    *viper.Viper
	Env      map[string]string
}

func GetConfigFromFile(s string) (*Config, error) {
	config := &Config{}
	config.Account = make(map[string]*chartsync.Account)
	b, err := loadConfig(s)
	if err != nil {
		return nil, err
	}

	account := b.Viper.Get("account")
	// This block supports the YAML format :
	// 	account:
	//   - github.com:
	//		 type: token
	//       user: username
	//       secret: 12345678901011112113114115
	switch account.(type) {
	case []interface{}:
		for _, v := range account.([]interface{}) {
			for kk, vv := range v.(map[interface{}]interface{}) {
				acc := &chartsync.Account{}
				switch vv.(type) {
				case map[interface{}]interface{}:
					for ik, iv := range vv.(map[interface{}]interface{}) {
						switch ik.(type) {
						case string:
							switch ik.(string) {
							case "user":
								acc.User = toString(iv)
							case "secret":
								acc.Secret = toString(iv)
							case "type":
								acc.Typ = toString(iv)
							default:
								return nil, errors.WithFields(errors.Fields{"Field": ik.(string)}).New("unknown field in account")
							}
						}
					}
				default:
					return nil, errors.WithFields(errors.Fields{
						"File":      b.FilePath,
						"ValueType": fmt.Sprintf("%T", vv),
					}).New("Failed to parse accounts in config file")
				}
				config.Account[kk.(string)] = acc
			}
		}
	default:
		return nil, errors.WithFields(errors.Fields{"File": b.FilePath}).New("failed to parse accounts in config file")
	}
	return config, nil
}

func loadConfig(s string) (*BarrelmanConfig, error) {
	barrelConfig := &BarrelmanConfig{FilePath: s}
	data, err := ioutil.ReadFile(s)
	if err != nil {
		return nil, errors.WithFields(errors.Fields{"File": s}).Wrap(err, "could not read config file")
	}

	barrelConfig.Viper = viper.New()
	barrelConfig.Viper.SetConfigType("yaml")
	if err := barrelConfig.Viper.ReadConfig(bytes.NewBuffer(data)); err != nil {
		return nil, errors.WithFields(errors.Fields{"File": barrelConfig.FilePath}).Wrap(err, "failed to read barrelman config file")
	}

	barrelConfig.Env = loadEnv()

	return barrelConfig, nil
}

func loadEnv() map[string]string {
	ret := make(map[string]string)
	env := os.Environ()
	envRx := regexp.MustCompile("^(.+)=(.+)")
	for _, v := range env {
		if iv := envRx.FindStringSubmatch(v); iv != nil {
			if len(iv) > 2 {
				ret[iv[1]] = iv[2]
			}
		}
	}
	return ret
}

func toString(v interface{}) string {
	switch v.(type) {
	case string:
		return v.(string)
	case int:
		return strconv.Itoa(v.(int))
	default:
		panic(fmt.Sprintf("unhandled type in toString(): %T\n", v))
	}
}
