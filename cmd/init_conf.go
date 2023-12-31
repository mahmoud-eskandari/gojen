package cmd

import (
	"os"

	"github.com/mahmoud-eskandari/goje"
	"github.com/mahmoud-eskandari/gojen/internal/config"
	"gopkg.in/yaml.v3"
)

// init config file
func initConfig(path string) {
	var conf config.Config

	conf.Tags = []string{"db", "json"}
	conf.Pkg = "models"
	conf.Dir = "./models"
	conf.Replace = true
	conf.DB = goje.DBConfig{
		Driver: "mysql",
		Host:   "127.0.0.1",
		Port:   3306,
		User:   "root",
		Schema: "database",
	}

	file, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	err = yaml.NewEncoder(file).Encode(conf)
	if err != nil {
		panic(err)
	}
}
