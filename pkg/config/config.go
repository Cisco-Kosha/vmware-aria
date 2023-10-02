package config

import (
	"flag"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	serverUrl string
	apiKey    string
}

func Get() *Config {
	conf := &Config{}
	flag.StringVar(&conf.apiKey, "apiKey", os.Getenv("API_TOKEN"), "API Token")

	flag.StringVar(&conf.serverUrl, "serverUrl", os.Getenv("SERVER_URL"), "Server Url")

	flag.Parse()

	return conf
}

func (c *Config) GetAPIKey() string {
	return c.apiKey
}

func (c *Config) GetServerURL() string {
	c.serverUrl = strings.TrimSuffix(c.serverUrl, "/")
	u, _ := url.Parse(c.serverUrl)
	if u.Scheme == "" {
		return "https://" + c.serverUrl
	} else {
		return c.serverUrl
	}
}

func (c *Config) GetServerHost() string {
	c.serverUrl = strings.TrimSuffix(c.serverUrl, "/")
	u, _ := url.Parse(c.serverUrl)
	if u.Scheme == "" {
		return u.Host
	} else {
		return u.Host
	}
}
