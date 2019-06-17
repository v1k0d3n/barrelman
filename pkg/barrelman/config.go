package barrelman

import (
	"fmt"
	"github.com/charter-oss/barrelman/pkg/cluster"
	"github.com/charter-oss/barrelman/pkg/manifest"
	"github.com/charter-oss/barrelman/pkg/version"
	"github.com/charter-oss/structured/log"
	"io"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/spf13/viper"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	helm_env "k8s.io/helm/pkg/helm/environment"

	"github.com/charter-oss/barrelman/pkg/manifest/chartsync"
	"github.com/charter-oss/structured/errors"
)

const (
	Token = iota
)

type valueFiles []string //from helm for template command

var (
	settings helm_env.EnvSettings
)

type Config struct {
	Account chartsync.AccountTable
}

type BarrelmanConfig struct {
	FilePath string
	Viper    *viper.Viper
	Env      map[string]string
}

type ConfigCmd struct {
	Options    *CmdOptions
	Config     *Config
	LogOptions *[]string
}

func GetEmptyConfig() *Config {
	return &Config{
		Account: make(map[string]*chartsync.Account),
	}
}

func GetConfigFromFile(s string) (*Config, error) {
	config := &Config{}
	config.Account = make(map[string]*chartsync.Account)

	if _, err := os.Stat(s); os.IsNotExist(err) {
		return config, nil
	}
	f, err := os.Open(s)
	if err != nil {
		return nil, err
	}
	defer func() {
		f.Close()
	}()

	b, err := toBarrelmanConfig(s, f)
	if err != nil {
		return nil, err
	}

	return config.LoadAcc(b)
}

//LoadAcc populates *config.Account from *BarrelmanConfig
func (config *Config) LoadAcc(b *BarrelmanConfig) (*Config, error) {

	if config.Account == nil {
		config.Account = make(map[string]*chartsync.Account)
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
					}).New("Failed to parse accounts in config")
				}
				config.Account[kk.(string)] = acc
			}
		}
	default:
		return nil, errors.WithFields(errors.Fields{"File": b.FilePath}).New("failed to parse accounts in config")
	}
	return config, nil
}

/*
func (cmd *ConfigCmd) Run(KubeConfig string) error {

	ver := version.Get()
	log.WithFields(log.Fields{
		"Version": ver.Version,
		"Commit":  ver.Commit,
		"Branch":  ver.Branch,
	}).Info("Barrelman")

	config, err := GetConfigFromFile(KubeConfig)
	if err != nil {
		return fmt.Errorf("ERROR %v", err)
	}
	fmt.Print("Config is: ", config)
	return nil
}
*/

func (cmd *ConfigCmd) Run(session cluster.Sessioner) error {
	var err error

	ver := version.Get()
	log.WithFields(log.Fields{
		"Version": ver.Version,
		"Commit":  ver.Commit,
		"Branch":  ver.Branch,
	}).Info("Barrelman")

	cmd.Config, err = GetConfigFromFile(cmd.Options.ConfigFile)
	if err != nil {
		return errors.Wrap(err, "got error while loading config")
	}

	if err := ensureWorkDir(cmd.Options.DataDir); err != nil {
		return errors.Wrap(err, "failed to create working directory")
	}

	if err = session.Init(); err != nil {
		return errors.Wrap(err, "failed to create new cluster session")
	}

	if session.GetKubeConfig() != "" {
		log.WithFields(log.Fields{
			"file": session.GetKubeConfig(),
		}).Info("Using kube config")
	}
	if session.GetKubeContext() != "" {
		log.WithFields(log.Fields{
			"file": session.GetKubeContext(),
		}).Info("Using kube context")
	}

	// Open and initialize the manifest
	mfest, err := manifest.New(&manifest.Config{
		DataDir:      cmd.Options.DataDir,
		ManifestFile: cmd.Options.ManifestFile,
		AccountTable: cmd.Config.Account,
	})
	if err != nil {
		return errors.Wrap(err, "error while initializing manifest")
	}

	if !cmd.Options.NoSync {
		if err := mfest.Sync(); err != nil {
			return errors.Wrap(err, "error while downloading charts")
		}
	}

	if _, err := getConfig("test"); err != nil {
		return errors.Wrap(err, "failed to get config")
	}
	return nil
}
func toBarrelmanConfig(s string, r io.Reader) (*BarrelmanConfig, error) {
	barrelConfig := &BarrelmanConfig{FilePath: s}

	barrelConfig.Viper = viper.New()
	barrelConfig.Viper.SetConfigType("yaml")
	if err := barrelConfig.Viper.ReadConfig(r); err != nil {
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
	case int, int8, int16, int32, int64:
		return strconv.Itoa(v.(int))
	case float32, float64:
		return fmt.Sprintf("%v", v)
	default:
		panic(fmt.Sprintf("unhandled type in toString(): %T\n", v))
	}
}

func ensureWorkDir(datadir string) error {
	return os.MkdirAll(datadir, os.ModePerm)
}

func userHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

func getConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func (v *valueFiles) String() string {
	return fmt.Sprint(*v)
}

func (v *valueFiles) Type() string {
	return "valueFiles"
}

func (v *valueFiles) Set(value string) error {
	for _, filePath := range strings.Split(value, ",") {
		*v = append(*v, filePath)
	}
	return nil
}
