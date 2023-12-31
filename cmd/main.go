package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mahmoud-eskandari/goje"
	"github.com/mahmoud-eskandari/gojen/internal/config"
	"github.com/mahmoud-eskandari/gojen/internal/generator"
	"github.com/mahmoud-eskandari/gojen/internal/repo"
	"gopkg.in/yaml.v3"
)

func Execute() {
	conf := flag.String("c", "config.gojen.yaml", "Set config path")

	if len(os.Args) > 1 && strings.ToLower(os.Args[1]) == "init" {
		initConfig(*conf)
		fmt.Println("Config file created successfully!")
		fmt.Printf("Edit [%s] to set database connection and other options", *conf)
		return
	}

	err := config.ReadConfig(*conf)
	if err != nil {
		log.Fatalf("Error when reading config: %+v", err)
	}

	err = repo.Connect()
	if err != nil {
		log.Fatalf("Error on connecting to database: %+v", err)
	}

	err = generator.Generate()
	if err != nil {
		log.Fatalf("Error on generate files: %+v", err)
	}

}

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
