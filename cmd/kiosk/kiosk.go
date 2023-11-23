package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"

	easyconfig "github.com/sebastianrau/go-easyConfig/pkg/easyConfig"
	"github.com/sebastianrau/kiosk/pkg/kiosk"
)

const (
	EXIT_NO_CONFIG             = 1
	EXIT_NO_URL                = 2
	EXIT_INVALID_URL           = 3
	EXIT_KEY_GENERATION_FAIL   = 4
	EXIT_PRIVATE_KEY_NOT_FOUND = 5
)

var (
	version = "0.0.0xxx"
)

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

func getHomeDir() string {
	homedir := os.Getenv("HOME")
	if homedir == "" {
		log.Panic("No config file found")
		os.Exit(EXIT_NO_CONFIG)
	}

	return homedir
}

func checkPathes(configPath *string, encryptionKeyFile *string) {

	if *configPath == "" {
		*configPath = fmt.Sprintf("%s/.kiosk/config.yml", getHomeDir())
	}

	if *encryptionKeyFile == "" {
		*encryptionKeyFile = fmt.Sprintf("%s/.kiosk/kiosk_id", getHomeDir())
	}
}

func main() {
	var (
		configPath        = flag.String("c", "", "Path to configuration file (config.yaml)")
		encryptionKeyPath = flag.String("k", "", "Encryption Key File")
		cfg               kiosk.Config
	)
	flag.Parse()

	log.Printf("Kiosk by SR (version: %s)\n", version)

	//Check if Pathes are given, else use home directory/.kiosk/
	checkPathes(configPath, encryptionKeyPath)

	// parse yaml confing
	err := easyconfig.FromFile(*configPath, &cfg)
	if err != nil {
		log.Printf("Error reading config file: %s", err.Error())
		os.Exit(EXIT_NO_CONFIG)
	}

	//
	err = easyconfig.DecryptFromFile(*encryptionKeyPath, &cfg)
	if err != nil {
		log.Printf("Error decrypting config file: %s", err.Error())
		os.Exit(EXIT_PRIVATE_KEY_NOT_FOUND)
	}

	// make sure the url has content
	if cfg.Url == "" {
		log.Printf("No URL found")
		os.Exit(EXIT_NO_URL)
	}

	// validate url
	url, err := url.ParseRequestURI(cfg.Url)
	if err != nil {
		log.Printf("No URL found")
		os.Exit(EXIT_INVALID_URL)
	}
	cfg.Url = url.String()

	setEnvironment()

	err = kiosk.Kiosk(&cfg)
	if err != nil {
		log.Println(err.Error())
	}

}
