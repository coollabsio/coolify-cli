package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Name string
var Fqdn string
var Token string
var Instance http.Client

var rootCmd = &cobra.Command{
	Use:   "coolify-cli",
	Short: "A brief description of your application",
	Long:  ``,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func Fetch(url string) (string, error) {
	url = Fqdn + "/api/v1/" + url
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "Bearer "+Token)
	resp, err := Instance.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("%d - Failed to fetch data from %s. Error: %s", resp.StatusCode, url, string(body))
	}

	if err != nil {
		return "", err
	}

	return string(body), nil
}
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&Token, "token", "", "", "Token for authentication (https://app.coolify.io/security/api-tokens)")
	rootCmd.PersistentFlags().StringVarP(&Fqdn, "host", "", "https://app.coolify.io", "Coolify instance hostname")
}
func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			viper.Set("instances", []interface{}{map[string]interface{}{
				"default": true,
				"fqdn":    "https://app.coolify.io",
				"token":   "",
			},
			})
			viper.Set("instances", append(viper.Get("instances").([]interface{}), map[string]interface{}{
				"fqdn":  "http://localhost:8000",
				"token": "",
			}))
			viper.SafeWriteConfig()
			return
			// Config file not found; ignore error if desired
		} else {
			fmt.Println("Error reading config file, ", err)
			return
			// Config file was found but another error was produced
		}
	}
	instancesMap := viper.Get("instances").([]interface{})
	for _, instance := range instancesMap {
		instanceMap := instance.(map[string]interface{})
		if instanceMap["default"] == true {
			Fqdn = instanceMap["fqdn"].(string)
			if Token == "" {
				Token = instanceMap["token"].(string)
			}
		}
	}
}
