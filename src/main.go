package main

import (
	"fmt"
	"log"
	"os"
	"plugin"

	"github.com/alecthomas/kingpin"
)

func getenvDefault(name string, defaultValue string) string {
	value := os.Getenv(name)
	if value == "" {
		return defaultValue
	}
	return value
}

type EncryptionPluginIface interface {
	Encrypt(cleartext []byte, args ...interface{}) (ciphertext []byte, err error)
	Decrypt(ciphertext []byte, args ...interface{}) (cleartext []byte, err error)
	Description() (pluginDescription string)
}

type encryptionPlugin struct {
	name        string
	encrypt     func(cleartext []byte, args ...interface{}) []byte
	decrypt     func(ciphertext []byte, args ...interface{}) []byte
	description func() string
	object      *plugin.Plugin
}

var encryptionPlugins []*encryptionPlugin

func loadEncPlugin(path string) (p *encryptionPlugin) {
	obj, err := plugin.Open(path)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	encryptFunc, err := obj.Lookup("Encrypt")
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	decryptFunc, err := obj.Lookup("Decrypt")
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	descriptionFunc, err := obj.Lookup("Description")
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	p = new(encryptionPlugin)
	p.name = path
	p.object = obj
	p.encrypt = encryptFunc.(func([]byte, ...interface{}) []byte)
	p.decrypt = decryptFunc.(func([]byte, ...interface{}) []byte)
	p.description = descriptionFunc.(func() string)

	return p
}

var (
	verbose = kingpin.Flag("verbose", "Verbose mode.").Short('v').Bool()
	plugins = kingpin.Flag("enc", "Load the specified plugin").Short('e').Strings()
)

func main() {
	kingpin.Version("0.0.1")
	kingpin.Parse()

	encryptionPlugins = make([]*encryptionPlugin, 0)

	for _, path := range *plugins {
		var p = loadEncPlugin(path)
		encryptionPlugins = append(encryptionPlugins, p)
		fmt.Print(string(p.encrypt(make([]byte, 0))))
		fmt.Print(string(p.decrypt(make([]byte, 0))))
		fmt.Print(p.description())
	}
}
