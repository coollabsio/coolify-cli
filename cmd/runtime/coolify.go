package runtime

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/adrg/xdg"
	"github.com/coollabsio/cli-coolify/cmd/coolTypes"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Package runtime provides a reuseable struct that holds configuration, http client and other common functions shared by all the commands.

var (
	// Version will be injected during build by goreleaser, without the 'v' prefix
	Version                       = "0.0.0-dev"
	DefaultConfigDirectory string = xdg.ConfigHome // Currently using xdg.ConfigHome but maybe we can expose this as a flag in future.
)

type Getter func() *Coolify

type Config struct {
	Directory  string
	FQDN       string
	Token      string
	JsonExists bool
	Timeout    time.Duration
	Insecure   bool
}

type Coolify struct {
	Version string
	Config  Config
	Client  *http.Client
	Logger  *logrus.Logger
}

func NewCoolify(fqdn, token string, logLevel string) *Coolify {
	// Create HTTP client with default settings
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
		},
	}

	// Initialize logger with default settings
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
	})

	// Create the Coolify instance
	coolify := &Coolify{
		Version: Version,
		Config: Config{
			Directory:  DefaultConfigDirectory,
			FQDN:       fqdn,
			Token:      token,
			JsonExists: false,
			Timeout:    30 * time.Second,
			Insecure:   false,
		},
		Client: client,
		Logger: logger,
	}

	// Set the log level immediately
	if logLevel != "" {
		coolify.SetLogLevel(logLevel)
	}

	return coolify
}

// ConfigureHTTPClient updates the HTTP client with the current configuration
func (c *Coolify) ConfigureHTTPClient() {
	// Create transport with TLS configuration
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: c.Config.Insecure,
		},
	}

	// Create new client with configured transport and timeout
	c.Client = &http.Client{
		Transport: transport,
		Timeout:   c.Config.Timeout,
	}
}

// SetTimeout sets the timeout for HTTP requests
func (c *Coolify) SetTimeout(timeout time.Duration) {
	// Only log if the timeout is different from the default
	if timeout != 30*time.Second {
		c.LogDebug("Setting HTTP client timeout to: %s", timeout)
	}
	c.Config.Timeout = timeout
	c.ConfigureHTTPClient()
}

// SetInsecure sets whether to skip TLS verification
func (c *Coolify) SetInsecure(insecure bool) {
	// Only log if insecure is true (different from default false)
	if insecure {
		c.LogDebug("Setting TLS verification skip to: %t", insecure)
	}
	c.Config.Insecure = insecure
	c.ConfigureHTTPClient()
}

// GetFormattedVersion returns the version with 'v' prefix for display
func (c *Coolify) GetFormattedVersion() string {
	// Tags on GitHub don't have 'v' prefix, but we want to display it
	return fmt.Sprintf("v%s", c.Version)
}

// Load reads the configuration file from the default directory and loads it into the Coolify struct.
func (c *Coolify) Load(instanceName string) error {
	baseDir := path.Join(c.Config.Directory, "coolify")
	viper.SetConfigType("json")
	viper.AddConfigPath(baseDir)

	c.LogDebug("Loading configuration from: %s", baseDir)

	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		c.LogDebug("Configuration directory does not exist: %s", baseDir)
		return nil // we return nil here because if the configuration directory doesnt exist, then the config file also doesnt exist.
	}

	if err := viper.ReadInConfig(); err != nil {
		c.LogError("Failed to read configuration file: %v", err)
		return err // we return the error here because if the configuration directory exists, then the config file should also exist and not error.
	}

	c.LogDebug("Configuration file loaded successfully")
	c.Config.JsonExists = true

	if viper.Get("instances") != nil {
		instances := make([]coolTypes.Instance, 0)
		if err := viper.UnmarshalKey("instances", &instances); err != nil {
			c.LogError("Failed to unmarshal instances: %v", err)
			return err
		}

		// if fqdn and token are not set, then we will set them to the default instance or name if provided from flags
		if c.Config.FQDN == "" && c.Config.Token == "" {
			c.LogDebug("FQDN and Token not provided via flags, looking for instance: %s", instanceName)
			for _, instance := range instances {
				if (instanceName != "" && instance.Name == instanceName) || (instance.Default && instanceName == "") {
					c.LogDebug("Using instance: %s with FQDN: %s", instance.Name, instance.Fqdn)
					c.Config.FQDN = instance.Fqdn
					c.Config.Token = instance.Token
					break
				}
			}
		}
	}
	return nil
}

// Save saves the configuration file
func (c *Coolify) Save() error {
	baseDir := path.Join(c.Config.Directory, "coolify")
	c.LogDebug("Saving configuration to: %s", baseDir)

	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		c.LogDebug("Creating configuration directory: %s", baseDir)
		os.MkdirAll(baseDir, 0o755)
	}

	var err error
	if c.Config.JsonExists {
		c.LogDebug("Updating existing configuration file")
		err = viper.WriteConfig()
	} else {
		c.LogDebug("Creating new configuration file")
		err = viper.SafeWriteConfig()
	}

	if err != nil {
		c.LogError("Failed to save configuration: %v", err)
	} else {
		c.LogDebug("Configuration saved successfully")
	}

	return err
}

// Delete removes the configuration directory
func (c *Coolify) Delete() error {
	configPath := path.Join(c.Config.Directory, "coolify")
	c.LogDebug("Deleting configuration directory: %s", configPath)

	err := os.RemoveAll(configPath)
	if err != nil {
		c.LogError("Failed to delete configuration directory: %v", err)
	} else {
		c.LogDebug("Configuration directory deleted successfully")
	}

	return err
}

// SetLogLevel sets the log level for the logger
func (c *Coolify) SetLogLevel(level string) {
	switch level {
	case "trace":
		c.Logger.SetLevel(logrus.TraceLevel)
	case "debug":
		c.Logger.SetLevel(logrus.DebugLevel)
	case "info":
		c.Logger.SetLevel(logrus.InfoLevel)
	case "warn", "warning":
		c.Logger.SetLevel(logrus.WarnLevel)
	case "error":
		c.Logger.SetLevel(logrus.ErrorLevel)
	case "fatal":
		c.Logger.SetLevel(logrus.FatalLevel)
	case "panic":
		c.Logger.SetLevel(logrus.PanicLevel)
	default:
		c.Logger.SetLevel(logrus.InfoLevel)
	}
}

// LogDebug logs a message at debug level
func (c *Coolify) LogDebug(format string, args ...interface{}) {
	c.Logger.Debugf(format, args...)
}

// LogInfo logs a message at info level
func (c *Coolify) LogInfo(format string, args ...interface{}) {
	c.Logger.Infof(format, args...)
}

// LogWarn logs a message at warn level
func (c *Coolify) LogWarn(format string, args ...interface{}) {
	c.Logger.Warnf(format, args...)
}

// LogError logs a message at error level
func (c *Coolify) LogError(format string, args ...interface{}) {
	c.Logger.Errorf(format, args...)
}

// LogTrace logs a message at trace level
func (c *Coolify) LogTrace(format string, args ...interface{}) {
	c.Logger.Tracef(format, args...)
}
