package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"net"
	"time"
	"trash-factory/pkg/commands"
	"trash-factory/pkg/crypto"
	"trash-factory/pkg/models"
)

type Client struct {
	address       string
	tokenKey      string
	tokenKeyBytes []byte
	token         []byte
}

func NewClient(addr string, tokenKey string, token string) *Client {
	tokenKeyBytes, err := hex.DecodeString(tokenKey)
	if err != nil {
		log.Fatal("tokenkey is incorrect")
	}

	tokenBytes, err := hex.DecodeString(token)
	if err != nil {
		log.Fatal("token is incorrect")
	}
	return &Client{
		address:       addr,
		tokenKey:      tokenKey,
		tokenKeyBytes: tokenKeyBytes,
		token:         tokenBytes,
	}
}

func (client *Client) readBytes(conn net.Conn) ([]byte, error) {
	buffer := make([]byte, 128)
	bytesCount, err := conn.Read(buffer)
	log.Info(bytesCount, buffer)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	if bytesCount < 1 {
		return nil, errors.New("incorrect bytes count")
	}
	return buffer, nil
}

func (client *Client) sendMessage(msg []byte) ([]byte, error) {
	dialer := net.Dialer{Timeout: 10 * time.Second}
	conn, err := dialer.Dial("tcp", client.address)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	magic, err := client.readBytes(conn)

	if err != nil {
		return nil, err
	}

	cryptor := crypto.NewCryptor(magic)
	ct, err := cryptor.EncryptMsg(client.tokenKey, client.token, msg)
	if err != nil {
		return nil, err
	}
	bytesCount, err := conn.Write(ct)
	if err != nil {
		return nil, err
	}
	if bytesCount != len(ct) {
		return nil, errors.New("bytes sending error")
	}

	response, err := client.readBytes(conn)
	if err != nil {
		return nil, err
	}

	if len(response) < 9 || bytes.Compare(client.tokenKeyBytes, response[:8]) != 0 {
		return nil, errors.New("incorrect response")
	}

	decryptedResp, err := cryptor.DecryptMsg(client.token, response[8:])
	if err != nil {
		return nil, err
	}

	return decryptedResp, nil
}

func (client *Client) createUser() (string, error) {
	msg := []byte{commands.CreateUser}
	tokenKey := make([]byte, 8)
	binary.LittleEndian.PutUint64(tokenKey, rand.Uint64())
	msg = append(msg, tokenKey...)
	token := make([]byte, 8)
	binary.LittleEndian.PutUint64(token, rand.Uint64())
	msg = append(msg, token...)
	response, err := client.sendMessage(msg)
	if err != nil {
		return "", err
	}
	if response[0] != '\x00' {
		return "", errors.New(fmt.Sprintf("cant create user: %02x", response[0]))
	}

	return hex.EncodeToString(tokenKey), nil
}

func (client *Client) getUser(tokenKey string) (*models.User, error) {
	tokenKeyBytes, err := hex.DecodeString(tokenKey)
	if err != nil {
		return nil, err
	}
	msg := []byte{commands.GetUser}
	msg = append(msg, tokenKeyBytes...)
	response, err := client.sendMessage(msg)
	if err != nil {
		return nil, err
	}
	if response[0] != '\x00' {
		return nil, errors.New(fmt.Sprintf("cant get user: %02x", response[0]))
	}
	user, err := models.DeserializeUser(response[1:])
	if err != nil {
		return nil, err
	}
	return &user, nil
}
