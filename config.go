package main

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/charter-se/barrelman/chartsync"
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
}

func GetConfig(s string) (*Config, error) {
	config := &Config{}
	config.Account = make(map[string]*chartsync.Account)
	b, err := loadConfig(s)
	if err != nil {
		return nil, err
	}
	keys := b.Viper.AllKeys()
	for k, v := range keys {
		fmt.Printf("Key: %k, val: %v\n", k, v)
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
								acc.User = iv.(string)
							case "secret":
								acc.Secret = iv.(string)
							case "type":
								acc.Typ = iv.(string)
							default:
								return nil, fmt.Errorf("Unknown field in account: %v", ik.(string))
							}
						}
					}
				default:
					return nil, fmt.Errorf("Failed to parse accounts in config file: %v, got type %T", b.FilePath, vv)
				}
				config.Account[kk.(string)] = acc
			}
		}
	default:
		return nil, fmt.Errorf("Failed to parse accounts in config file: %v", b.FilePath)
	}
	return config, nil
}

func loadConfig(s string) (*BarrelmanConfig, error) {
	barrelConfig := &BarrelmanConfig{FilePath: s}
	data, err := ioutil.ReadFile(s)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("could not read config file '%v': %v", s, err))
	}

	barrelConfig.Viper = viper.New()
	barrelConfig.Viper.SetConfigType("yaml")
	if err := barrelConfig.Viper.ReadConfig(bytes.NewBuffer(data)); err != nil {
		return nil, fmt.Errorf("Failed to read barrelman config file [%v]: %v", barrelConfig.FilePath, err)
	}

	return barrelConfig, nil
}
