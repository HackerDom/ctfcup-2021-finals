package crypto

import (
	"bytes"
	"encoding/hex"
	"errors"
)

type Cryptor struct {
	magic []byte
}

func NewCryptor(magic []byte) *Cryptor {
	return &Cryptor{
		magic: magic,
	}
}

func (cryptor *Cryptor) EncryptMsg(tokenKey string, token, payload []byte) ([]byte, error) {
	ct, err := hex.DecodeString(tokenKey)
	if err != nil {
		return nil, err
	}

	payload = append(cryptor.magic, payload...)
	for i, b := range payload {
		ct = append(ct, b ^ token[i % len(token)])
	}
	return ct, nil
}


func (cryptor *Cryptor) DecryptMsg(token, msg []byte) ([]byte, error) {
	magicLen := len(cryptor.magic)
	if len(msg) <  magicLen {
		return nil, errors.New("incorrect message len")
	}

	decryptedBytes := make([]byte, len(msg))
	for i, b := range msg {
		decryptedBytes[i] = b ^ token[i % len(token)]
	}
	if bytes.Compare(decryptedBytes[:magicLen], cryptor.magic) != 0 {
		return nil, errors.New("incorrect message magic")
	}

	return decryptedBytes[magicLen:], nil
}