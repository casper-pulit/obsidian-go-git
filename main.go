package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
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

	fmt.Println("Opened config file")
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
	fmt.Println("Starting obsidian")

	// Check if vault contains .git
	if _, err := os.Stat(vaultDir + "/.git"); errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}

func main() {

	curr_os := runtime.GOOS
	fmt.Println(curr_os)
	config, err := applyConfig()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Config ok!")

	err = beforeObsidian(config.VaultDir)

	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	fmt.Println("beforeObsidian ok!")
}
