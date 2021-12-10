package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"math"
	"math/rand"
	"sort"
	"time"
	"trash-factory/pkg/commands"
	"trash-factory/pkg/crypto"
	"trash-factory/pkg/models"
	"trash-factory/pkg/serializeb"
)

type ControlPanel struct {
	Commands         map[byte]interface{}
	DB               *DataBase
	Cryptor          *crypto.Cryptor
	stats            *models.Statistic
	AdminCredentials *models.User
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
		commands.GetStatistic:     cp.GetStatistic,
	}
	cp.DB = NewDataBase()
	cp.Cryptor = crypto.NewCryptor(magic)
	statistic, err := cp.CalculateStatistic()
	if err != nil {
		panic(err)
	}
	cp.stats = statistic
	rand.Seed(time.Now().Unix())
	CreateAdminUser(err, &cp)

	return &cp
}

func CreateAdminUser(err error, cp *ControlPanel) {
	adminTokenKey := fmt.Sprintf("%08x", rand.Uint64())
	adminToken := fmt.Sprintf("%08x", rand.Uint64())
	cp.AdminCredentials = &models.User{
		TokenKey:      adminTokenKey,
		Token:         []byte(adminToken),
		ContainersIds: []string{},
	}
	_, err = cp.CreateUser(adminTokenKey, commands.CreateUserOp{TokenKey: adminTokenKey, Token: []byte(adminToken)}.Serialize())
	if err != nil {
		panic(fmt.Sprintf("%s. %S", "Can't crteate admin user", err))
	}
}

func (cp *ControlPanel) ProcessMessage(msg []byte) (byte, []byte) {
	if len(msg) < 9 {
		log.Warnf("Incorrect length of command: %x", msg)
		return commands.StatusIncorrectSignature, nil
	}

	tokenKey := hex.EncodeToString(msg[:8])

	user, err := cp.DB.GetUser(tokenKey)
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
	users, err := cp.DB.GetAllUsers()
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
	users, err := cp.DB.GetAllUsers()
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
	op, err := commands.DeserializeGetUserOp(opBytes)
	if err != nil {
		return nil, err
	}

	if tokenKey != cp.AdminCredentials.TokenKey && tokenKey != op.TokenKey {
		return nil, errors.New("Forbidden")
	}

	user, err := cp.DB.GetUser(op.TokenKey)
	if err != nil {
		return nil, err
	}

	return user.Serialize(), nil
}

func (cp *ControlPanel) CreateUser(tokenKey string, opBytes []byte) ([]byte, error) {
	op, err := commands.DeserializeCreateUserOp(opBytes)

	if tokenKey != cp.AdminCredentials.TokenKey {
		return nil, errors.New("Forbidden")
	}

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

	container, err := cp.DB.GetContainer(tokenKey, op.ContainerID)
	if err != nil {
		return nil, err
	}

	return container.Items[op.ItemIndex].Serialize(), nil
}

func (cp *ControlPanel) PutItem(tokenKey string, opBytes []byte) ([]byte, error) {
	op, err := commands.DeserializePutItemOpOp(opBytes)

	cp.stats.AddItem(tokenKey, op.Item)

	container, err := cp.DB.GetContainer(tokenKey, op.ContainerId)
	if err != nil {
		return nil, err
	}

	if container.Size < uint8(len(container.Items)) {
		return nil, errors.New("Container is full")
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

	container, err := cp.DB.GetContainer(tokenKey, op.ContainerID)
	if err != nil {
		return nil, err
	}

	return container.Serialize()
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
	return []byte(container.ID), cp.DB.SaveContainer(tokenKey, &container)
}

func (cp *ControlPanel) GetStatistic(tokenKey string, opBytes []byte) ([]byte, error) {
	op, err := commands.DeserializeGetStatisticOp(opBytes)
	if err != nil {
		return nil, err
	}

	userStats := make([]*models.UserStatistic, 0, len(cp.stats.Users))

	for _, userStatistic := range cp.stats.Users {
		userStats = append(userStats, userStatistic)
	}

	sort.Slice(userStats, func(i, j int) bool {
		return userStats[i].Total > userStats[j].Total
	})

	if op.Skip >= len(userStats) {
		statistic := models.NewStatistic()
		return statistic.Serialize(), nil
	}

	maxTake := int(math.Min(float64(len(userStats)-op.Skip), float64(op.Take)))

	statistic := models.Statistic{
		Users: userStats[op.Skip : op.Skip+maxTake],
	}
	return statistic.Serialize(), nil
}

func (cp ControlPanel) CalculateStatistic() (*models.Statistic, error) {
	stats := models.NewStatistic()
	users, err := cp.DB.GetAllUsers()
	if err != nil {
		return nil, err
	}

	for _, userToken := range *users {
		user, err := cp.DB.GetUser(userToken)
		if err != nil {
			log.Error(err)
			continue
		}

		for _, containerId := range user.ContainersIds {
			container, err := cp.DB.GetContainer(userToken, containerId)
			if err != nil {
				log.Error(err)
				continue
			}

			for _, item := range container.Items {
				stats.AddItem(userToken, item)
			}
		}
	}
	return stats, nil
}

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
