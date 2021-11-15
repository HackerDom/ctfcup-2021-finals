package main

import (
	"bytes"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

type DataBase struct {
	dbPath string
	userDBPath string
	containerDBPath string
}

type User struct {
	tokenKey string
	token []byte
	containersIds []string
}

type Container struct {
	ID string
	Size uint8
	Items []Item
	Description string
}

type Item struct {
	Type uint8
	Weight uint8
	Description string
}

func NewDataBase() *DataBase {
	db := DataBase{ dbPath: "db/" }
	db.userDBPath = db.dbPath + "users/"
	db.containerDBPath = db.dbPath + "containers/"

	err := os.MkdirAll(db.userDBPath, 0777)
	if err != nil {
		log.Fatal("Can't create users folder: " + err.Error())
	}
	err = os.MkdirAll(db.containerDBPath, 0777)
	if err != nil {
		log.Fatal("Can't create containers folder: " + err.Error())
	}

	return &db
}

func (db *DataBase) GetUser(tokenKey string) (*User, error) {
	userInfo, err := os.ReadFile(db.userDBPath + tokenKey)
	if err != nil {
		return nil, errors.New("user not found")
	}

	splitedUserInfo := bytes.Split(userInfo, []byte{'\n'})
	user := new(User)
	user.tokenKey = tokenKey
	user.token = splitedUserInfo[0]
	for i := 1; i < len(splitedUserInfo); i++ {
		user.containersIds = append(user.containersIds, fmt.Sprintf("%08x", splitedUserInfo[i]))
	}
	fmt.Println(user.containersIds)
	return user, nil
}

func (db *DataBase) SaveUser(user *User) error {
	filename := user.tokenKey
	data := user.token
	for _, containerId := range user.containersIds {
		data = append(data, '\n')
		data = append(data, []byte(containerId)...)
	}
	err := os.WriteFile(db.userDBPath + filename,  data, 0666)
	return err
}

func (item *Item) serialize() []byte {
	return append(
		[]byte{ item.Type, item.Weight },
		[]byte(strings.Replace(item.Description, "\n", "", -1))...
		)
}

func (db *DataBase) SaveContainer(tokenKey string, container *Container) error {
	data := append([]byte{ container.Size, '\n' }, []byte(container.Description)...)
	for n, item := range container.Items {
		if uint8(n + 1) > container.Size {
			break
		}
		data = append(data, '\n')
		data = append(data, item.serialize()...)
	}

	if err := os.WriteFile(db.containerDBPath + container.ID, data, 0666); err != nil {
		return err
	}

	user, err := db.GetUser(tokenKey)
	if err != nil {
		return err
	}
	user.containersIds = append(user.containersIds, container.ID)
	return db.SaveUser(user)
}