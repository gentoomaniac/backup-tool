package main

import (
	"encoding/base64"

	"./lib/crypt/aes256"

	"github.com/alecthomas/kingpin"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)
}

var (
	verbose = kingpin.Flag("verbose", "Verbose mode.").Short('v').Bool()
)

func main() {
	kingpin.Version("0.0.1")
	kingpin.Parse()

	iv, _ := aes256.GenerateIV()
	log.WithFields(log.Fields{
		"iv": base64.StdEncoding.EncodeToString(iv),
	}).Debug("iv generated")

	secret, _ := aes256.GenerateSecret()
	log.WithFields(log.Fields{
		"secret": base64.StdEncoding.EncodeToString(secret),
	}).Debug("secret generated")

	encrypted, _ := aes256.Encrypt([]byte("Hello, World!"), secret, iv)
	log.WithFields(log.Fields{
		"encrypted": base64.StdEncoding.EncodeToString(encrypted),
	}).Debug("text encrypted")

	decrypted, _ := aes256.Decrypt(encrypted, secret, iv)
	log.WithFields(log.Fields{
		"decrypted": string(decrypted),
	}).Debug("text decrypted")
}
