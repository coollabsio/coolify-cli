package config

import (
	"github.com/spf13/viper"
)

// Transitional functions while config is refactored to this file. Relies in setup
// cmd/root.go:initConfig() which would eventually be moved here.
func GetToken() string {
	// Assumes config has been initialized via cmd/root.go. This will be refactored
	instancesMap := viper.Get("instances").([]interface{})
	for _, instance := range instancesMap {
		instanceMap := instance.(map[string]interface{})
		if instanceMap["default"] == true {
			return instanceMap["token"].(string)
		}
	}
	return ""
}

func GetBaseUrl() string {
	// Assumes config has been initialized via cmd/root.go. This will be refactored
	instancesMap := viper.Get("instances").([]interface{})
	for _, instance := range instancesMap {
		instanceMap := instance.(map[string]interface{})
		if instanceMap["default"] == true {
			return instanceMap["fqdn"].(string)
		}
	}
	return ""
}
