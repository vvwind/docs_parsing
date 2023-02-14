package main

import "C"
import (
	"github.com/spf13/viper"
	"log"
	"parser/internal/services"
)

func main() {
	viper.SetConfigName("config")
	viper.SetConfigFile("config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	myScraper := services.Scraper{}
	errorScraping := myScraper.Start()
	if errorScraping != nil {
		log.Fatalf("Got error scraping: %v", errorScraping)
	}

	myGDoc := services.Docs{}
	if errGDoc := myGDoc.Init(); errGDoc != nil {
		log.Fatalf("Got error creating token: %v", errGDoc)
	}

	if errWriting := myGDoc.Start(&myScraper); errWriting != nil {
		log.Fatalf("Got error writing doc: %v", errWriting)
	}

}
