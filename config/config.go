package config

import (
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

type Rule struct {
	SrcAddr  string `yaml:"srcAddr,omitempty" validate:"ip|fqdn"`
	SrcPort  int    `yaml:"srcPort,omitempty" validate:"min=1,max=65535"`
	DstAddr  string `yaml:"dstAddr,omitempty" validate:"ip|fqdn"`
	DstPort  int    `yaml:"dstPort,omitempty" validate:"min=1,max=65535"`
	Protocol string `yaml:"protocol,omitempty" validate:"oneof=tcp udp both"`
}

type Config struct {
	Rules []Rule `yaml:"rules,omitempty" validate:"dive,required"`
}

func IsRemote(src string) bool {
	_, err := url.ParseRequestURI(src)
	return err == nil
}

func NewConfig(src string) (Config, error) {
	if IsRemote(src) {
		return NewRemoteConfig(src)
	}

	return NewLocalConfig(src)
}

func NewLocalConfig(src string) (Config, error) {
	b, err := readFromLocal(src)
	if err != nil {
		return Config{}, err
	}

	return newConfig(b)
}
func NewRemoteConfig(src string) (Config, error) {
	b, err := readFromRemote(src)
	if err != nil {
		return Config{}, err
	}

	return newConfig(b)
}

func newConfig(b []byte) (Config, error) {
	c, err := parse(b)
	if err != nil {
		return Config{}, err
	}

	err = validate(&c)
	if err != nil {
		return Config{}, err
	}

	return c, nil
}

func readFromLocal(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func readFromRemote(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func parse(b []byte) (Config, error) {
	var c Config
	err := yaml.Unmarshal(b, &c)
	if err != nil {
		return Config{}, err
	}

	return c, nil
}

func validate(c *Config) error {
	addr, err := defaultSrcAddr()
	if err != nil {
		return err
	}

	for i := range c.Rules {
		if c.Rules[i].SrcAddr == "" {
			c.Rules[i].SrcAddr = addr
		}
	}

	return validator.New().Struct(c)
}

func defaultSrcAddr() (string, error) {
	conn, err := net.Dial("udp", "223.5.5.5:53")
	if err != nil {
		return "", err
	}
	defer conn.Close()
	return strings.Split(conn.LocalAddr().String(), ":")[0], nil
}
