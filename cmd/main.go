package cmd

import (
	"flag"
	"log"

	"github.com/mahmoud-eskandari/gojen/internal/config"
	"github.com/mahmoud-eskandari/gojen/internal/generator"
	"github.com/mahmoud-eskandari/gojen/internal/repo"
)

func Execute() {
	conf := flag.String("c", "config.gojen.yaml", "Set config path")
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
