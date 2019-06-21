// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import "time"

type Config struct {
	Period            time.Duration `config:"period"`
	Channel           string        `config:"channel"`
	Organization      string        `config:"organization"`
	ConnectionProfile string        `config:"connection_profile"`
}

var DefaultConfig = Config{
	Period:            1 * time.Second,
	Channel:           "mychannel",
	Organization:      "Org1",
	ConnectionProfile: "connection.yaml",
}
