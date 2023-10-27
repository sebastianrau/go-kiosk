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

func createKeyFile(createKeyFiles string) {

	fmt.Println("Key will be generated")

	keyPair, err := encryptedconfigvalue.RSA.GenerateKeyPair()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	serializedPublicKey := keyPair.EncryptionKey.ToSerializable()
	serializedPrivateKey := keyPair.DecryptionKey.ToSerializable()

	f, err := os.Create(createKeyFiles + ".pup")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(EXIT_KEY_GENERATION_FAIL)
	}
	f.Write([]byte(serializedPublicKey))
	f.Close()

	f, err = os.Create(createKeyFiles)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(EXIT_KEY_GENERATION_FAIL)
	}
	f.Write([]byte(serializedPrivateKey))
	f.Close()

	fmt.Println("Keys generated.")
}

func encryptString(encryptString string, encryptionKeyFile string) {

	fmt.Println("Encrypting:")
	fmt.Println(encryptString)

	fmt.Println("Read private key")
	publicKeyFileBytes, err := os.ReadFile(encryptionKeyFile)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(EXIT_PRIVATE_KEY_NOT_FOUND)
	}

	fmt.Println(string(publicKeyFileBytes))

	privateKey, err := encryptedconfigvalue.NewKeyWithType(string(publicKeyFileBytes))
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(EXIT_KEY_GENERATION_FAIL)
	}

	fmt.Println("Encrypt string")
	encryptedVal, err := encryptedconfigvalue.RSA.Encrypter().Encrypt(encryptString, privateKey)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(EXIT_KEY_GENERATION_FAIL)
	}

	fmt.Println()
	fmt.Println("${" + encryptedVal.ToSerializable() + "}")
	fmt.Println()

}

func checkEncryption(configPath string) bool {
	test, _ := os.ReadFile(configPath)
	return encryptedconfigvalue.ContainsEncryptedConfigValueStringVars(test)
}

func DecryptConfig(cfg *kiosk.Config, encryptionKeyFile string) {

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
		createKeyFiles    = flag.String("createKeys", "", "Generate a new AES Key Pair with given filename")
		stringToEncrypt   = flag.String("s", "", "String to encrypt")
		encryptionKeyPath = flag.String("k", "", "Encryption Key File")
	)
	flag.Parse()

	var cfg kiosk.Config

	log.Printf("Kiosk by SR (version: %s)\n", version)

	//Check if Pathes are given, else use home directory/.kiosk/
	checkPathes(configPath, encryptionKeyPath)

	// If create key files argument is given create a new keyfile and exit
	if *createKeyFiles != "" {
		createKeyFile(*createKeyFiles)
		os.Exit(0)
	}

	// Encrypt String argument is given. encrypt and exit
	if *stringToEncrypt != "" && *encryptionKeyPath != "" {
		encryptString(*stringToEncrypt, *encryptionKeyPath)
		os.Exit(0)
	}

	if err := cleanenv.ReadConfig(*configPath, &cfg); err != nil {
		log.Panic("Error reading config file: ", err)
		os.Exit(EXIT_NO_CONFIG)
	}

	if checkEncryption(*configPath) {
		log.Println("Config file contains encrypted string")
		DecryptConfig(&cfg, *encryptionKeyPath)
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
