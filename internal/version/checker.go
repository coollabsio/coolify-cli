package version

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"time"

	compareVersion "github.com/hashicorp/go-version"
	"github.com/spf13/viper"
)

// Version variables injected by GoReleaser at build time via ldflags
var (
	version = "v1.0.3"
)

func GetVersion() string {
	return version
}

// CheckInterval for version checking
const CheckInterval = 10 * time.Minute

// Tag represents a git tag for version checking
type Tag struct {
	Ref string `json:"ref"`
}

// CheckLatestVersionOfCli checks for CLI updates
func CheckLatestVersionOfCli(debug bool) (string, error) {
	lastCheck := viper.GetString("lastupdatechecktime")
	if lastCheck != "" {
		lastCheckTime, err := time.Parse(time.RFC3339, lastCheck)
		if err == nil && lastCheckTime.Add(CheckInterval).After(time.Now()) {
			if debug {
				log.Println("Skipping update check. Last check was less than 10 minutes ago.")
			}
			return GetVersion(), nil
		}
	}

	// Update check time
	viper.Set("lastupdatechecktime", time.Now().Format(time.RFC3339))
	viper.WriteConfig()

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
	latestVersion := versions[len(versions)-1]

	// Compare versions properly using semantic versioning
	currentVersion, err := compareVersion.NewVersion(GetVersion())
	if err != nil {
		return latestVersion.String(), err
	}

	if latestVersion.GreaterThan(currentVersion) {
		fmt.Printf("There is a new version of Coolify CLI available.\nPlease update with 'coolify update'.\n\n")
	}
	return latestVersion.String(), nil
}
