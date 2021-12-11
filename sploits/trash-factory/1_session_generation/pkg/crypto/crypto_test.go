package crypto

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"testing"
	"1_session_generation/pkg/models"
)

func generateBytes() []byte {
	buff := make([]byte, 8)
	binary.LittleEndian.PutUint64(buff, rand.Uint64())
	return buff
}

func GetUser() *models.User {
	return &models.User{
		TokenKey: hex.EncodeToString(generateBytes()),
		Token: generateBytes(),
	}
}

func GetCryptor() (*Cryptor, []byte) {
	magic := "031337"
	magicBytes, err := hex.DecodeString(magic)
	if err != nil {
		log.Panicln(err)
	}
	return &Cryptor{
		Magic: magicBytes,
	}, magicBytes
}


func TestEncryption(t *testing.T) {
	user := GetUser()
	cryptor, magicBytes := GetCryptor()
	payload := generateBytes()

	ct, err := cryptor.EncryptMsg(user.TokenKey, user.Token, payload)
	if err != nil {
		log.Panicln(err)
	}

	tokenKeyBytes, err := hex.DecodeString(user.TokenKey)
	if err != nil {
		log.Panicln(err)
	}

	if bytes.Compare(ct[:8], tokenKeyBytes) != 0 {
		t.Fatalf(`unexpected tokenKey. expected: %x, got: %x\n`, tokenKeyBytes, ct[:8])
	}

	payloadWithMagic := append(magicBytes, payload...)
	for i, b := range ct[8:] {
		if b ^ user.Token[i % len(user.Token)] != payloadWithMagic[i] {
			t.Fatalf(`cant decrypt message. payload: %x, encrypted payload %x. \n Byte %d expect: %x, got: %x\n`,
				payload, ct[8:], i, payloadWithMagic[i], b ^ user.Token[i % len(user.Token)])
		}
	}
}

func TestDecryption(t *testing.T) {
	user := GetUser()
	cryptor, _ := GetCryptor()
	payload := generateBytes()
	ct, err := cryptor.EncryptMsg(user.TokenKey, user.Token, payload)
	if err != nil {
		log.Panicln(err)
	}
	pt, err := cryptor.DecryptMsg(user.Token, ct[8:])
	if bytes.Compare(payload, pt) != 0 || err != nil {
		t.Fatalf(`cant get same message. Payload: %x, pt: %x. Err: %v`, payload, pt, err)
	}
}
