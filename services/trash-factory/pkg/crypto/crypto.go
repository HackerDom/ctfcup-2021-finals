package crypto

import (
	"bytes"
	"encoding/hex"
	"errors"
)

type Cryptor struct {
	Magic []byte
}

func NewCryptor(magic []byte) *Cryptor {
	return &Cryptor{
		Magic: magic,
	}
}

func (cryptor *Cryptor) EncryptMsg(tokenKey string, token, payload []byte) ([]byte, error) {
	ct, err := hex.DecodeString(tokenKey)
	if err != nil {
		return nil, err
	}
	payloadWithMagic := append(cryptor.Magic, payload...)
	for i, b := range payloadWithMagic {
		ct = append(ct, b ^ token[i % len(token)])
	}
	return ct, nil
}


func (cryptor *Cryptor) DecryptMsg(token, msg []byte) ([]byte, error) {
	magicLen := len(cryptor.Magic)
	if len(msg) <  magicLen {
		return nil, errors.New("incorrect message len")
	}

	decryptedBytes := make([]byte, len(msg))
	for i, b := range msg {
		decryptedBytes[i] = b ^ token[i % len(token)]
	}
	if bytes.Compare(decryptedBytes[:magicLen], cryptor.Magic) != 0 {
		return nil, errors.New("incorrect message Magic")
	}

	return decryptedBytes[magicLen:], nil
}