package email

import (
	"encoding/json"
	"io/ioutil"
)

// Config Stores default values
type Config struct {
	From         string `json:"from"`
	SMTPUsername string `json:"smtp_username"`
	SMTPPassword string `json:"smtp_password"`
	Server       string `json:"server"`
	Port         int64  `json:port`
}

var config Config

func configGetFrom() string {
	return config.From
}

// SetConfig store a global config setting
func SetConfig(newConfig Config) {
	config = newConfig
}

// GetConfig returns current global config settings
func GetConfig() *Config {
	return &config
}

// ReadConfig reads json file path and try to unmashal it to config
func ReadConfig(path string) error {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(file, &config)
}
