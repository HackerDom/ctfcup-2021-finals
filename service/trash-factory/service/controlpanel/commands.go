package main

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"math/rand"
)

type ControlPanel struct {
	Commands map[byte]interface{}
	DB *DataBase
}

func NewControlPanel() *ControlPanel {
	cp := ControlPanel{}
	cp.Commands = map[byte]interface{}{
		'\x09': cp.CreateContainer,
	}
	cp.DB = NewDataBase()
	return &cp
}

func (cp *ControlPanel) ProcessMessage(msg []byte) byte {
	if len(msg) < 9 {
		log.Warnf("Incorrect length of command: %x", msg)
		return '\x01'
	}

	plainText, err := cp.DecryptMsg(msg)
	if err != nil {
		log.Error(err)
		return '\x02'
	}

	statusCode := cp.RunCommand(plainText)

	return statusCode
}

func (cp *ControlPanel) DecryptMsg(msg []byte) ([]byte, error) {
	tokenKey := msg[:8]
	user, err := cp.DB.GetUser(fmt.Sprintf("%08x", tokenKey))
	if err != nil {
		return nil, err
	}
	payload := msg[8:]
	decrypted := msg[:8]
	magic := ""
	for i := 0; i < len(payload); i++ {
		decryptedByte := payload[i] ^ user.token[i % 8]
		if i == 0 || i == 1 || i == 2 {
			magic += fmt.Sprintf("%02x", decryptedByte)
			if i == 2 && magic != fmt.Sprintf("%02x", greeting) {
				return make([]byte,0), errors.New("decryption failed: incorrect magic key")
			}
			continue
		}
		decrypted = append(decrypted, decryptedByte)
	}
	return decrypted, nil
}

func (cp *ControlPanel) RunCommand(msg []byte) byte {
	tokenKey := fmt.Sprintf("%08x", msg[:8])
	command := msg[8]
	args := make([]byte, 0)
	if len(msg) > 9 {
		args = msg[9:]
	}

	if value, ok := cp.Commands[command]; ok {
		err := value.(func(string, []byte) error)(tokenKey, args)
		if err != nil {
			log.Errorf("func \\x%02x exec error: %s", command, err)
			return '\x03'
		}
		return '\x00'
	}
	log.Errorf("command \\x%x not found", command)
	return '\x04'
}

func (cp *ControlPanel) CreateContainer(tokenKey string, args []byte) error {

	fmt.Printf("Container creating... TOKEN %02x  ARGS %02x\n", tokenKey, args)
	if len(args) < 2 {
		return errors.New("not enough args (< 2)")
	}

	containerSize := args[0]
	if containerSize > '\x05' || containerSize < '\x01' {
		return errors.New("incorrect container size")
	}

	descSize := 50
	if descSize > len(args[1:]) {
		descSize = len(args[1:])
	}
	fmt.Println(args[1:descSize])
	container := Container{
		ID: fmt.Sprintf("%08x", rand.Uint64()),
		Size: containerSize,
		Description: fmt.Sprintf("%s", args[1:descSize]),
	}
	return cp.DB.SaveContainer(tokenKey, &container)
}

