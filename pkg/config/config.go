package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
)

const (
	ConfigFileName       = "weave-gitops-config.json"
	WrongConfigFormatMsg = `Your CLI configuration for Weave GitOps should represent a JSON object format:
{
	"analytics": true,
	"userId": ""
}
Please set you configuration with: gitops set config`
)

var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789+/")

type GitopsCLIConfig struct {
	Analytics bool   `json:"analytics"`
	UserID    string `json:"userId"`
}

func (config *GitopsCLIConfig) String() (string, error) {
	data, err := json.MarshalIndent(&config, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error encoding config: %w", err)
	}

	return string(data), nil
}

// GetConfig reads the CLI configuration for Weave GitOps from the config file
func GetConfig(shouldCreate bool) (*GitopsCLIConfig, error) {
	configPath, err := getConfigPath(ConfigFileName)
	if err != nil {
		return nil, err
	}

	configFile, err := openConfigFile(configPath, shouldCreate)
	if err != nil {
		return nil, fmt.Errorf("error opening config file: %w", err)
	}

	defer configFile.Close()

	config := &GitopsCLIConfig{}

	data, err := readData(configFile)
	if err != nil && !shouldCreate {
		return nil, fmt.Errorf("error reading config data: %w", err)
	}

	if len(data) == 0 {
		if !shouldCreate {
			return nil, fmt.Errorf("empty config file detected at path: %s", configPath)
		}
	} else if err = parseConfig(data, config); err != nil {
		if shouldCreate {
			// just replace invalid config with default config
		} else {
			return nil, fmt.Errorf("error reading config from file: %w", err)
		}
	}

	return config, nil
}

// SaveConfig saves the CLI configuration for Weave GitOps to the config file
func SaveConfig(config *GitopsCLIConfig) error {
	configPath, err := getConfigPath(ConfigFileName)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(&config, "", "  ")
	if err != nil {
		return fmt.Errorf("error encoding config: %w", err)
	}

	configFile, err := openConfigFile(configPath, true)
	if err != nil {
		return fmt.Errorf("error opening config file: %w", err)
	}

	defer configFile.Close()

	err = configFile.Truncate(0)
	if err != nil {
		return fmt.Errorf("error truncating config file: %w", err)
	}

	_, err = configFile.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("error setting offset in config file: %w", err)
	}

	_, err = configFile.Write(data)
	if err != nil {
		return fmt.Errorf("error writing to config file: %w", err)
	}

	return nil
}

// getConfigPath composes the path of the config file for the GitOps CLI
func getConfigPath(filename string) (string, error) {
	var err error

	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("user config directory could not be determined: %w", err)
	}

	configPath := filepath.Join(userConfigDir, filename)

	return configPath, nil
}

// openConfigFile opens existing config file or creates one if it does not exist
func openConfigFile(configPath string, shouldCreate bool) (*os.File, error) {
	_, err := os.Stat(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) && shouldCreate {
			// we'll just create a new file
		} else {
			return nil, err
		}
	}

	var configFile *os.File

	var (
		flag int
		perm fs.FileMode
	)

	if shouldCreate {
		flag = os.O_RDWR | os.O_CREATE
		perm = 0666
	} else {
		flag = os.O_RDONLY
		perm = 0444
	}

	configFile, err = os.OpenFile(configPath, flag, perm)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}

	return configFile, nil
}

// readData reads the data from the config file
func readData(configFile *os.File) ([]byte, error) {
	data, err := io.ReadAll(configFile)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// parseConfig unmarshals the CLI configuration for Weave GitOps from the JSON blob
func parseConfig(data []byte, config *GitopsCLIConfig) error {
	err := json.Unmarshal(data, &config)
	if err != nil {
		return fmt.Errorf("error parsing config JSON: %w", err)
	}

	return nil
}

// GenerateUserID generates a string of specified length made of random characters and encodes it in base64 format
func GenerateUserID(numChars int, seed int64) string {
	rand.Seed(seed)

	b := make([]byte, numChars)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}
