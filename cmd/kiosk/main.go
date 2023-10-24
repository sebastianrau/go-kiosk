package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/SebastianRau/kiosk/pkg/kiosk"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/palantir/go-encrypted-config-value/encryptedconfigvalue"
)

var (
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

func CreateKeyFile(createKeyFiles string) {
	if createKeyFiles != "" {
		fmt.Println("Key will be generated")

		keyPair, err := encryptedconfigvalue.AES.GenerateKeyPair()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(-1)
		}

		f, err := os.Create(createKeyFiles)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(EXIT_KEY_GENERATION_FAIL)
		}
		defer f.Close()
		f.Write([]byte(keyPair.EncryptionKey.ToSerializable()))

		fmt.Println("Key generated.")
		os.Exit(0)

	}
}

func EncryptString(encryptString string, encryptionKeyFile string) {
	if encryptString != "" && encryptionKeyFile != "" {
		fmt.Println("Encrypting:")
		fmt.Println(encryptString)

		fmt.Println("Read private key")
		privateKeyFileBytes, err := os.ReadFile(encryptionKeyFile)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(EXIT_PRIVATE_KEY_NOT_FOUND)
		}

		privateKey, err := encryptedconfigvalue.NewKeyWithType(string(privateKeyFileBytes))
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(EXIT_KEY_GENERATION_FAIL)
		}

		fmt.Println("Encrypt string")
		encryptedVal, err := encryptedconfigvalue.AES.Encrypter().Encrypt(encryptString, privateKey)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(EXIT_KEY_GENERATION_FAIL)
		}

		fmt.Println()
		fmt.Println("${" + encryptedVal.ToSerializable() + "}")
		fmt.Println()
		os.Exit(0)
	}
}

func CheckEncryption(configPath string) bool {
	test, _ := os.ReadFile(configPath)
	return encryptedconfigvalue.ContainsEncryptedConfigValueStringVars(test)
}

func DecryptConfig(cfg *kiosk.Config, encryptionKeyFile string) {
	if encryptionKeyFile == "" {
		encryptionKeyFile = fmt.Sprintf("%s/.kiosk/kiosk_id", getHomeDir())
	}

	log.Println("Read private key")
	privateKeyFileBytes, err := os.ReadFile(encryptionKeyFile)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(EXIT_PRIVATE_KEY_NOT_FOUND)
	}

	privateKey, err := encryptedconfigvalue.NewKeyWithType(string(privateKeyFileBytes))
	if err != nil {
		log.Panic(err.Error())
		os.Exit(EXIT_KEY_GENERATION_FAIL)
	}
	encryptedconfigvalue.DecryptEncryptedStringVariables(cfg, privateKey)
}

func getHomeDir() string {
	homedir := os.Getenv("HOME")
	if homedir == "" {
		log.Panic("No config file found")
		os.Exit(EXIT_NO_CONFIG)
	}

	return homedir
}

func main() {

	var (
		configPath        = flag.String("c", "", "Path to configuration file (config.yaml)")
		createKeyFiles    = flag.String("createKeys", "", "Generate a new AES Key Pair with given filename")
		encryptString     = flag.String("s", "", "String to encrypt")
		encryptionKeyFile = flag.String("k", "", "Encryption Key File")
	)
	flag.Parse()

	var cfg kiosk.Config

	log.Printf("Kiosk by SR (version: %s)\n", version)

	CreateKeyFile(*createKeyFiles)

	EncryptString(*encryptString, *encryptionKeyFile)

	if *configPath == "" {
		*configPath = fmt.Sprintf("%s/.kiosk/config.yml", getHomeDir())
	}

	if err := cleanenv.ReadConfig(*configPath, &cfg); err != nil {
		log.Panic("Error reading config file: ", err)
		os.Exit(EXIT_NO_CONFIG)
	}

	if CheckEncryption(*configPath) {
		log.Println("Config file contains encrypted string")
		DecryptConfig(&cfg, *encryptionKeyFile)
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
	os.Exit(0)
	kiosk.Kiosk(&cfg)

}
