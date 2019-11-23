package main

var (
	description = "Some text here"
)

type SHA256 struct{}

func (p *SHA256) Encrypt(cleartext []byte, args ...interface{}) (ciphertext []byte, err error) {

	return
}

func (p *SHA256) Decrypt(ciphertext []byte, args ...interface{}) (cleartext []byte, err error) {

	return
}

func (p *SHA256) Description() string {
	return description
}
