package config

import (
	"io/fs"
	"os"

	"github.com/mahmoud-eskandari/goje"
	"gopkg.in/yaml.v3"
)

var Conf Config

func ReadConfig(path string) error {
	file, err := os.OpenFile(path, os.O_RDONLY, fs.FileMode(os.O_RDONLY))
	if err != nil {
		return err
	}

	err = yaml.NewDecoder(file).Decode(&Conf)
	if err != nil {
		return err
	}

	return nil
}

type Config struct {
	DB      goje.DBConfig `yaml:"db"`
	Ignore  []string      `yaml:"ignore"`
	Tags    []string      `yaml:"tags"`
	Pkg     string        `yaml:"pkg"`
	Dir     string        `yaml:"dir"`
	Replace bool          `yaml:"replace"`
}
