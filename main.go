package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"time"

	"github.com/mitchellh/go-ps"
)

// Config struct for reading in json config file
type Config struct {
	ObsidianPath string `json:"obsidian-path"`
	VaultDir     string `json:"vault-dir"`
	DateFormat   string `json:"commit-date-format"`
	SyncFreqSec  int    `json:"sync-freq-sec"`
}

func applyConfig() (Config, error) {
	// Open json file
	jsonFile, err := os.Open("config.json")

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("opened config file")
	// defer file closure
	defer jsonFile.Close()
	// read in json as a byte array
	byteValue, _ := io.ReadAll(jsonFile)
	// create config var from the config struct
	var config Config
	// unmarshal byte array into our config struct
	json.Unmarshal(byteValue, &config)

	// review error messages
	if config.ObsidianPath == "" {
		return config, errors.New("no obsidian path specified. run ogg --config and provide the correct path")
	}

	if config.VaultDir == "" {
		return config, errors.New("no vault directory specified. run ogg -config and provide the correct path")
	}

	if err != nil {
		return config, err
	}

	return config, nil

}

// pulls remote changes before starting obsidian
func beforeObsidian(vaultDir string) error {

	// Check if vault contains .git
	if _, err := os.Stat(vaultDir + "/.git"); errors.Is(err, os.ErrNotExist) {
		return err
	}

	// Change directory to the vault directory defined in config file
	os.Chdir(vaultDir)

	fmt.Println("pulling most recent changes from remote...")

	// git pull
	cmd := exec.Command("git", "pull")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	if err != nil {
		fmt.Println("error pulling from repository: ", err)
	}

	return nil
}

// adds, commits and pushes the github repo
// called if sync is required or after obisidian closed
func commitChanges(commitFormat string) {
	fmt.Println("evaluating changes and pushing to repository")
	commit_time := time.Now().Format(commitFormat)

	cmd := exec.Command("git", "add", ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	if err != nil {
		fmt.Println(err)
	}

	cmd = exec.Command("git", "status")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()

	if err != nil {
		fmt.Println(err)
	}

	cmd = exec.Command("git", "commit", "-m", fmt.Sprintf("Updated on %s", commit_time))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()

	if err != nil {
		fmt.Println(err)
	}

	cmd = exec.Command("git", "push")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()

	if err != nil {
		fmt.Println(err)
	}
}

// main func to be cleaned up
func main() {

	// read in config and check if as expected
	config, err := applyConfig()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// change dir to vault dir, check if a git repo and git pull
	err = beforeObsidian(config.VaultDir)

	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	// run obsidian
	cmd := exec.Command(config.ObsidianPath)
	err = cmd.Start()

	if err != nil {
		fmt.Println("error starting obsidian: ", err)
		os.Exit(3)
	}

	// get process id
	processID := cmd.Process.Pid
	// init state (linux only)
	state := "not found"
	// init time_started to check syncing
	time_started := time.Now()

	// Wait for process to close
	// Sync periodically
	for {

		time.Sleep(1 * time.Second)

		process, err := ps.FindProcess(processID)

		process_details := fmt.Sprintf("%+v\n", process)

		re := regexp.MustCompile(`state:(\d+)`)

		match := re.FindStringSubmatch(process_details)

		if len(match) > 1 {
			state = match[1]
		}

		// Only sync if sync time greater than 0
		if config.SyncFreqSec > 0 {
			// get duration since last sync
			duration := time.Since(time_started)

			time_since_last_sync := (int(duration.Seconds()) % 60)

			// if time sync last sync greater than or equal to frequency commitChanges and restart timer
			if time_since_last_sync >= config.SyncFreqSec {
				commitChanges(config.DateFormat)
				time_started = time.Now()
			}
		}

		// if process closed break out of loop
		if err != nil || process == nil || state == "90" {
			break
		}

	}

	// commit changes after obsidian is closed
	commitChanges(config.DateFormat)

}
