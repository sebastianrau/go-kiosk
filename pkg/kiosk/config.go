package kiosk

import "log"

// Config configuration for backend.
type Config struct {
	WindowPosition          string `yaml:"window-position"`
	IgnoreCertificateErrors bool   `yaml:"ignore-certificate-errors"`
	LoginMethod             string `yaml:"login-method"`
	Username                string `yaml:"username"`
	Password                string `yaml:"password"`
	Url                     string `yaml:"url"`
	Token                   string `yaml:"key"`
}

func (c *Config) LogPrintConfig() {
	log.Printf("LoginMethod: %s", c.LoginMethod)
	log.Printf("Username: %s", c.Username)
	log.Printf("Password: %s", "*REDACTED*")
	log.Printf("Apikey: %s", c.Token)
	log.Printf("Url: %s", c.Url)
	log.Printf("IgnoreCertificateErrors: %t", c.IgnoreCertificateErrors)
	log.Printf("WindowPosition: %s", c.WindowPosition)
}
