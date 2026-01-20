//go:generate goversioninfo

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type launcherCommune struct {
	Patchline       string         `json:"last_patchline"`
	Username        string         `json:"last_username"`
	SelectedVersion int            `json:"last_version"`
	LatestVersions  map[string]int `json:"last_version_scan_result"`
	GameFolder      string         `json:"install_directory"`
	Mode            string         `json:"mode"`
}

var (
	wCommune = launcherCommune{
		Patchline: "release",
		Username:  "TransRights",
		LatestVersions: map[string]int{
			"release":     4,
			"pre-release": 8,
		},
		SelectedVersion: 4,
		GameFolder:      DefaultGameFolder(),
		Mode:            "fakeonline",
	}
	wProgress = 0
	wDisabled = false
)

func checkForUpdates() {
	lastRelease := wCommune.LatestVersions["release"]
	lastPreRelease := wCommune.LatestVersions["pre-release"]

	latestRelease := findLatestVersionNoAuth(lastRelease, runtime.GOARCH, runtime.GOOS, "release")
	latestPreRelease := findLatestVersionNoAuth(lastPreRelease, runtime.GOARCH, runtime.GOOS, "pre-release")

	fmt.Printf("latestRelease: %d\n", latestRelease)
	fmt.Printf("latestPreRelease: %d\n", latestPreRelease)

	if latestRelease > lastRelease {
		fmt.Printf("Found new release version: %d", latestRelease)
		wCommune.LatestVersions["release"] = latestRelease
	}

	if latestPreRelease > lastPreRelease {
		fmt.Printf("Found new pre-release version: %d", latestPreRelease)
		wCommune.LatestVersions["pre-release"] = latestPreRelease
	}

	writeSettings()
}

func writeSettings() {
	fmt.Printf("Saving settings ...\n")
	jlauncher, _ := json.Marshal(wCommune)

	err := os.MkdirAll(filepath.Dir(getLauncherJson()), 0666)
	if err != nil {
		fmt.Printf("error writing settings: %s\n", err)
		return
	}

	err = os.WriteFile(getLauncherJson(), jlauncher, 0666)
	if err != nil {
		fmt.Printf("error writing settings: %s\n", err)
		return
	}
}

func getDefaultSettings() {
	writeSettings()
	go checkForUpdates()
}

func getLauncherJson() string {
	return filepath.Join(LauncherFolder(), "launcher.json")
}

func readSettings() {
	_, err := os.Stat(getLauncherJson())
	if err != nil {
		getDefaultSettings()
	} else {
		data, err := os.ReadFile(getLauncherJson())
		if err != nil {
			getDefaultSettings()
			return
		}
		json.Unmarshal(data, &wCommune)

		if wCommune.GameFolder != GameFolder() {
			wCommune.GameFolder = GameFolder()
		}

		fmt.Printf("Reading last settings: \n")
		fmt.Printf("username: %s\n", wCommune.Username)
		fmt.Printf("patchline: %s\n", wCommune.Patchline)
		fmt.Printf("last used version: %d\n", wCommune.SelectedVersion)
		fmt.Printf("newest known release: %d\n", wCommune.LatestVersions["release"])
		fmt.Printf("newest known pre-release: %d\n", wCommune.LatestVersions["pre-release"])

		// check for updates in the background
		go checkForUpdates()
	}
}

func valToChannel(vchl int) string {
	switch vchl {
	case 0:
		return "release"
	case 1:
		return "pre-release"
	default:
		return "release"
	}
}

func channelToVal(channel string) int {
	switch channel {
	case "release":
		return 0
	case "pre-release":
		return 1
	default:
		return 0
	}
}

func updateProgress(done int64, total int64) {
	lastProgress := wProgress
	newProgress := int((float64(done) / float64(total)) * 100.0)

	if newProgress != lastProgress {
		wProgress = newProgress
	}
}

func main() {
	os.MkdirAll(MainFolder(), 0775)
	os.MkdirAll(LauncherFolder(), 0775)
	os.MkdirAll(UserDataFolder(), 0775)
	os.MkdirAll(JreFolder(), 0775)
	os.MkdirAll(ServerDataFolder(), 0775)
	readSettings()
	os.MkdirAll(GameFolder(), 0775)

	err := runGioUI()
	if err != nil {
		fmt.Println("Error: ", err)
	}
}
