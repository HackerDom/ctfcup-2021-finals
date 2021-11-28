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

	payload, err := cp.Cryptor.DecryptMsg(user.Token, msg[8:])
	if err != nil {
		log.Warn(err)
		return commands.StatusIncorrectSignature, nil
	}

	statusCode, response := cp.RunCommand(user, payload)
	cipherText, err := cp.Cryptor.EncryptMsg(user.TokenKey, user.Token, response)
	if err != nil {
		return commands.StatusInternalError, nil
	}
	return statusCode, cipherText
}

func (cp *ControlPanel) RunCommand(user *models.User, msg []byte) (byte, []byte) {
	command := msg[0]
	args := msg[1:]
	if value, ok := cp.Commands[command]; ok {
		response, err := value.(func(string, []byte) ([]byte, error))(user.TokenKey, args)
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
	writer.WriteArraySize(len(containersIds))
	for _, id := range containersIds {
		writer.WriteString(id)
	}
	return writer.GetBytes(), nil
}

func (cp *ControlPanel) ListUsers(tokenKey string, opBytes []byte) ([]byte, error) {
	users, err := cp.DB.GetAllUsers() //TODO: tokens only?
	if err != nil {
		return nil, err
	}

	writer := serializeb.NewWriter()

	writer.WriteArraySize(len(*users))
	for _, user := range *users {
		writer.WriteString(user)
	}

	return writer.GetBytes(), nil
}

func (cp *ControlPanel) GetUser(tokenKey string, opBytes []byte) ([]byte, error) {
	user, err := cp.DB.GetUser(tokenKey)
	if err != nil {
		return nil, err
	}

	return user.Serialize(), nil
}

func (cp *ControlPanel) CreateUser(tokenKey string, opBytes []byte) ([]byte, error) {
	op, err := commands.DeserializeCreateUserOp(opBytes)

	//TODO: validate admin token

	user := models.User{
		TokenKey:      op.TokenKey,
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

	return container.Items[op.ItemIndex].Serialize(), nil
}

func (cp *ControlPanel) PutItem(tokenKey string, opBytes []byte) ([]byte, error) {
	op, err := commands.DeserializePutItemOpOp(opBytes)

	user, err := cp.DB.GetUser(tokenKey)
	if err != nil {
		return nil, err
	}

	for _, id := range user.ContainersIds {
		container, err := cp.DB.GetContainer(id)
		if err != nil {
			return nil, err
		}

		if container.Size > uint8(len(container.Items)) {
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
	}

	return nil, errors.New("Can't find acceptable container")
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
