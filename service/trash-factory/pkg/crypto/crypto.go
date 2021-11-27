package crypto

import (
	"bytes"
	"encoding/hex"
	"errors"
	"trash-factory/pkg/models"
)

type Cryptor struct {
	magic []byte
}

func NewCryptor(magic []byte) *Cryptor {
	return &Cryptor{
		magic: magic,
	}
}

func (cryptor *Cryptor) EncryptMsg(user *models.User, payload []byte) ([]byte, error) {
	ct, err := hex.DecodeString(user.TokenKey)
	if err != nil {
		return nil, err
	}

	payload = append(cryptor.magic, payload...)
	for i, b := range payload {
		ct = append(ct, b ^ user.Token[i % len(user.Token)])
	}
	return ct, nil
}


func (cryptor *Cryptor) DecryptMsg(user *models.User, msg []byte) ([]byte, error) {
	magicLen := len(cryptor.magic)
	if len(msg) <  magicLen {
		return nil, errors.New("incorrect message len")
	}

	decryptedBytes := make([]byte, len(msg))
	for i, b := range msg {
		decryptedBytes[i] = b ^ user.Token[i % len(user.Token)]
	}
	if bytes.Compare(decryptedBytes[:magicLen], cryptor.magic) != 0 {
		return nil, errors.New("incorrect message magic")
	}

	return decryptedBytes[magicLen:], nil
}