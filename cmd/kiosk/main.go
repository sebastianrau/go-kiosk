package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/SebastianRau/kiosk/pkg/kiosk"
	"github.com/ilyakaznacheev/cleanenv"
)

var (
	EXIT_NO_CONFIG   = 1
	EXIT_NO_URL      = 2
	EXIT_INVALID_URL = 3
)

var (
	version = "0.0.0xxx"
)

func main() {

	var (
		configPath = flag.String("c", "", "Path to configuration file (config.yaml)")
	)
	flag.Parse()

	var cfg kiosk.Config

	log.Printf("Kiosk by SR (version: %s)\n", version)

	if *configPath == "" {
		homedir := os.Getenv("HOME")
		if homedir == "" {
			log.Panic("No config file found")
			os.Exit(EXIT_NO_CONFIG)
		}
		*configPath = fmt.Sprintf("%s/.kiosk/config.yml", homedir)
	}

	if err := cleanenv.ReadConfig(*configPath, &cfg); err != nil {
		log.Panic("Error reading config file: ", err)
		os.Exit(EXIT_NO_CONFIG)
	}

	cfg.LogPrintConfig()

	// make sure the url has content
	if cfg.Url == "" {
		log.Panicf("No URL found")
		os.Exit(EXIT_NO_URL)
	}

	// validate url
	url, err := url.ParseRequestURI(cfg.Url)
	if err != nil {
		log.Panicf("No URL found")
		os.Exit(EXIT_INVALID_URL)
	}
	cfg.Url = url.String()

	setEnvironment()

	kiosk.Kiosk(&cfg)

}

func setEnvironment() {
	// for linux/X display must be set
	var displayEnv = os.Getenv("DISPLAY")
	if displayEnv == "" {
		log.Println("DISPLAY not set, autosetting to :0.0")
		if err := os.Setenv("DISPLAY", ":0.0"); err != nil {
			log.Println("Error setting DISPLAY", err.Error())
		}
		displayEnv = os.Getenv("DISPLAY")
	}

	log.Println("DISPLAY=", displayEnv)

	var xAuthorityEnv = os.Getenv("XAUTHORITY")
	if xAuthorityEnv == "" {
		log.Println("XAUTHORITY not set, autosetting")
		// use HOME of current user
		var homeEnv = os.Getenv("HOME")

		if err := os.Setenv("XAUTHORITY", homeEnv+"/.Xauthority"); err != nil {
			log.Println("Error setting XAUTHORITY", err.Error())
		}
		xAuthorityEnv = os.Getenv("XAUTHORITY")
	}

	log.Println("XAUTHORITY=", xAuthorityEnv)
}
