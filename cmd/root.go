package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/adrg/xdg"
	compareVersion "github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var CliVersion = "0.0.1"
var LastUpdateCheckTime time.Time
var CheckInverval = 10 * time.Minute

var ConfigDir = xdg.ConfigHome

var Version string
var Name string
var Fqdn string
var Token string
var Instance http.Client
var SensitiveInformationOverlay = "********"

// Flags
var Debug bool
var ShowSensitive bool
var Force bool
var JsonMode bool
var PrettyMode bool
var SetDefaultInstance bool

var w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.Debug)

type Tag struct {
	Ref string `json:"ref"`
}

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
		log.Println(err)
		os.Exit(0)
	}
	currentVersion, err := compareVersion.NewVersion(Version)
	if err != nil {
		log.Println(err)
		os.Exit(0)
	}
	if currentVersion.LessThan(requiredVersion) {
		log.Printf("Minimum required Coolify API version is: %s\n", version)
		log.Print("Please upgrade your Coolify instance for this command.\n")
		os.Exit(0)
	}
}
func FetchVersion() (string, error) {
	data, err := Fetch("version")
	if err != nil {
		log.Println(err)
		os.Exit(0)
		return "", err
	}
	Version = data
	return data, nil
}
func Fetch(url string) (string, error) {
	url = Fqdn + "/api/v1/" + url
	if Debug {
		log.Println("Fetching data from", url)
	}
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

func CheckLatestVersionOfCli() (string, error) {
	getLastUpdateCheckTime()
	if LastUpdateCheckTime.Add(CheckInverval).After(time.Now()) {
		if Debug {
			log.Println("Skipping update check. Last check was less than 10 minutes ago.")
		}
		return CliVersion, nil
	}
	setLastUpdateCheckTime()
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

	versions := make([]*compareVersion.Version, len(versionsRaw))
	for i, raw := range versionsRaw {
		v, err := compareVersion.NewVersion(raw)
		if err != nil {
			return "", err
		}
		versions[i] = v
	}

	sort.Sort(compareVersion.Collection(versions))
	latestVersion := versions[len(versions)-1].String()
	if latestVersion != CliVersion {
		fmt.Printf("There is a new version of Coolify CLI available.\nPlease update with 'coolify --update'.\n\n")
	}
	return latestVersion, nil

}
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(0)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&Token, "token", "", "", "Token for authentication (https://app.coolify.io/security/api-tokens)")
	rootCmd.PersistentFlags().StringVarP(&Fqdn, "host", "", "", "Coolify instance hostname")

	rootCmd.PersistentFlags().BoolVarP(&JsonMode, "json", "", false, "Json mode")
	rootCmd.PersistentFlags().BoolVarP(&PrettyMode, "pretty", "", false, "Pretty json mode")
	rootCmd.PersistentFlags().BoolVarP(&ShowSensitive, "show-sensitive", "s", false, "Show sensitive information")
	rootCmd.PersistentFlags().BoolVarP(&Force, "force", "f", false, "Force")
	rootCmd.PersistentFlags().BoolVarP(&Debug, "debug", "", false, "Debug mode")
}
func setLastUpdateCheckTime() {
	timeNow := time.Now()
	viper.Set("lastupdatechecktime", timeNow)
	viper.WriteConfig()
	LastUpdateCheckTime = timeNow
}
func getLastUpdateCheckTime() {
	lastUpdateCheckTimeString := viper.Get("lastupdatechecktime").(string)
	lastUpdateCheckTime, err := time.Parse(time.RFC3339, lastUpdateCheckTimeString)
	if err != nil {
		log.Fatalf("Error parsing time: %v", err)
	}
	LastUpdateCheckTime = lastUpdateCheckTime

}
func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(ConfigDir + "/coolify")
	if _, err := os.Stat(ConfigDir + "/coolify"); os.IsNotExist(err) {
		os.MkdirAll(ConfigDir+"/coolify", 0755)
	}
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("Config file not found. Creating a new one at", ConfigDir+"/coolify/config.json")
			viper.Set("lastUpdateCheckTime", time.Now())
			viper.Set("instances", []interface{}{map[string]interface{}{
				"name":    "cloud",
				"default": true,
				"fqdn":    "https://app.coolify.io",
				"token":   "",
			},
			})
			viper.Set("instances", append(viper.Get("instances").([]interface{}), map[string]interface{}{
				"name":  "localhost",
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

	if Debug {
		log.Println("Using config file:", viper.ConfigFileUsed())
	}
	instancesMap := viper.Get("instances").([]interface{})
	for _, instance := range instancesMap {
		instanceMap := instance.(map[string]interface{})
		if instanceMap["default"] == true {
			if Fqdn == "" {
				Fqdn = instanceMap["fqdn"].(string)
			}
			if Token == "" {
				Token = instanceMap["token"].(string)
			}
		}
	}
	data, err := CheckLatestVersionOfCli()
	if err != nil {
		log.Println(err)
	}
	if data != CliVersion {
		log.Printf("New version of Coolify CLI is available: %s\n", data)
	}
}
