package runtime

import (
	"os"
	"path"

	"github.com/adrg/xdg"
	"github.com/coollabsio/coolify-cli/cmd/coolTypes"
	"github.com/spf13/viper"
)

// Package runtime provides a reuseable struct that holds configuration, http client and other common functions shared by all the commands.

var (
	Version                string = "v0.0.1"       // Hardcoded for now but should be passed during build time
	DefaultConfigDirectory string = xdg.ConfigHome // Currently using xdg.ConfigHome but maybe we can expose this as a flag in future.
)

type Getter func() *Coolify

type Config struct {
	Directory  string
	FQDN       string
	Token      string
	JsonExists bool
}

type Coolify struct {
	Version string
	Config  Config
}

func NewCoolify(fqdn, token string) *Coolify {
	// we need to create the base http here as we dont know if they are using a config file or not
	return &Coolify{
		Version: Version,
		Config: Config{
			Directory:  DefaultConfigDirectory,
			FQDN:       fqdn,
			Token:      token,
			JsonExists: false,
		},
	}
}

// Load reads the configuration file from the default directory and loads it into the Coolify struct.
func (conf *Config) Load(instanceName string) error {
	baseDir := path.Join(conf.Directory, "coolify")
	viper.SetConfigType("json")
	viper.AddConfigPath(baseDir)
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		return nil // we return nil here because if the configuration directory doesnt exist, then the config file also doesnt exist.
	}
	if err := viper.ReadInConfig(); err != nil {
		return err // we return the error here because if the configuration directory exists, then the config file should also exist and not error.
	}
	conf.JsonExists = true
	if viper.Get("instances") != nil {
		instances := make([]coolTypes.Instance, 0)
		if err := viper.UnmarshalKey("instances", &instances); err != nil {
			return err
		}
		// if fqdn and token are not set, then we will set them to the default instance or name if provided from flags
		if conf.FQDN == "" && conf.Token == "" {
			for _, instance := range instances {
				if (instanceName != "" && instance.Name == instanceName) || (instance.Default && instanceName == "") {
					conf.FQDN = instance.Fqdn
					conf.Token = instance.Token
					break
				}
			}
		}
	}
	return nil
}

func (conf *Config) Save() error {
	baseDir := path.Join(conf.Directory, "coolify")
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		os.MkdirAll(baseDir, 0o755)
	}
	if conf.JsonExists {
		return viper.WriteConfig()
	}
	return viper.SafeWriteConfig()
}

func (conf *Config) Delete() error {
	return os.RemoveAll(path.Join(conf.Directory, "coolify"))
}
