package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"trash-factory/pkg/commands"
	"trash-factory/pkg/crypto"
	"trash-factory/pkg/models"
	"trash-factory/pkg/serializeb"
)

type ControlPanel struct {
	Commands map[byte]interface{}
	DB       *DataBase
	Cryptor  *crypto.Cryptor
}

func NewControlPanel() *ControlPanel {
	cp := ControlPanel{}
	cp.Commands = map[byte]interface{}{
		commands.ContainerCreate:  cp.CreateContainer,
		commands.ContainerList:    cp.ListContainers,
		commands.GetContainerInfo: cp.GetContainerInfo,
		commands.PutItem:          cp.PutItem,
		commands.GetItem:          cp.GetItem,
		commands.CreateUser:       cp.CreateUser,
		commands.GetUser:          cp.GetUser,
		commands.ListUsers:        cp.ListUsers,
	}
	cp.DB = NewDataBase()
	cp.Cryptor = crypto.NewCryptor(magic)
	return &cp
}

func (cp *ControlPanel) ProcessMessage(msg []byte) (byte, []byte) {
	if len(msg) < 9 {
		log.Warnf("Incorrect length of command: %x", msg)
		return commands.StatusIncorrectSignature, nil
	}

	user, err := cp.DB.GetUser(hex.EncodeToString(msg[:8]))
	if err != nil {
		log.Warn(err)
		return commands.StatusIncorrectSignature, nil
	}

	plainText, err := cp.Cryptor.DecryptMsg(user, msg[8:])
	if err != nil {
		log.Warn(err)
		return commands.StatusIncorrectSignature, nil
	}

	log.Debugf("PLAIN TEXT: %02x\n", plainText)

	statusCode, response := cp.RunCommand(plainText)
	cipherText, err := cp.Cryptor.EncryptMsg(user, response)
	if err != nil {
		return commands.StatusInternalError, nil
	}
	return statusCode, cipherText
}

func (cp *ControlPanel) RunCommand(msg []byte) (byte, []byte) {
	tokenKey := fmt.Sprintf("%08x", msg[:8])
	command := msg[8]
	args := make([]byte, 0)
	if len(msg) > 9 {
		args = msg[9:]
	}

	if value, ok := cp.Commands[command]; ok {
		response, err := value.(func(string, []byte) ([]byte, error))(tokenKey, args)
		if err != nil {
			log.Errorf("func \\x%02x exec error: %s", command, err)
			return commands.StatusCommandExecError, nil
		}
		return commands.StatusOK, response
	}
	log.Errorf("command \\x%x not found", command)
	return commands.StatusCommandNotFound, nil
}

func (cp *ControlPanel) ListContainers(tokenKey string, opBytes []byte) ([]byte, error) {
	users, err := cp.DB.GetAllUsers() //TODO: can't remember, but seems like here should be all containers id
	if err != nil {
		return nil, err
	}

	containersIds := make([]string, 0)
	for _, userTokenKey := range *users {
		user, err := cp.DB.GetUser(userTokenKey)
		if err != nil {
			return nil, err
		}
		containersIds = append(containersIds, user.ContainersIds...)

	}

	writer := serializeb.NewWriter()
	writer.WriteArray(serializeb.ToGenericArray(containersIds), func(item interface{}, writer *serializeb.Writer) {
		writer.WriteString(item.(string))
	})

	return writer.GetBytes()
}

func (cp *ControlPanel) ListUsers(tokenKey string, opBytes []byte) ([]byte, error) {
	users, err := cp.DB.GetAllUsers() //TODO: tokens only?
	if err != nil {
		return nil, err
	}

	writer := serializeb.NewWriter()

	//TODO:rewrite array serialization
	writer.WriteArray(serializeb.ToGenericArray(users), func(item interface{}, writer *serializeb.Writer) {
		writer.WriteString(item.(string))
	})

	return writer.GetBytes()
}

func (cp *ControlPanel) GetUser(tokenKey string, opBytes []byte) ([]byte, error) {
	user, err := cp.DB.GetUser(tokenKey)
	if err != nil {
		return nil, err
	}

	return user.Serialize()
}

func (cp *ControlPanel) CreateUser(tokenKey string, opBytes []byte) ([]byte, error) {
	op, err := commands.DeserializeCreateUserOp(opBytes)

	user := models.User{
		TokenKey:      tokenKey,
		Token:         op.Token,
		ContainersIds: make([]string, 0),
	}

	err = cp.DB.SaveUser(&user)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (cp *ControlPanel) GetItem(tokenKey string, opBytes []byte) ([]byte, error) {
	op, err := commands.DeserializeGetItemOp(opBytes)

	err = cp.EnsureUserAuthorize(tokenKey, op.ContainerID)
	if err != nil {
		return nil, err
	}

	container, err := cp.DB.GetContainer(op.ContainerID)
	if err != nil {
		return nil, err
	}

	return container.Items[op.ItemIndex].Serialize()
}

func (cp *ControlPanel) PutItem(tokenKey string, opBytes []byte) ([]byte, error) {
	op, err := commands.DeserializePutItemOpOp(opBytes)

	containerId := "...." //TODO: implement container peeking logic. add id to request, or type field to container model
	//TODO: validate size
	container, err := cp.DB.GetContainer(containerId)
	if err != nil {
		return nil, err
	}
	container.Items = append(container.Items, op.Item)
	err = cp.DB.SaveContainer(tokenKey, container)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (cp *ControlPanel) GetContainerInfo(tokenKey string, opBytes []byte) ([]byte, error) {
	op, err := commands.DeserializeGetContainerInfoOp(opBytes)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Can't deserialize container. %s", err))
	}

	err = cp.EnsureUserAuthorize(tokenKey, op.ContainerID)
	if err != nil {
		return nil, err
	}

	container, err := cp.DB.GetContainer(op.ContainerID)
	if err != nil {
		return nil, err
	}

	return container.Serialize() //TODO: replace with ContainerInfo
}

func (cp *ControlPanel) EnsureUserAuthorize(tokenKey string, id string) error {
	user, err := cp.DB.GetUser(tokenKey)
	if err != nil {
		return errors.New(fmt.Sprintf("Can't get user. %s", err))
	}

	if !Contains(user.ContainersIds, id) {
		return errors.New(fmt.Sprintf("User %s not owner of container %s. Not your trash.", user.TokenKey, id))
	}
	return nil
}

func (cp *ControlPanel) CreateContainer(tokenKey string, opBytes []byte) ([]byte, error) {
	op, err := commands.DeserializeCreateContainerOp(opBytes)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Can't deserialize container. %s", err))
	}

	fmt.Printf("Container creating... TOKEN %02x  ARGS %02x\n", tokenKey, opBytes)

	if op.Size > '\x05' || op.Size < '\x01' {
		return nil, errors.New("incorrect container size")
	}

	fmt.Println(op.Description)
	container := models.Container{
		ID:          fmt.Sprintf("%08x", rand.Uint64()),
		Size:        op.Size,
		Description: op.Description,
	}
	return nil, cp.DB.SaveContainer(tokenKey, &container)
}

//TODO: move somewhere else
func Contains(ids []string, target string) bool {
	contains := false
	for _, id := range ids {
		if id == target {
			contains = true
			break
		}
	}
	return contains
}
