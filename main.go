package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"time"

	"github.com/mitchellh/go-ps"
)

// Config struct for reading in json config file
type Config struct {
	ObsidianPath string `json:"obsidian-path"`
	VaultDir     string `json:"vault-dir"`
	DateFormat   string `json:"commit-date-format"`
	SyncFreqMins int    `json:"sync-freq-min"`
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

func beforeObsidian(vaultDir string) error {

	// Check if vault contains .git
	if _, err := os.Stat(vaultDir + "/.git"); errors.Is(err, os.ErrNotExist) {
		return err
	}

	os.Chdir(vaultDir)

	fmt.Println("pulling most recent changes from remote...")

	cmd := exec.Command("git", "pull")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	if err != nil {
		fmt.Println("error pulling from repository: ", err)
	}

	return nil
}

func commitChanges(commitFormat string) {
	fmt.Println("obsidian closed")
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

func main() {

	curr_os := runtime.GOOS
	fmt.Println(curr_os)
	config, err := applyConfig()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = beforeObsidian(config.VaultDir)

	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	cmd := exec.Command(config.ObsidianPath)
	err = cmd.Start()

	if err != nil {
		fmt.Println("error starting Obsidian: ", err)
		os.Exit(3)
	}

	processID := cmd.Process.Pid
	state := "not found"
	time_started := time.Now()

	// Wait for process to close

	for {

		time.Sleep(1 * time.Second)

		process, err := ps.FindProcess(processID)

		process_details := fmt.Sprintf("%+v\n", process)

		re := regexp.MustCompile(`state:(\d+)`)

		match := re.FindStringSubmatch(process_details)

		if len(match) > 1 {
			state = match[1]
		}

		fmt.Println(time_started)

		duration := time.Since(time_started)

		time_since_last_sync := (int(duration.Seconds()) % 60)

		if time_since_last_sync >= config.SyncFreqMins {
			commitChanges(config.DateFormat)
			time_started = time.Now()
		}

		if err != nil || process == nil || state == "90" {
			break
		}

	}

	commitChanges(config.DateFormat)

}
