package api

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
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
	isAdmin       bool
}

type Response struct {
	statusCode       byte
	payload          []byte
	decryptedPayload []byte
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
		isAdmin:       false,
	}
}

func NewAdminClient(addr string) *Client {
	return &Client{
		address:       addr,
		tokenKey:      "",
		tokenKeyBytes: nil,
		token:         nil,
		isAdmin:       true,
	}
}

func (client *Client) ParseResponse(cryptor *crypto.Cryptor, data []byte) (*Response, error) {
	if len(data) < 1+len(client.tokenKeyBytes)+len(cryptor.Magic) {
		return nil, errors.New("incorrect response len")
	}

	if bytes.Compare(data[1:len(client.tokenKeyBytes)+1], client.tokenKeyBytes) != 0 {
		return nil, errors.New("incorrect returned tokenkey")
	}

	payload := data[1+len(client.tokenKeyBytes):]
	decryptedPayload, err := cryptor.DecryptMsg(client.token, payload)
	if err != nil {
		return nil, err
	}

	return &Response{
		statusCode:       data[0],
		payload:          payload,
		decryptedPayload: decryptedPayload,
	}, nil
}

func (client *Client) sendMessage(msg []byte) (*Response, error) {
	dialer := net.Dialer{Timeout: 10 * time.Second}
	conn, err := dialer.Dial("tcp", client.address)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	magic, err := client.readBytesOnce(conn)

	if err != nil {
		return nil, err
	}

	if client.isAdmin {
		err = client.FetchAdminCredentials(err, conn)
		if err != nil {
			return nil, err
		}
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

	response, err := client.readaAllByte(conn)
	if err != nil {
		return nil, err
	}
	log.Infof("RESPONSE: %x", response)

	parsedResponse, err := client.ParseResponse(cryptor, response)
	if err != nil {
		return nil, err
	}

	return parsedResponse, nil
}

func (client *Client) FetchAdminCredentials(err error, conn net.Conn) error {
	adminUserBytes, err := client.readBytesOnce(conn)
	if err != nil {
		return err
	}
	adminUser, err := models.DeserializeUser(adminUserBytes)
	if err != nil {
		return err
	}

	client.tokenKey = adminUser.TokenKey
	client.token = adminUser.Token
	client.tokenKeyBytes, err = hex.DecodeString(adminUser.TokenKey)
	if err != nil {
		return err
	}
	return nil
}

func (client *Client) readBytesOnce(conn net.Conn) ([]byte, error) {
	buffer := make([]byte, 128)
	bytesCount, err := conn.Read(buffer)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	if bytesCount < 1 {
		return nil, errors.New("incorrect bytes count")
	}
	return buffer[:bytesCount], nil
}

func (client *Client) readaAllByte(conn net.Conn) ([]byte, error) {
	allBytes, err := ioutil.ReadAll(conn)
	if err != nil {
		return nil, err
	}
	if len(allBytes) < 1 {
		return nil, errors.New("incorrect bytes bytesCount")
	}
	return allBytes, nil
}

func (client *Client) CreateUser() (string, error) {
	msg := []byte{commands.CreateUser}

	token := make([]byte, 8)
	binary.LittleEndian.PutUint64(token, rand.Uint64())
	tokenKey := make([]byte, 8)
	binary.LittleEndian.PutUint64(tokenKey, rand.Uint64())

	createUserOp := commands.CreateUserOp{
		Token:    token,
		TokenKey: hex.EncodeToString(tokenKey),
	}
	msg = append(msg, createUserOp.Serialize()...)
	response, err := client.sendMessage(msg)
	if err != nil {
		return "", err
	}
	if response.statusCode != '\x00' {
		return "", errors.New(fmt.Sprintf("cant create user: %02x", response.statusCode))
	}

	return hex.EncodeToString(tokenKey), nil
}

func (client *Client) GetUser(tokenKey string) (*models.User, error) {
	msg := []byte{commands.GetUser}
	getUserOp := commands.GetUserOp{
		TokenKey: tokenKey,
	}
	msg = append(msg, getUserOp.Serialize()...)
	response, err := client.sendMessage(msg)
	if err != nil {
		return nil, err
	}
	if response.statusCode != '\x00' {
		return nil, errors.New(fmt.Sprintf("cant get user: %02x", response.statusCode))
	}
	user, err := models.DeserializeUser(response.decryptedPayload)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (client *Client) CreateContainer(size int, description string) (string, error) {
	msg := []byte{commands.ContainerCreate}
	createContainerOp := commands.CreateContainerOp{
		Size:        uint8(size),
		Description: description,
	}
	msg = append(msg, createContainerOp.Serialize()...)
	response, err := client.sendMessage(msg)
	if err != nil {
		return "", err
	}
	if response.statusCode != '\x00' {
		return "", errors.New(fmt.Sprintf("cant create container: %02x", response.statusCode))
	}
	return string(response.decryptedPayload), nil
}

func (client *Client) GetContainerInfo(containerID string) (models.Container, error) {
	msg := []byte{commands.GetContainerInfo}
	getContainerInfoOp := commands.GetContainerInfoOp{
		ContainerID: containerID,
	}
	msg = append(msg, getContainerInfoOp.Serialize()...)
	response, err := client.sendMessage(msg)
	if err != nil {
		return models.Container{}, err
	}
	if response.statusCode != '\x00' {
		return models.Container{}, errors.New(fmt.Sprintf("cant get container: %02x", response.statusCode))
	}

	container, err := models.DeserializeContainer(response.decryptedPayload)
	if err != nil {
		return models.Container{}, err
	}

	return container, nil
}

func (client *Client) PutItem(item models.Item, containerId string) error {
	msg := []byte{commands.PutItem}
	putItemOp := commands.PutItemOp{
		Item:        item,
		ContainerId: containerId,
	}
	msg = append(msg, putItemOp.Serialize()...)
	response, err := client.sendMessage(msg)
	if err != nil {
		return err
	}
	if response.statusCode != '\x00' {
		return errors.New(fmt.Sprintf("cant put item: %02x", response.statusCode))
	}

	return nil
}

func (client *Client) GetItem(containerID string, index int) (models.Item, error) {
	msg := []byte{commands.GetItem}
	getItemOp := commands.GetItemOp{
		ContainerID: containerID,
		ItemIndex:   index,
	}
	msg = append(msg, getItemOp.Serialize()...)
	response, err := client.sendMessage(msg)
	if err != nil {
		return models.Item{}, err
	}
	if response.statusCode != '\x00' {
		return models.Item{}, errors.New(fmt.Sprintf("cant get item: %02x", response.statusCode))
	}

	container, err := models.DeserializeItem(response.decryptedPayload)
	if err != nil {
		return models.Item{}, err
	}

	return container, nil
}

func (client *Client) GetStat(skip, take int) (*models.Statistic, error) {
	msg := []byte{commands.GetStatistic}
	getStatisticOp := commands.GetStatisticOp{
		Skip: skip,
		Take: take,
	}
	msg = append(msg, getStatisticOp.Serialize()...)
	response, err := client.sendMessage(msg)
	if err != nil {
		return nil, err
	}
	if response.statusCode != '\x00' {
		return nil, errors.New(fmt.Sprintf("cant get statistic: %02x", response.statusCode))
	}
	statistic, err := models.DeserializeStatistic(response.decryptedPayload)
	if err != nil {
		return nil, err
	}
	return &statistic, nil
}
