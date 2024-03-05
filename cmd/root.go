package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/hashicorp/go-version"
	compareVersion "github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var CliVersion = "0.0.0"
var Version string
var Name string
var Fqdn string
var Token string
var Instance http.Client
var SensitiveInformationOverlay = "********"

// Flags
var ShowSensitive bool
var Force bool
var JsonMode bool
var PrettyMode bool
var SetDefaultInstance bool

var w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.Debug)

var rootCmd = &cobra.Command{
	Use:   "coolify",
	Short: "Coolify CLI",
	Long:  `A CLI tool to interact with Coolify API.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func CheckMinimumVersion(version string) {
	FetchVersion()
	requiredVersion, err := compareVersion.NewVersion(version)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	currentVersion, err := compareVersion.NewVersion(Version)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if currentVersion.LessThan(requiredVersion) {
		fmt.Printf("Minimum required Coolify API version is: %s\n", version)
		fmt.Print("Please upgrade your Coolify instance for this command.\n\n")
		os.Exit(1)
	}
}
func FetchVersion() (string, error) {
	data, err := Fetch("version")
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	Version = data
	return data, nil
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

	return string(body), nil
}

type Tag struct {
	Ref string `json:"ref"`
}

func CheckLatestVersionOfCli() (string, error) {
	url := "https://api.github.com/repos/coollabsio/coolify-cli/git/refs/tags"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
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

	var tags []Tag
	if err := json.Unmarshal(body, &tags); err != nil {
		return "", err
	}

	versionsRaw := make([]string, 0, len(tags))
	for _, tag := range tags {
		versionStr := tag.Ref[10:]
		versionsRaw = append(versionsRaw, versionStr)
	}

	versions := make([]*version.Version, len(versionsRaw))
	for i, raw := range versionsRaw {
		v, err := version.NewVersion(raw)
		if err != nil {
			return "", err
		}
		versions[i] = v
	}

	sort.Sort(version.Collection(versions))
	latestVersion := versions[len(versions)-1].String()
	if latestVersion != CliVersion {
		fmt.Printf("There is a new version of Coolify CLI available.\nPlease update with 'coolify --update'.\n\n")
	}
	return latestVersion, nil

}
func Execute() {
	_, err := CheckLatestVersionOfCli()
	if err != nil {
		fmt.Println(err)
	}

	err = rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&Token, "token", "", "", "Token for authentication (https://app.coolify.io/security/api-tokens)")
	rootCmd.PersistentFlags().StringVarP(&Fqdn, "host", "", "https://app.coolify.io", "Coolify instance hostname")

	rootCmd.PersistentFlags().BoolVarP(&JsonMode, "json", "", false, "Json mode")
	rootCmd.PersistentFlags().BoolVarP(&PrettyMode, "pretty", "", false, "Make json output pretty")
	rootCmd.PersistentFlags().BoolVarP(&ShowSensitive, "show-sensitive", "s", false, "Show sensitive information")
	rootCmd.PersistentFlags().BoolVarP(&Force, "force", "f", false, "Force")
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
